package provider

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactSensitiveFields_RealCertificateStruct(t *testing.T) {
	cert := Certificate{
		CertificateID:       "test-cert-id",
		CertificateName:     "test-cert-name",
		Certificate:         "test-pkcs12-blob",
		CertificatePassword: "fake-password-123",
		ApplicationID:       "test-app-id",
		Domain:              "example.com",
	}
	marshaled, err := json.Marshal(cert)
	if err != nil {
		t.Fatalf("failed to marshal Certificate: %v", err)
	}
	redacted := redactSensitiveFields(string(marshaled))

	if strings.Contains(redacted, "fake-password-123") {
		t.Errorf("dummy password leaked in redacted output: %s", redacted)
	}
	if strings.Contains(redacted, "test-pkcs12-blob") {
		t.Errorf("dummy certificate blob leaked in redacted output: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker in output: %s", redacted)
	}
	if !strings.Contains(redacted, "test-cert-name") {
		t.Errorf("non-sensitive certificateName was redacted: %s", redacted)
	}
	if !strings.Contains(redacted, "test-cert-id") {
		t.Errorf("non-sensitive certificateId was redacted: %s", redacted)
	}
}

func TestRedactSensitiveFields_NegativeControl(t *testing.T) {
	cert := Certificate{
		CertificateName:     "control-test",
		Certificate:         "test-pkcs12-blob",
		CertificatePassword: "fake-password-123",
	}
	marshaled, err := json.Marshal(cert)
	if err != nil {
		t.Fatalf("failed to marshal Certificate: %v", err)
	}
	raw := string(marshaled)
	if !strings.Contains(raw, "fake-password-123") {
		t.Fatalf("raw JSON does not contain dummy password: %s", raw)
	}
	redacted := redactSensitiveFields(raw)
	if strings.Contains(redacted, "fake-password-123") {
		t.Errorf("redaction did not remove password: %s", redacted)
	}
}

func TestRedactSensitiveFields_PreservesNonSensitive(t *testing.T) {
	input := `{"certificateName":"my-cert","certificatePassword":"secret123","domain":"example.com"}`
	redacted := redactSensitiveFields(input)

	if strings.Contains(redacted, "secret123") {
		t.Errorf("certificatePassword leaked: %s", redacted)
	}
	if !strings.Contains(redacted, "my-cert") {
		t.Errorf("non-sensitive certificateName was redacted: %s", redacted)
	}
	if !strings.Contains(redacted, "example.com") {
		t.Errorf("non-sensitive domain was redacted: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker: %s", redacted)
	}
}

func TestRedactSensitiveFields_NonJSON(t *testing.T) {
	input := "plain text body"
	redacted := redactSensitiveFields(input)
	if redacted != unparseableBodyMarker {
		t.Errorf("non-JSON body not withheld: got %q, want marker", redacted)
	}
}

func TestRedactSensitiveFields_InvalidJSON(t *testing.T) {
	input := `{"invalid": unclosed`
	redacted := redactSensitiveFields(input)
	if redacted != unparseableBodyMarker {
		t.Errorf("malformed JSON not withheld: got %q, want marker", redacted)
	}
}

func TestRedactSensitiveFields_UnparseableBodyFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		secret string
	}{
		{"non-JSON text", `error: authentication failed for password=hunter2 please retry`, "hunter2"},
		{"non-JSON with spaced value", `password=foo bar baz`, "bar"},
		{"truncated JSON", `{"user":"alice","password":"hunter2`, "hunter2"},
		{"escaped quote before secret", `{"password":"abc\"hunter2`, "hunter2"},
		{"non-JSON secret with spaces", `token = my secret token value`, "secret token value"},
		// Trailing content beginning with '}' or ']' is the dec.More() blind spot: More() is a
		// container-iteration predicate that returns false on those bytes, so these must be caught
		// by the stricter second-Decode-to-io.EOF check instead. Non-sensitive keys are used so the
		// secret would actually leak (not merely be [REDACTED]) if the guard were bypassed.
		{"trailing bracket after object", `{"data":"hunter2"}]`, "hunter2"},
		{"trailing brace after object", `{"data":"hunter2"}}`, "hunter2"},
		{"trailing bracket after array", `["hunter2"]]`, "hunter2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := redactSensitiveFields(tt.input)
			if strings.Contains(redacted, tt.secret) {
				t.Errorf("secret leaked from unparseable body: %s", redacted)
			}
			if redacted != unparseableBodyMarker {
				t.Errorf("expected unparseable body marker, got: %s", redacted)
			}
		})
	}
}

