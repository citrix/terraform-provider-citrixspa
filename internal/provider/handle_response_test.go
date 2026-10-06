package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHandleResponse_Non2xxErrorSurfacesRawBody(t *testing.T) {
	// Response bodies are not redacted (the SPA API never echoes request-only secrets back), so a
	// non-2xx error must surface the upstream body verbatim — including a non-JSON gateway/WAF/
	// auth-proxy page, which redaction would have withheld wholesale as unparseable. That withheld
	// body is exactly the diagnostic that is hardest to reconstruct, so it must survive alongside
	// the status code.
	req, err := http.NewRequest(http.MethodGet, "https://example.com/app", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body:       io.NopCloser(strings.NewReader("<html><body>502 Bad Gateway: upstream connect error</body></html>")),
		Header:     make(http.Header),
		Request:    req,
	}

	err = (&APIClient{}).handleResponse(context.Background(), resp, nil)
	if err == nil {
		t.Fatal("expected an error for a non-2xx response")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected status code in error: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "upstream connect error") {
		t.Errorf("expected raw upstream body preserved in error: %s", err.Error())
	}
	if strings.Contains(err.Error(), unparseableBodyMarker) {
		t.Errorf("non-JSON body was withheld instead of surfaced: %s", err.Error())
	}
}

func TestDebugLoggingEnabled(t *testing.T) {
	// Every case sets all five variables explicitly (the three TF_LOG* levels plus the two
	// acc-test log-path vars) so any ambient value in the runner's environment cannot leak into
	// the result. Cannot use t.Parallel with t.Setenv.
	tests := []struct {
		name         string
		citrixspa    string // TF_LOG_PROVIDER_CITRIXSPA (most specific)
		providerWide string // TF_LOG_PROVIDER
		global       string // TF_LOG (least specific)
		accLogPath   string // TF_ACC_LOG_PATH
		logPathMask  string // TF_LOG_PATH_MASK
		want         bool
	}{
		{name: "all unset", want: false},
		{name: "global debug", global: "DEBUG", want: true},
		{name: "global trace", global: "TRACE", want: true},
		{name: "global lowercase accepted", global: "debug", want: true},
		{name: "global json is trace-level", global: "JSON", want: true},
		{name: "global info is not debug", global: "INFO", want: false},
		{name: "global off is not debug", global: "OFF", want: false},
		{name: "global unknown defaults to info (debug off)", global: "verbose", want: false},
		{name: "provider-wide debug", providerWide: "DEBUG", want: true},
		{name: "provider-wide info overrides global debug", providerWide: "INFO", global: "DEBUG", want: false},
		{name: "citrixspa error suppresses broader debug", citrixspa: "ERROR", providerWide: "DEBUG", global: "TRACE", want: false},
		{name: "citrixspa off suppresses broader debug", citrixspa: "OFF", providerWide: "DEBUG", global: "TRACE", want: false},
		{name: "citrixspa debug wins over broader info", citrixspa: "DEBUG", providerWide: "INFO", global: "ERROR", want: true},
		{name: "acc log path enables debug when TF_LOG unset", accLogPath: "/tmp/acc.log", want: true},
		{name: "log path mask alone does not enable debug (sink stays off)", logPathMask: "/tmp/mask.log", want: false},
		{name: "explicit provider level still wins over acc log path", citrixspa: "ERROR", accLogPath: "/tmp/acc.log", want: false},
		{name: "acc log path overrides TF_LOG info disable", global: "INFO", accLogPath: "/tmp/acc.log", want: true},
		{name: "acc log path overrides TF_LOG off disable", global: "OFF", accLogPath: "/tmp/acc.log", want: true},
		{name: "acc log path overrides TF_LOG_PROVIDER info disable", providerWide: "INFO", accLogPath: "/tmp/acc.log", want: true},
		{name: "acc log path overrides unrecognized TF_LOG", global: "verbose", accLogPath: "/tmp/acc.log", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TF_LOG_PROVIDER_CITRIXSPA", tt.citrixspa)
			t.Setenv("TF_LOG_PROVIDER", tt.providerWide)
			t.Setenv("TF_LOG", tt.global)
			t.Setenv("TF_ACC_LOG_PATH", tt.accLogPath)
			t.Setenv("TF_LOG_PATH_MASK", tt.logPathMask)
			if got := debugLoggingEnabled(); got != tt.want {
				t.Errorf("debugLoggingEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleResponse_Success2xxUnmarshals(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com/app", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"id":"app-1","certificatePassword":"hunter2"}`)),
		Header:     make(http.Header),
		Request:    req,
	}

	var target struct {
		ID                  string `json:"id"`
		CertificatePassword string `json:"certificatePassword"`
	}
	if err := (&APIClient{}).handleResponse(context.Background(), resp, &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.ID != "app-1" {
		t.Errorf("expected id app-1, got %q", target.ID)
	}
	if target.CertificatePassword != "hunter2" {
		t.Errorf("expected raw value to reach target regardless of redaction, got %q", target.CertificatePassword)
	}
}
