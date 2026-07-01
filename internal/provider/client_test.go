package provider

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// roundTripFunc allows using a plain function as an http.RoundTripper,
// so we can intercept outgoing requests without starting an HTTP server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestClient creates an APIClient with a custom transport that captures
// the outgoing request instead of making a real HTTP call.
func newTestClient(customerID, authToken string, suppressASBNotifications bool, captured *http.Request) *APIClient {
	client := NewAPIClient("https://test.example.com", customerID, authToken, nil, false, suppressASBNotifications, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		*captured = *req
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	})
	return client
}

func TestMakeRequest_SuppressASBNotificationsEnabled(t *testing.T) {
	var captured http.Request
	client := newTestClient("test-customer", "test-token", true, &captured)

	resp, err := client.makeRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	got := captured.Header.Get("X-Send-ASB-Notification")
	if got != "false" {
		t.Errorf("expected X-Send-ASB-Notification header to be \"false\", got %q", got)
	}
}

func TestMakeRequest_SuppressASBNotificationsDisabled(t *testing.T) {
	var captured http.Request
	client := newTestClient("test-customer", "test-token", false, &captured)

	resp, err := client.makeRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if got := captured.Header.Get("X-Send-ASB-Notification"); got != "" {
		t.Errorf("expected no X-Send-ASB-Notification header, got %q", got)
	}
}

func TestMakeRequest_StandardHeaders(t *testing.T) {
	var captured http.Request
	client := newTestClient("cust-123", "tok-abc", false, &captured)

	resp, err := client.makeRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	checks := map[string]string{
		"Citrix-CustomerId": "cust-123",
		"Authorization":     "CWSAuth bearer=tok-abc",
		"Content-Type":      "application/json; charset=utf-8",
		"Accept":            "application/json",
		"User-Agent":        "test-agent",
	}

	for header, want := range checks {
		got := captured.Header.Get(header)
		if got != want {
			t.Errorf("header %s: expected %q, got %q", header, want, got)
		}
	}

	if captured.Header.Get("Citrix-TransactionId") == "" {
		t.Error("expected Citrix-TransactionId header to be set")
	}
}

func TestMakeRequest_429RetryWithRetryAfterHeader(t *testing.T) {
	var callCount int32
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempt := atomic.AddInt32(&callCount, 1)
		if attempt == 1 {
			// First call returns 429 with Retry-After: 0
			header := make(http.Header)
			header.Set("Retry-After", "0")
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"statusCode":429,"message":"Rate limit is exceeded. Try again in 0 seconds."}`))),
				Header:     header,
			}, nil
		}
		// Second call succeeds
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123"}`))),
			Header:     make(http.Header),
		}, nil
	})

	resp, err := client.makeRequest(context.Background(), http.MethodGet, "/test", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", callCount)
	}
}

func TestMakeRequest_429RetriesExhausted(t *testing.T) {
	var callCount int32
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&callCount, 1)
		// Always return 429
		header := make(http.Header)
		header.Set("Retry-After", "0")
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"statusCode":429,"message":"Rate limit is exceeded. Try again in 0 seconds."}`))),
			Header:     header,
		}, nil
	})

	_, err := client.makeRequest(context.Background(), http.MethodGet, "/test", nil)

	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	if !strings.Contains(err.Error(), "API rate limit exceeded after 3 retries") {
		t.Errorf("unexpected error message: %v", err)
	}
	// Initial attempt + 3 retries = 4 total calls
	if atomic.LoadInt32(&callCount) != 4 {
		t.Errorf("expected 4 HTTP calls (1 initial + 3 retries), got %d", callCount)
	}
}

func TestMakeRequest_429RetryWithBody(t *testing.T) {
	var callCount int32
	var lastBody string
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempt := atomic.AddInt32(&callCount, 1)
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			lastBody = string(b)
		}
		if attempt == 1 {
			header := make(http.Header)
			header.Set("Retry-After", "5")
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"statusCode":429,"message":"Rate limit is exceeded. Try again in 5 seconds."}`))),
				Header:     header,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	})

	body := map[string]string{"name": "test-app"}
	start := time.Now()
	resp, err := client.makeRequest(context.Background(), http.MethodPost, "/applications", body)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the retry delay was respected (should be at least 5 seconds)
	expectedDelay := 5 * time.Second
	if elapsed < expectedDelay {
		t.Errorf("expected request to take at least %v due to Retry-After header, but it took %v", expectedDelay, elapsed)
	}
	// Verify it didn't take too long (with 2 second tolerance for test overhead)
	maxDelay := expectedDelay + 2*time.Second
	if elapsed > maxDelay {
		t.Errorf("expected request to take less than %v, but it took %v", maxDelay, elapsed)
	}

	// Verify body was resent on retry
	if !strings.Contains(lastBody, "test-app") {
		t.Errorf("expected request body to contain 'test-app' on retry, got %q", lastBody)
	}
}

