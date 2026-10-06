package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

// These tests exercise redaction through the real makeRequest path (not redactSensitiveFields in
// isolation) so they fail if the redaction call site is ever removed or the debug-logging guard
// regresses — the helper-only tests in redact_test.go would keep passing in that case. They assert
// three things end-to-end: (1) certificate/password values never reach the debug log, (2) the bytes
// actually uploaded to the server are the unredacted originals and stay identical across retries,
// and (3) with debug logging disabled the body is not logged at all.

const (
	testSecretBlob     = "SUPER-SECRET-PKCS12-BLOB"
	testSecretPassword = "SUPER-SECRET-PASSWORD"
)

// secretCertificate returns a Certificate whose sensitive fields (certificate, certificatePassword)
// carry recognizable sentinels, plus non-sensitive fields used to prove the body was logged
// (redacted) rather than merely dropped.
func secretCertificate() Certificate {
	return Certificate{
		CertificateID:       "cert-id-1",
		CertificateName:     "my-cert-name",
		Certificate:         testSecretBlob,
		CertificatePassword: testSecretPassword,
		ApplicationID:       "app-1",
		Domain:              "example.com",
	}
}

// setDebugLogging sets all three TF_LOG* vars plus the two acc-test log-path vars explicitly so
// ambient values in the runner's environment cannot leak into debugLoggingEnabled(). Cannot be
// combined with t.Parallel.
func setDebugLogging(t *testing.T, level string) {
	t.Helper()
	t.Setenv("TF_LOG_PROVIDER_CITRIXSPA", level)
	t.Setenv("TF_LOG_PROVIDER", "")
	t.Setenv("TF_LOG", "")
	t.Setenv("TF_ACC_LOG_PATH", "")
	t.Setenv("TF_LOG_PATH_MASK", "")
}

// findLogEntry returns the first decoded log entry whose @message matches, or nil.
func findLogEntry(entries []map[string]any, message string) map[string]any {
	for _, e := range entries {
		if m, ok := e["@message"].(string); ok && m == message {
			return e
		}
	}
	return nil
}

// TestMakeRequest_DebugLogRedactsSecrets drives a real POST through makeRequest with debug logging
// enabled and asserts the captured log never contains the certificate blob or password, that the
// redaction marker is present, and that the unredacted body is still what actually went on the wire.
func TestMakeRequest_DebugLogRedactsSecrets(t *testing.T) {
	setDebugLogging(t, "DEBUG")

	var sentBody string
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, 0, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			req.Body.Close()
			sentBody = string(b)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	})

	var buf bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &buf)

	resp, err := client.makeRequest(ctx, http.MethodPost, "/certificates", secretCertificate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	logs := buf.String()
	if strings.Contains(logs, testSecretBlob) {
		t.Errorf("certificate blob leaked into debug logs:\n%s", logs)
	}
	if strings.Contains(logs, testSecretPassword) {
		t.Errorf("certificate password leaked into debug logs:\n%s", logs)
	}

	entries, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("failed to decode captured logs: %v", err)
	}
	reqLog := findLogEntry(entries, "spa-terraform-provider: SPA API request")
	if reqLog == nil {
		t.Fatalf("did not find request debug log entry in:\n%s", logs)
	}
	body, ok := reqLog["body"].(string)
	if !ok {
		t.Fatalf("request log entry has no string body field: %#v", reqLog)
	}
	if !strings.Contains(body, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker in logged body, got %q", body)
	}
	// The non-sensitive name must survive, proving the body was logged (and redacted) rather than
	// dropped wholesale — otherwise "no secret in logs" would pass trivially.
	if !strings.Contains(body, "my-cert-name") {
		t.Errorf("non-sensitive certificateName missing from logged body, got %q", body)
	}

	// The redaction must only affect the logged copy: the real secrets must have gone on the wire.
	if !strings.Contains(sentBody, testSecretBlob) || !strings.Contains(sentBody, testSecretPassword) {
		t.Errorf("uploaded body was redacted instead of the log copy, got %q", sentBody)
	}
}

// TestMakeRequest_UploadBytesUnchangedAcrossRetries proves redaction never mutates the request
// payload: the bytes uploaded on the initial attempt and the 429 retry are byte-for-byte identical
// and equal to the original marshaled body (secrets intact).
func TestMakeRequest_UploadBytesUnchangedAcrossRetries(t *testing.T) {
	setDebugLogging(t, "DEBUG")

	var mu sync.Mutex
	var attempts []string
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, 0, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var body string
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			req.Body.Close()
			body = string(b)
		}
		mu.Lock()
		attempts = append(attempts, body)
		n := len(attempts)
		mu.Unlock()

		if n == 1 {
			// Retry-After: 0 keeps the retry immediate so the test does not sleep.
			header := make(http.Header)
			header.Set("Retry-After", "0")
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(strings.NewReader(`{"statusCode":429}`)),
				Header:     header,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	})

	cert := secretCertificate()
	want, err := json.Marshal(cert)
	if err != nil {
		t.Fatalf("failed to marshal certificate: %v", err)
	}

	ctx := tflogtest.RootLogger(context.Background(), &bytes.Buffer{})
	resp, err := client.makeRequest(ctx, http.MethodPost, "/certificates", cert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if len(attempts) != 2 {
		t.Fatalf("expected 2 upload attempts (initial + retry), got %d", len(attempts))
	}
	if attempts[0] != attempts[1] {
		t.Errorf("retry uploaded different bytes than initial attempt:\n initial: %q\n retry:   %q", attempts[0], attempts[1])
	}
	if attempts[0] != string(want) {
		t.Errorf("uploaded bytes were altered from the original body:\n got:  %q\n want: %q", attempts[0], string(want))
	}
	// Sanity: the real secrets must be present on the wire (redaction touched only the log copy).
	if !strings.Contains(attempts[1], testSecretBlob) || !strings.Contains(attempts[1], testSecretPassword) {
		t.Errorf("retry upload lost the original secret values: %q", attempts[1])
	}
}

// TestMakeRequest_DebugDisabledOmitsBody covers the production default: with all TF_LOG* unset the
// request body is never added to the log fields, so no secret can leak even though the capturing
// logger records at TRACE. This fails if the debugLoggingEnabled() guard is removed.
func TestMakeRequest_DebugDisabledOmitsBody(t *testing.T) {
	setDebugLogging(t, "")

	var sentBody string
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, 0, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			req.Body.Close()
			sentBody = string(b)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	})

	var buf bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &buf)

	resp, err := client.makeRequest(ctx, http.MethodPost, "/certificates", secretCertificate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	logs := buf.String()
	if strings.Contains(logs, testSecretBlob) || strings.Contains(logs, testSecretPassword) {
		t.Errorf("secret leaked into logs while debug logging was disabled:\n%s", logs)
	}

	entries, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("failed to decode captured logs: %v", err)
	}
	reqLog := findLogEntry(entries, "spa-terraform-provider: SPA API request")
	if reqLog == nil {
		t.Fatalf("did not find request debug log entry in:\n%s", logs)
	}
	if _, hasBody := reqLog["body"]; hasBody {
		t.Errorf("request body was logged despite debug logging being disabled: %#v", reqLog)
	}

	// The request itself must still be sent in full with the real secrets on the wire.
	if !strings.Contains(sentBody, testSecretBlob) || !strings.Contains(sentBody, testSecretPassword) {
		t.Errorf("request body was not uploaded intact with debug disabled: %q", sentBody)
	}
}
