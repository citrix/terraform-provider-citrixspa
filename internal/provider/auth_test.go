package provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fastTokenRetry shrinks the token-retry backoff for the duration of a test so
// retry paths are exercised without real-time sleeps.
func fastTokenRetry(t *testing.T) {
	t.Helper()
	orig := tokenRetryBaseDelay
	tokenRetryBaseDelay = time.Millisecond
	t.Cleanup(func() { tokenRetryBaseDelay = orig })
}

// newTestAuthClient builds an AuthClient whose HTTP transport is driven by fn.
func newTestAuthClient(fn roundTripFunc) *AuthClient {
	return &AuthClient{
		BaseURL:    "https://test.example.com",
		CustomerID: "cust-123",
		HTTPClient: &http.Client{Transport: fn},
	}
}

func TestGetBearerToken_RetriesOn429ThenSucceeds(t *testing.T) {
	fastTokenRetry(t)
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		attempt := atomic.AddInt32(&calls, 1)
		if attempt == 1 {
			header := make(http.Header)
			header.Set("Retry-After", "0")
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"rate limited"}`))),
				Header:     header,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"token_type":"bearer","access_token":"tok-abc","expires_in":"3600"}`))),
			Header:     make(http.Header),
		}, nil
	})

	tok, err := a.GetBearerToken(context.Background(), "client-id", "client-secret")
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if tok.AccessToken != "tok-abc" {
		t.Errorf("expected access_token 'tok-abc', got %q", tok.AccessToken)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 calls (429 then 200), got %d", got)
	}
}

func TestGetBearerToken_RetriesOn5xxThenSucceeds(t *testing.T) {
	fastTokenRetry(t)
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		attempt := atomic.AddInt32(&calls, 1)
		if attempt <= 2 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"unavailable"}`))),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"token_type":"bearer","access_token":"tok-xyz","expires_in":"3600"}`))),
			Header:     make(http.Header),
		}, nil
	})

	tok, err := a.GetBearerToken(context.Background(), "client-id", "client-secret")
	if err != nil {
		t.Fatalf("expected success after 5xx retries, got error: %v", err)
	}
	if tok.AccessToken != "tok-xyz" {
		t.Errorf("expected access_token 'tok-xyz', got %q", tok.AccessToken)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 calls (two 503s then 200), got %d", got)
	}
}

func TestGetBearerToken_DoesNotRetryOn401(t *testing.T) {
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"invalid credentials"}`))),
			Header:     make(http.Header),
		}, nil
	})

	if _, err := a.GetBearerToken(context.Background(), "client-id", "client-secret"); err == nil {
		t.Fatal("expected error on 401, got nil")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 call (no retry on 401), got %d", got)
	}
}

func TestGetBearerToken_RetriesOnNetworkErrorThenSucceeds(t *testing.T) {
	fastTokenRetry(t)
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			return nil, errors.New("connection reset by peer")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"token_type":"bearer","access_token":"tok-net","expires_in":"3600"}`))),
			Header:     make(http.Header),
		}, nil
	})

	tok, err := a.GetBearerToken(context.Background(), "client-id", "client-secret")
	if err != nil {
		t.Fatalf("expected success after network-error retry, got error: %v", err)
	}
	if tok.AccessToken != "tok-net" {
		t.Errorf("expected access_token 'tok-net', got %q", tok.AccessToken)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 calls (network error then 200), got %d", got)
	}
}

func TestGetBearerToken_PersistentNetworkErrorFails(t *testing.T) {
	fastTokenRetry(t)
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return nil, errors.New("connection refused")
	})

	_, err := a.GetBearerToken(context.Background(), "client-id", "client-secret")
	if err == nil {
		t.Fatal("expected error after exhausting network-error retries, got nil")
	}
	if !strings.Contains(err.Error(), "failed to obtain token") {
		t.Errorf("expected wrapped 'failed to obtain token' error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 4 {
		t.Errorf("expected 4 calls (maxRetries+1), got %d", got)
	}
}