func TestMakeRequest_429RetryContextCancelled(t *testing.T) {
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set("Retry-After", "60") // Long wait
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"statusCode":429,"message":"Rate limit is exceeded. Try again in 60 seconds."}`))),
			Header:     header,
		}, nil
	})

	// Create a context that cancels immediately after the timer would start waiting
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.makeRequest(ctx, http.MethodGet, "/test", nil)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("expected context-related error, got: %v", err)
	}
}

func TestMakeRequest_307RegionalRedirect(t *testing.T) {
	const redirectBody = `{
		"type": "https://errors-api.cloud.com/common/Redirect",
		"detail": "Regional customer must use their designated regional endpoint",
		"parameters": [
			{"name": "Citrix-CustomerId", "value": "d8cd538d1e48"},
			{"name": "customerDataRegion", "value": "eu"},
			{"name": "currentAPIMRegion", "value": "us"}
		]
	}`

	client := NewAPIClient("https://api.cloud.com/accessSecurity", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set("Location", "https://api-eu.cloud.com/feature-data_accesssecurity/sessionPolicy")
		header.Set("Content-Type", "application/json")
		return &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Body:       io.NopCloser(bytes.NewReader([]byte(redirectBody))),
			Header:     header,
		}, nil
	})

	_, err := client.makeRequest(context.Background(), http.MethodGet, "/sessionPolicy", nil)

	if err == nil {
		t.Fatal("expected error for 307 regional redirect, got nil")
	}

	errMsg := err.Error()
	checks := []string{
		"307 Temporary Redirect",
		`"eu"`,
		`"us"`,
		"https://api.cloud.com/accessSecurity",
		"https://api-eu.cloud.com/accessSecurity",
		"Regional customer must use their designated regional endpoint",
	}
	for _, want := range checks {
		if !strings.Contains(errMsg, want) {
			t.Errorf("expected error to contain %q, got:\n%s", want, errMsg)
		}
	}
}

func TestMakeRequest_307UntrustedLocationHost(t *testing.T) {
	const redirectBody = `{
		"type": "https://errors-api.cloud.com/common/Redirect",
		"detail": "Regional customer must use their designated regional endpoint",
		"parameters": [
			{"name": "customerDataRegion", "value": "eu"},
			{"name": "currentAPIMRegion", "value": "us"}
		]
	}`

	client := NewAPIClient("https://api.cloud.com/accessSecurity", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set("Location", "https://evil.example.com/accessSecurity")
		return &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Body:       io.NopCloser(bytes.NewReader([]byte(redirectBody))),
			Header:     header,
		}, nil
	})

	_, err := client.makeRequest(context.Background(), http.MethodGet, "/sessionPolicy", nil)

	if err == nil {
		t.Fatal("expected error for 307 redirect, got nil")
	}

	errMsg := err.Error()
	// Untrusted host must not appear in the error — correctBaseURL falls back to c.BaseURL
	if strings.Contains(errMsg, "evil.example.com") {
		t.Errorf("untrusted Location host must not appear in error message, got:\n%s", errMsg)
	}
	// The configured base_url should be used as the fallback
	if !strings.Contains(errMsg, "https://api.cloud.com/accessSecurity") {
		t.Errorf("expected fallback to current base URL, got:\n%s", errMsg)
	}
}

func TestMakeRequest_307MissingLocationHeader(t *testing.T) {
	const redirectBody = `{
		"type": "https://errors-api.cloud.com/common/Redirect",
		"detail": "Regional customer must use their designated regional endpoint",
		"parameters": [
			{"name": "customerDataRegion", "value": "asp"},
			{"name": "currentAPIMRegion", "value": "us"}
		]
	}`

	client := NewAPIClient("https://api.cloud.com/accessSecurity", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTemporaryRedirect,
			Body:       io.NopCloser(bytes.NewReader([]byte(redirectBody))),
			Header:     make(http.Header), // no Location header
		}, nil
	})

	_, err := client.makeRequest(context.Background(), http.MethodGet, "/sessionPolicy", nil)

	if err == nil {
		t.Fatal("expected error for 307 redirect with missing Location header, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "307 Temporary Redirect") {
		t.Errorf("expected error to contain '307 Temporary Redirect', got: %v", err)
	}
	// When Location is missing, correctBaseURL falls back to the configured BaseURL
	if !strings.Contains(errMsg, "https://api.cloud.com/accessSecurity") {
		t.Errorf("expected error to contain current base URL fallback, got: %v", err)
	}
	if !strings.Contains(errMsg, `"asp"`) {
		t.Errorf("expected error to contain customer data region 'asp', got: %v", err)
	}
}

// newTestClientWithStatus creates an APIClient whose transport always returns
// the given status code with an empty JSON body.
func newTestClientWithStatus(statusCode int) *APIClient {
	client := NewAPIClient("https://test.example.com", "cust-123", "tok-abc", nil, false, false, nil, "test-agent")
	client.HTTPClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	return client
}

func TestDeleteApplication_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteApplication(context.Background(), "app-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteApplication_200Success(t *testing.T) {
	client := newTestClientWithStatus(http.StatusOK)
	err := client.DeleteApplication(context.Background(), "app-123")
	if err != nil {
		t.Fatalf("expected nil error for 200 on delete, got: %v", err)
	}
}

func TestDeleteApplication_500ReturnsError(t *testing.T) {
	client := newTestClientWithStatus(http.StatusInternalServerError)
	err := client.DeleteApplication(context.Background(), "app-123")
	if err == nil {
		t.Fatal("expected error for 500 on delete, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain '500', got: %v", err)
	}
}

func TestDeleteAccessPolicy_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteAccessPolicy(context.Background(), "policy-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteSecurityGroup_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteSecurityGroup(context.Background(), "sg-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteRoutingDomain_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteRoutingDomain(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteCertificate_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteCertificate(context.Background(), "cert-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteTerminateMachineAccess_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteTerminateMachineAccess(context.Background(), "machine-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteTerminateUserAccess_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteTerminateUserAccess(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}

func TestDeleteSessionPolicy_404TreatedAsSuccess(t *testing.T) {
	client := newTestClientWithStatus(http.StatusNotFound)
	err := client.DeleteSessionPolicy(context.Background(), "policy-123")
	if err != nil {
		t.Fatalf("expected nil error for 404 on delete, got: %v", err)
	}
}