func TestRedactSensitiveFields_ExactKeyMatch(t *testing.T) {
	// Matching is exact and case-sensitive against the wire spellings `certificate` and
	// `certificatePassword` — the only secret-bearing fields in the typed SPA API bodies. Variant
	// spellings (snake_case, kebab-case, different casing) are intentionally NOT redacted: they are
	// not schema keys, and the free-form maps that could carry arbitrary spellings are config-only.
	redactedTests := []struct {
		name   string
		input  string
		secret string
	}{
		{"exact certificate blob", `{"certificate":"topsecret"}`, "topsecret"},
		{"exact certificatePassword", `{"certificatePassword":"topsecret"}`, "topsecret"},
		{"exact certificate nested in map", `{"metadata":{"certificate":"topsecret"}}`, "topsecret"},
	}
	for _, tt := range redactedTests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := redactSensitiveFields(tt.input)
			if strings.Contains(redacted, tt.secret) {
				t.Errorf("secret leaked for exact key: %s", redacted)
			}
			if !strings.Contains(redacted, "[REDACTED]") {
				t.Errorf("expected [REDACTED] marker: %s", redacted)
			}
		})
	}

	preservedTests := []struct {
		name  string
		input string
		keep  string
	}{
		{"snake_case certificate_password not a schema key", `{"certificate_password":"keepme"}`, "keepme"},
		{"kebab-case certificate-password not a schema key", `{"certificate-password":"keepme"}`, "keepme"},
		{"different-case CertificatePassword not a schema key", `{"CertificatePassword":"keepme"}`, "keepme"},
	}
	for _, tt := range preservedTests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := redactSensitiveFields(tt.input)
			if !strings.Contains(redacted, tt.keep) {
				t.Errorf("non-schema key %q was unexpectedly redacted: %s", tt.name, redacted)
			}
			if strings.Contains(redacted, "[REDACTED]") {
				t.Errorf("unexpected redaction of non-schema key: %s", redacted)
			}
		})
	}
}

func TestRedactSensitiveFields_FreeFormMapCarriesCertificate(t *testing.T) {
	// The exact-match design assumes the schema's free-form map[string]any fields
	// (customProperties, metadata, ...) are config-only, but they are still walked so that a
	// certificate/certificatePassword nested inside one is redacted by key — while a benign
	// sibling value in the same map is preserved (not over-redacted).
	tests := []struct {
		name  string
		input string
		keep  string
	}{
		{"certificate in customProperties", `{"customProperties":{"certificate":"topsecret","label":"prod"}}`, "prod"},
		{"certificatePassword in rule metadata", `{"metadata":{"certificatePassword":"topsecret","owner":"team-x"}}`, "team-x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := redactSensitiveFields(tt.input)
			if strings.Contains(redacted, "topsecret") {
				t.Errorf("secret leaked from free-form map: %s", redacted)
			}
			if !strings.Contains(redacted, "[REDACTED]") {
				t.Errorf("expected [REDACTED] marker: %s", redacted)
			}
			if !strings.Contains(redacted, tt.keep) {
				t.Errorf("non-secret sibling value %q was over-redacted: %s", tt.keep, redacted)
			}
		})
	}
}