func TestGetBearerToken_RetryExhaustionOn503(t *testing.T) {
	fastTokenRetry(t)
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"unavailable"}`))),
			Header:     make(http.Header),
		}, nil
	})

	_, err := a.GetBearerToken(context.Background(), "client-id", "client-secret")
	if err == nil {
		t.Fatal("expected error after exhausting 503 retries, got nil")
	}
	if !strings.Contains(err.Error(), "status 503") {
		t.Errorf("expected 'status 503' error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 4 {
		t.Errorf("expected 4 calls (maxRetries+1), got %d", got)
	}
}

func TestGetBearerToken_NegativeRetryAfterDoesNotBusyRetry(t *testing.T) {
	fastTokenRetry(t)
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			header := make(http.Header)
			header.Set("Retry-After", "-1") // negative must be clamped to 0, not fire a negative timer
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"rate limited"}`))),
				Header:     header,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"token_type":"bearer","access_token":"tok-neg","expires_in":"3600"}`))),
			Header:     make(http.Header),
		}, nil
	})

	tok, err := a.GetBearerToken(context.Background(), "client-id", "client-secret")
	if err != nil {
		t.Fatalf("expected success after negative-Retry-After retry, got error: %v", err)
	}
	if tok.AccessToken != "tok-neg" {
		t.Errorf("expected access_token 'tok-neg', got %q", tok.AccessToken)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 calls (429 with negative Retry-After then 200), got %d", got)
	}
}

func TestGetBearerToken_ContextCancelledDuringBackoff(t *testing.T) {
	// Default (1s) backoff so the wait is long; the transport cancels the context
	// on the first response, so the backoff select aborts via ctx.Done immediately
	// and no second request is made.
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		cancel()
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"rate limited"}`))),
			Header:     make(http.Header),
		}, nil
	})

	_, err := a.GetBearerToken(ctx, "client-id", "client-secret")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 call (cancelled during backoff, no retry), got %d", got)
	}
}

func TestGetBearerToken_ContextCancelledDuringRequest(t *testing.T) {
	// Simulate the context being cancelled while the request is in flight: Do
	// returns the context error. The network-error branch must surface it
	// directly (no misleading "retrying") and make no further request.
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		cancel()
		return nil, context.Canceled
	})

	_, err := a.GetBearerToken(ctx, "client-id", "client-secret")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 call (cancelled during request, no retry), got %d", got)
	}
}

func TestGetBearerToken_MalformedSuccessBodyFails(t *testing.T) {
	var calls int32
	a := newTestAuthClient(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{not valid json`))),
			Header:     make(http.Header),
		}, nil
	})

	if _, err := a.GetBearerToken(context.Background(), "client-id", "client-secret"); err == nil {
		t.Fatal("expected parse error on malformed 200 body, got nil")
	} else if !strings.Contains(err.Error(), "failed to parse token response") {
		t.Errorf("expected 'failed to parse token response', got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 call (no retry on parse error), got %d", got)
	}
}

func TestRetryAfterWait(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	base := 2 * time.Second
	max := 30 * time.Second

	tests := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{"empty falls back to base", "", base},
		{"unparseable falls back to base", "not-a-number", base},
		{"numeric seconds honored", "5", 5 * time.Second},
		{"negative clamped to zero", "-1", 0},
		{"oversized numeric clamped to max", "3600", max},
		{"overflow-sized numeric clamped to max", "9999999999", max},
		{"future http-date honored", now.Add(10 * time.Second).Format(http.TimeFormat), 10 * time.Second},
		{"past http-date clamped to zero", now.Add(-10 * time.Second).Format(http.TimeFormat), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := retryAfterWait(tc.header, now, base, max); got != tc.want {
				t.Errorf("retryAfterWait(%q) = %v, want %v", tc.header, got, tc.want)
			}
		})
	}
}