func TestRedactSensitiveFields_DoesNotOverRedact(t *testing.T) {
	// Non-secret keys that superficially resemble sensitive ones must be preserved for debugging.
	input := `{"certificateId":"cid-1","certificateName":"my-cert","tagKey":"role","keywords":"a,b"}`
	redacted := redactSensitiveFields(input)
	for _, want := range []string{"cid-1", "my-cert", "role", "a,b"} {
		if !strings.Contains(redacted, want) {
			t.Errorf("non-sensitive value %q was over-redacted: %s", want, redacted)
		}
	}
	if strings.Contains(redacted, "[REDACTED]") {
		t.Errorf("unexpected redaction of non-sensitive keys: %s", redacted)
	}
}

func TestRedactSensitiveFields_EmptyBody(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"   "},
		{"\n\t"},
	}
	for _, tt := range tests {
		redacted := redactSensitiveFields(tt.input)
		if redacted != tt.input {
			t.Errorf("empty/whitespace body modified: got %q, want %q", redacted, tt.input)
		}
	}
}

func TestRedactSensitiveFields_NestedJSON(t *testing.T) {
	input := `{"user":{"name":"alice","certificatePassword":"secret123"},"domain":"example.com"}`
	redacted := redactSensitiveFields(input)

	if strings.Contains(redacted, "secret123") {
		t.Errorf("nested certificatePassword leaked: %s", redacted)
	}
	if !strings.Contains(redacted, "alice") {
		t.Errorf("non-sensitive name was redacted: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker: %s", redacted)
	}
}

func TestRedactSensitiveFields_RootArray(t *testing.T) {
	input := `[{"name":"item1","certificatePassword":"test-pass-1"},{"name":"item2","certificate":"test-secret-2"}]`
	redacted := redactSensitiveFields(input)

	if strings.Contains(redacted, "test-pass-1") {
		t.Errorf("certificatePassword in array element leaked: %s", redacted)
	}
	if strings.Contains(redacted, "test-secret-2") {
		t.Errorf("certificate in array element leaked: %s", redacted)
	}
	if !strings.Contains(redacted, "item1") || !strings.Contains(redacted, "item2") {
		t.Errorf("non-sensitive name fields were lost: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker: %s", redacted)
	}
}

func TestRedactSensitiveFields_PreservesLargeNumbers(t *testing.T) {
	// Integers larger than 2^53 lose precision when coerced to float64, and very large
	// values re-marshal in scientific notation. Parsing with UseNumber keeps the exact
	// digits, so IDs, counters and epoch timestamps stay accurate in the logged body.
	input := `{"id":9223372036854775807,"count":10000000000000000000,"certificatePassword":"hunter2"}`
	redacted := redactSensitiveFields(input)

	if strings.Contains(redacted, "hunter2") {
		t.Errorf("password leaked: %s", redacted)
	}
	if !strings.Contains(redacted, "9223372036854775807") {
		t.Errorf("large integer id lost precision or was reformatted: %s", redacted)
	}
	if !strings.Contains(redacted, "10000000000000000000") {
		t.Errorf("large integer count lost precision or was reformatted: %s", redacted)
	}
	if strings.Contains(redacted, "e+") || strings.Contains(redacted, "E+") {
		t.Errorf("number rewritten in scientific notation: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Errorf("expected [REDACTED] marker: %s", redacted)
	}
}

func TestRedactSensitiveFields_PreservesHTMLCharacters(t *testing.T) {
	// json.Marshal HTML-escapes '&', '<' and '>' by default (turning them into six-character
	// unicode escapes); re-encoding with SetEscapeHTML(false) keeps them literal so URLs and
	// messages in the logged body stay readable. If escaping were still on, the substrings
	// below would be broken up by the escapes and these Contains checks would fail.
	input := `{"url":"https://example.com/a?x=1&y=2","note":"a < b && b > c"}`
	redacted := redactSensitiveFields(input)

	if !strings.Contains(redacted, "x=1&y=2") {
		t.Errorf("'&' was HTML-escaped instead of preserved: %s", redacted)
	}
	if !strings.Contains(redacted, "a < b && b > c") {
		t.Errorf("'<' or '>' was HTML-escaped instead of preserved: %s", redacted)
	}
}
