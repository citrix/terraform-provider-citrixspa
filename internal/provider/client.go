package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

// PerformanceMetrics tracks API request metrics for bulk/scale testing.
// All counters use atomic operations for safe concurrent access.
type PerformanceMetrics struct {
	TotalRequests       atomic.Int64
	SuccessfulRequests  atomic.Int64
	FailedRequests      atomic.Int64
	RateLimitHits       atomic.Int64 // 429 responses received
	RateLimitRetryOK    atomic.Int64 // 429s that recovered after retry
	RateLimitRetryFail  atomic.Int64 // 429s that exhausted all retries
	TotalRateLimitDelay atomic.Int64 // Nanoseconds spent waiting on internal rate limiter
	TotalRetryDelay     atomic.Int64 // Nanoseconds spent sleeping for 429 Retry-After
	StartTime           time.Time

	// Error counts by HTTP status code
	errorCounts   map[int]*atomic.Int64
	errorCountsMu sync.Mutex
}

// NewPerformanceMetrics creates an initialized PerformanceMetrics instance.
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		StartTime:   time.Now(),
		errorCounts: make(map[int]*atomic.Int64),
	}
}

// RecordError safely increments the counter for a given HTTP status code.
func (m *PerformanceMetrics) RecordError(statusCode int) {
	m.errorCountsMu.Lock()
	counter, ok := m.errorCounts[statusCode]
	if !ok {
		counter = &atomic.Int64{}
		m.errorCounts[statusCode] = counter
	}
	m.errorCountsMu.Unlock()
	counter.Add(1)
}

// LogSummary outputs the aggregated performance metrics via tflog.
func (m *PerformanceMetrics) LogSummary(ctx context.Context) {
	elapsed := time.Since(m.StartTime)
	rateLimitDelayMs := m.TotalRateLimitDelay.Load() / int64(time.Millisecond)
	retryDelayMs := m.TotalRetryDelay.Load() / int64(time.Millisecond)

	fields := map[string]any{
		"total_requests":           m.TotalRequests.Load(),
		"successful_requests":      m.SuccessfulRequests.Load(),
		"failed_requests":          m.FailedRequests.Load(),
		"rate_limit_429_hits":      m.RateLimitHits.Load(),
		"rate_limit_retry_success": m.RateLimitRetryOK.Load(),
		"rate_limit_retry_fail":    m.RateLimitRetryFail.Load(),
		"rate_limiter_delay_ms":    rateLimitDelayMs,
		"retry_after_delay_ms":     retryDelayMs,
		"elapsed_seconds":          elapsed.Seconds(),
	}

	// Append error breakdown
	m.errorCountsMu.Lock()
	for code, counter := range m.errorCounts {
		fields[fmt.Sprintf("http_%d_count", code)] = counter.Load()
	}
	m.errorCountsMu.Unlock()

	tflog.Info(ctx, "spa-terraform-provider: === PERFORMANCE METRICS SUMMARY ===", fields)
}

// LogPeriodicSummary logs a summary every summaryInterval requests.
const metricsSummaryInterval = 10

func (m *PerformanceMetrics) LogPeriodicSummary(ctx context.Context) {
	total := m.TotalRequests.Load()
	if total > 0 && total%metricsSummaryInterval == 0 {
		m.LogSummary(ctx)
	}
}

// SPAClient defines the interface for SPA API operations
type SPAClient interface {
	makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error)

	// Application management methods
	GetApplications(ctx context.Context, offset, limit int, name, appType string) (*ApplicationsResponse, error)
	GetApplication(ctx context.Context, id string) (*Application, error)
	GetApplicationAwaitSSO(ctx context.Context, id string) (*Application, error)
	CreateApplication(ctx context.Context, app *Application) (*Application, error)
	UpdateApplication(ctx context.Context, id string, app *Application) error
	DeleteApplication(ctx context.Context, id string) error
	CompleteApplication(ctx context.Context, id string) error
	AssignCertificateToApplication(ctx context.Context, applicationID, domain string, cert *Certificate) error
	UnassignCertificateFromApplication(ctx context.Context, applicationID, domain string) error

	// Detailed Application listing method (only available on AuthenticatedClient)
	// This interface method allows the applications data source to request detailed info when supported
	GetApplicationsDetailed(ctx context.Context, offset, limit int, name, appType string, detailed bool) (*ApplicationsResponse, error)

	// Access Policy management methods
	GetAccessPolicies(ctx context.Context, offset, limit int, name, orderBy string) (*AccessPoliciesResponse, error)
	GetAccessPolicy(ctx context.Context, id string) (*AccessPolicy, error)
	CreateAccessPolicy(ctx context.Context, policy *AccessPolicy) (*AccessPolicy, error)
	UpdateAccessPolicy(ctx context.Context, id string, policy *AccessPolicy) error
	DeleteAccessPolicy(ctx context.Context, id string) error

	// Detailed Access Policy listing method (only available on AuthenticatedClient)
	// This interface method allows the access policies data source to request detailed info when supported
	GetAccessPoliciesDetailed(ctx context.Context, offset, limit int, name, orderBy string, detailed bool) (*AccessPoliciesResponse, error)

	// Security Group management methods
	GetSecurityGroups(ctx context.Context, offset, limit int) (*SecurityGroupsResponse, error)
	GetSecurityGroup(ctx context.Context, id string) (*SecurityGroup, error)
	CreateSecurityGroup(ctx context.Context, sg *SecurityGroup) (*SecurityGroup, error)
	UpdateSecurityGroup(ctx context.Context, id string, sg *SecurityGroup) error
	DeleteSecurityGroup(ctx context.Context, id string) error

	// Routing Domain management methods
	GetRoutingDomains(ctx context.Context, offset, limit int) (*RoutingDomainsResponse, error)
	GetRoutingDomain(ctx context.Context, fqdn string) (*RoutingDomain, error)
	CreateRoutingDomain(ctx context.Context, rd *RoutingDomain) (*RoutingDomain, error)
	UpdateRoutingDomain(ctx context.Context, fqdn string, rd *RoutingDomain) error
	DeleteRoutingDomain(ctx context.Context, fqdn string) error

	// Certificate management methods
	GetCertificates(ctx context.Context, offset, limit int) (*CertificatesResponse, error)
	CreateCertificate(ctx context.Context, cert *Certificate) (*Certificate, error)
	DeleteCertificate(ctx context.Context, id string) error

	// Browser Mode methods
	GetBrowserMode(ctx context.Context) (*BrowserMode, error)

	// Hybrid Config methods
	GetHybridConfig(ctx context.Context) (*HybridConfig, error)

	// Last Activity methods
	GetLastActivity(ctx context.Context) (*LastActivity, error)

	// Terminate Machine Access methods
	GetTerminateMachineAccess(ctx context.Context, offset, limit int) (*TerminateMachineAccessResponse, error)
	GetTerminateUserAccess(ctx context.Context, offset, limit int) (*TerminateUserAccessResponse, error)
	GetTerminateMachineAccessByID(ctx context.Context, id string) (*TerminateMachineAccess, error)
	CreateTerminateMachineAccess(ctx context.Context, machine *TerminateMachineAccess) (*TerminateMachineAccess, error)
	DeleteTerminateMachineAccess(ctx context.Context, id string) error

	// Terminate User Access methods
	GetTerminateUserAccessByID(ctx context.Context, id string) (*TerminateUserAccess, error)
	CreateTerminateUserAccess(ctx context.Context, user *TerminateUserAccess) (*TerminateUserAccess, error)
	UpdateTerminateUserAccess(ctx context.Context, id string, user *TerminateUserAccess) error
	DeleteTerminateUserAccess(ctx context.Context, id string) error

	// Session Policy management methods
	GetSessionPolicies(ctx context.Context, offset, limit int, name, orderBy string) (*SessionPoliciesResponse, error)
	GetSessionPolicy(ctx context.Context, id string) (*SessionPolicy, error)
	CreateSessionPolicy(ctx context.Context, policy *SessionPolicy) (*SessionPolicy, error)
	UpdateSessionPolicy(ctx context.Context, id string, policy *SessionPolicy) error
	DeleteSessionPolicy(ctx context.Context, id string) error
}

type TokenProvider interface {
	// getToken returns a valid auth token for the API client
	GetToken(ctx context.Context) (string, error)
}

// APIClient is a client for the SPA API
type APIClient struct {
	BaseURL                  string
	CustomerID               string
	AuthToken                string
	HTTPClient               *http.Client
	Limiter                  *rate.Limiter       // Rate limiter for API requests
	Semaphore                *semaphore.Weighted // Concurrency limiter for parallel requests
	tokenProvider            TokenProvider       // Token provider for getting auth tokens
	FetchDetailsOnList       bool                // When true, detailed listing methods will fetch individual item details
	SuppressASBNotifications bool                // When true, suppress ASB notifications on API requests
	UserAgent                string              // Custom User-Agent header for API requests
	Metrics                  *PerformanceMetrics
}

// Ensure APIClient implements SPAClient
var _ SPAClient = (*APIClient)(nil)

func NewAPIClient(baseURL, customerID, authToken string, limiter *rate.Limiter, maxConcurrent int64, fetchDetailsOnList bool, suppressASBNotifications bool, tp TokenProvider, userAgent string) *APIClient {
	var sem *semaphore.Weighted
	if maxConcurrent > 0 {
		sem = semaphore.NewWeighted(maxConcurrent)
	}

	p := &APIClient{
		BaseURL:    strings.TrimSuffix(baseURL, "/"), // Ensure no trailing slash
		CustomerID: customerID,
		AuthToken:  authToken,
		HTTPClient: &http.Client{
			Timeout: 90 * time.Second,
			// Disable automatic redirect following so that 307 regional redirects
			// are intercepted in makeRequest and surfaced as actionable errors.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Limiter:                  limiter,
		Semaphore:                sem,
		FetchDetailsOnList:       fetchDetailsOnList,       // Set the flag for detailed listing
		SuppressASBNotifications: suppressASBNotifications, // Set the flag for suppressing ASB notifications
		tokenProvider:            tp,                       // Set the token provider for dynamic token management
		UserAgent:                userAgent,
		Metrics:                  NewPerformanceMetrics(),
	}

	if p.tokenProvider == nil {
		p.tokenProvider = p // Fallback to self if no provider is set
	}
	return p
}

func (c *APIClient) GetToken(ctx context.Context) (string, error) {
	return c.AuthToken, nil
}

// redirectErrorResponse models the structured JSON body returned by the API
// when a 307 Temporary Redirect is issued during data regionalization.
type redirectErrorResponse struct {
	Type       string `json:"type"`
	Detail     string `json:"detail"`
	Parameters []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"parameters"`
}

// sensitiveKeyExact holds the exact JSON keys whose values are redacted. These are the wire field
// names of the only secret-bearing fields in the SPA API request bodies: the PKCS#12 blob
// (`certificate`) and its password (`certificatePassword`). The API never echoes these back, so
// response bodies do not carry them.
//
// The API bodies are fully typed, so the secret surface is closed and enumerable. Matching is an
// exact, case-sensitive comparison against these two keys — no case-folding or separator
// normalization. That means non-secret keys that merely resemble them (certificateId,
// certificateName, tagKey, keywords) are never over-redacted, while a certificate/certificatePassword
// nested inside a free-form map[string]any field (customProperties, customerDomainFields, sso, data,
// and access-policy rule / session-policy condition metadata) is still redacted when it uses one of
// these exact keys. When a new secret field is added to the schema, add its wire key here.
var sensitiveKeyExact = []string{"certificate", "certificatePassword"}

// isSensitiveKey reports whether a JSON key's value must be redacted: an exact match against the
// enumerated secret keys in sensitiveKeyExact.
func isSensitiveKey(k string) bool {
	for _, e := range sensitiveKeyExact {
		if k == e {
			return true
		}
	}
	return false
}

// unparseableBodyMarker replaces a request body that cannot be parsed as JSON. A
// regex-based scrub cannot reliably delimit a value in malformed/non-JSON content (escaped
// quotes, embedded spaces, truncation all defeat it and leak partial secrets), so an
// unparseable body is withheld in full rather than emitted with best-effort redaction.
const unparseableBodyMarker = "[REDACTED: body was not valid JSON and was withheld to avoid leaking secrets]"

// tfLogEnvVars lists the environment variables that gate whether Terraform surfaces this
// provider's debug logs, ordered most specific first. TF_LOG_PROVIDER_CITRIXSPA is the exact
// variable terraform-plugin-go wires to this provider's root logger (TF_LOG_PROVIDER + "_" +
// provider type "citrixspa"); TF_LOG_PROVIDER then TF_LOG are Terraform's broader controls. The
// first one set decides the level; TF_ACC_LOG_PATH can still override a broad-variable
// (TF_LOG / TF_LOG_PROVIDER) disable (see below).
var tfLogEnvVars = []string{"TF_LOG_PROVIDER_CITRIXSPA", "TF_LOG_PROVIDER", "TF_LOG"}

// debugLoggingEnabled reports whether the provider's debug logs will actually be surfaced by
// Terraform, so hot-path callers can skip redactSensitiveFields (a full JSON parse + tree walk +
// re-encode) when they would not be — the production default. tflog.Debug's arguments are evaluated
// eagerly, so without this gate redaction would run on every request even when the line is discarded.
//
// The provider's own logger level is not a usable signal: in the real runtime there is no sink, so
// NewRootProviderLogger defaults an unset level to hclog.Trace and tflog.Debug always writes —
// Terraform core discards it when TF_LOG* is unset. helper/logging.IsDebugOrHigher is not the
// reference either (it reads only TF_LOG, not the provider-scoped vars), so we mirror core's TF_LOG*
// precedence directly.
//
// Level semantics follow tfsdklog/sink.go (terraform-plugin-log v0.10.0), the sink that governs
// tflog.Debug output: DEBUG/TRACE (case-insensitive) and JSON enable it; INFO/WARN/ERROR/OFF
// disable it, as does any unrecognized value — tfsdklog warns it will default such a value to OFF
// but actually leaves logLevel at NoLevel, which hclog coerces to its DefaultLevel (INFO), so DEBUG
// is still dropped. We mirror that: only DEBUG/TRACE/JSON enable the body.
func debugLoggingEnabled() bool {
	for _, name := range tfLogEnvVars {
		level := strings.ToUpper(strings.TrimSpace(os.Getenv(name)))
		if level == "" {
			continue
		}
		switch level {
		case "DEBUG", "TRACE", "JSON":
			// DEBUG/TRACE enable debug output; JSON is trace-level JSON output. Debug on.
			return true
		default:
			// INFO/WARN/ERROR/OFF, or an unrecognized value (which tfsdklog warns it sets to OFF but
			// actually leaves at hclog's default, INFO — see above): DEBUG is dropped, so debug is
			// off. One exception: an acc-log path forces the sink to TRACE. In acc-test mode the
			// in-process sink reads only TF_LOG + TF_ACC_LOG_PATH, and the provider logger's level
			// comes solely from TF_LOG_PROVIDER_CITRIXSPA; bare TF_LOG and TF_LOG_PROVIDER are read by
			// nothing in-process, so neither can suppress the line and the body must still be written.
			// The override therefore excludes only TF_LOG_PROVIDER_CITRIXSPA: an explicit disable there
			// lowers the provider logger itself, dropping the line at the source where the acc path
			// cannot rescue it.
			if name != "TF_LOG_PROVIDER_CITRIXSPA" && accLogPathSet() {
				return true
			}
			return false
		}
	}
	// No explicit TF_LOG* level: the acc-test log-path alone pins the sink to TRACE.
	return accLogPathSet()
}

// accLogPathSet reports whether TF_ACC_LOG_PATH is set. tfsdklog hardcodes the sink to TRACE when it
// is, regardless of TF_LOG, so the provider's debug lines — and their (redacted) bodies — reach the
// acc-test log file. TF_LOG_PATH_MASK is deliberately not treated as a debug signal: its tfsdklog
// branch only renames the file and never raises the level, so honoring it would run redaction to
// produce a line the (off) sink discards.
func accLogPathSet() bool {
	return strings.TrimSpace(os.Getenv("TF_ACC_LOG_PATH")) != ""
}

// redactSensitiveFields redacts sensitive HTTP request body content before it is written to the
// provider's debug log (it is only applied on the request path, the sole place these secrets
// travel). A key's value is redacted when isSensitiveKey reports it sensitive: an exact match
// against the schema's known secret keys (certificate, certificatePassword). Any body that fails
// JSON parsing is replaced wholesale with unparseableBodyMarker so redaction fails closed rather
// than emitting the body verbatim.
func redactSensitiveFields(bodyContent string) string {
	// Empty or whitespace-only body returns as-is
	if strings.TrimSpace(bodyContent) == "" {
		return bodyContent
	}

	// Attempt to parse as JSON. UseNumber keeps numeric values in their exact textual
	// form, so large integer IDs/counters and epoch timestamps are preserved verbatim
	// instead of being rounded through float64 or rewritten in scientific notation in
	// the logged request body.
	dec := json.NewDecoder(strings.NewReader(bodyContent))
	dec.UseNumber()
	var data any
	if err := dec.Decode(&data); err != nil {
		// Not JSON or malformed → fail closed. A regex scrub cannot safely delimit values in
		// malformed content (escaped quotes / spaces / truncation leak partial secrets), so
		// withhold the whole request body from the debug log.
		return unparseableBodyMarker
	}
	// Enforce single-value strictness: json.Decoder (unlike json.Unmarshal) accepts trailing
	// content, and dec.More() misses a trailing '}'/']' (it returns false on those bytes).
	// Requiring a second Decode to return io.EOF fails closed on ANY trailing token (syntax
	// error or a second value → not io.EOF; only a clean end-of-input yields io.EOF).
	var discard any
	if err := dec.Decode(&discard); err != io.EOF {
		return unparseableBodyMarker
	}

	// Redact sensitive fields in place. redactSensitiveFields owns `data` exclusively (it was just
	// decoded from a fresh reader), so redactValue mutates the decoded tree directly instead of
	// allocating a second parallel copy.
	redactedData := redactValue(data)

	// Re-marshal with SetEscapeHTML(false) so '&', '<' and '>' in URLs/messages survive verbatim
	// instead of becoming six-character unicode escapes (json.Marshal escapes them by default).
	// Note: this sorts keys alphabetically and normalizes whitespace.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(redactedData); err != nil {
		// If re-marshaling fails (unlikely), fail closed: withhold the body rather than
		// returning the original, which could put the unredacted secrets back into the debug log.
		return unparseableBodyMarker
	}

	// json.Encoder.Encode appends a trailing newline; drop it so output matches json.Marshal.
	return strings.TrimSuffix(buf.String(), "\n")
}

// redactValue replaces sensitive-key values in place; the caller must own the decoded tree.
func redactValue(obj any) any {
	switch v := obj.(type) {
	case map[string]any:
		for k, val := range v {
			if isSensitiveKey(k) {
				v[k] = "[REDACTED]"
			} else {
				v[k] = redactValue(val)
			}
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = redactValue(item)
		}
		return v
	default:
		return v
	}
}

// makeRequest performs an HTTP request with proper headers and error handling
func (c *APIClient) makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	// Acquire semaphore slot for mutating operations only (POST, PUT, DELETE).
	// GET requests are not limited as the backend handles read operations more efficiently.
	// NOTE: the slot is deliberately held for the entire call — including the 429
	// retry/backoff loop below — so that max_concurrent also bounds retry throughput
	// and applies global backpressure when the API is rate-limiting us, rather than
	// releasing the slot for another request to pile on during backoff.
	if c.Semaphore != nil && method != http.MethodGet {
		if err := c.Semaphore.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to acquire concurrency semaphore: %w", err)
		}
		defer c.Semaphore.Release(1)
	}

	requestStart := time.Now()
	c.Metrics.TotalRequests.Add(1)

	// Get valid token before making the request
	token, err := c.tokenProvider.GetToken(ctx)
	if err != nil {
		c.Metrics.FailedRequests.Add(1)
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	var reqBody io.Reader
	var bodyContent string
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			c.Metrics.FailedRequests.Add(1)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyBytes)
		bodyContent = string(bodyBytes)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		c.Metrics.FailedRequests.Add(1)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers with the obtained token
	headers := map[string]string{
		"Content-Type":           "application/json; charset=utf-8",
		"Accept":                 "application/json",
		"Citrix-CustomerId":      c.CustomerID, // Fixed header name from Citrix-Customerid
		"Authorization":          fmt.Sprintf("CWSAuth bearer=%s", token),
		"Cache-Control":          "no-cache, no-store",
		"X-Content-Type-Options": "nosniff",
	}

	if c.UserAgent != "" {
		headers["User-Agent"] = c.UserAgent
	}

	if c.SuppressASBNotifications {
		headers["X-Send-ASB-Notification"] = "false"
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add transaction ID for tracking
	transactionID := uuid.New().String()
	req.Header.Set("Citrix-TransactionId", transactionID)

	// Log the request for debugging. The request body only feeds this debug log, so skip the
	// cost of redactSensitiveFields entirely unless the log will actually be surfaced — the
	// argument would otherwise be evaluated eagerly on every request even in production.
	requestFields := map[string]any{
		"method":         method,
		"url":            fullURL,
		"transaction_id": transactionID,
		// "headers":        req.Header,
	}
	if debugLoggingEnabled() {
		requestFields["body"] = redactSensitiveFields(bodyContent)
	}
	tflog.Debug(ctx, "spa-terraform-provider: SPA API request", requestFields)

	const maxRetries = 3
	var rateLimitHitsThisRequest int64
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Rate limit every attempt (initial + retries) to avoid bursts
		if c.Limiter != nil {
			waitStart := time.Now()
			if err := c.Limiter.Wait(ctx); err != nil {
				c.Metrics.FailedRequests.Add(1)
				return nil, fmt.Errorf("rate limit exceeded (transaction ID: %s): %w", transactionID, err)
			}
			waited := time.Since(waitStart)
			if waited > time.Millisecond {
				c.Metrics.TotalRateLimitDelay.Add(int64(waited))
				tflog.Debug(ctx, "spa-terraform-provider: Request delayed due to rate limiting", map[string]any{
					"delay_ms":       waited.Milliseconds(),
					"method":         method,
					"url":            fullURL,
					"transaction_id": transactionID,
				})
			}
		}

		// Rebuild request body for retries since it's consumed after each attempt
		if attempt > 0 && body != nil {
			req.Body = io.NopCloser(bytes.NewBufferString(bodyContent))
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.Metrics.FailedRequests.Add(1)
			return nil, fmt.Errorf("failed to make request (transaction ID: %s): %w", transactionID, err)
		}

		// Detect regional redirect (307 Temporary Redirect) and surface an actionable error.
		// The API returns this when the configured base_url does not match the customer's data region.
		if resp.StatusCode == http.StatusTemporaryRedirect {
			redirectBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			location := resp.Header.Get("Location")

			var redirectErr redirectErrorResponse
			_ = json.Unmarshal(redirectBody, &redirectErr)

			customerDataRegion := ""
			currentAPIMRegion := ""
			for _, p := range redirectErr.Parameters {
				switch p.Name {
				case "customerDataRegion":
					customerDataRegion = p.Value
				case "currentAPIMRegion":
					currentAPIMRegion = p.Value
				}
			}

			detail := redirectErr.Detail
			if detail == "" {
				detail = "Regional customer must use their designated regional endpoint"
			}

			// Derive the correct base_url by swapping only the host of the configured
			// BaseURL with the host from the Location header. Scheme and path are always
			// taken from the trusted BaseURL, so empty-scheme or scheme-relative Location
			// values are never an issue. The API only redirects to known regional
			// *.cloud.com hosts; anything else is rejected to prevent a compromised
			// upstream from misleading the operator into a malicious URL.
			correctBaseURL := c.BaseURL
			if location != "" {
				if loc, parseErr := url.Parse(location); parseErr == nil && loc.Host != "" {
					if !strings.HasSuffix(loc.Host, ".cloud.com") {
						tflog.Warn(ctx, "spa-terraform-provider: 307 Location header points to an untrusted host, ignoring derived base_url", map[string]any{
							"location":       location,
							"transaction_id": transactionID,
						})
					} else if current, parseErr := url.Parse(c.BaseURL); parseErr == nil {
						current.Host = loc.Host
						correctBaseURL = current.String()
					}
				}
			}

			tflog.Error(ctx, "spa-terraform-provider: Regional redirect detected — update base_url in provider configuration", map[string]any{
				"current_base_url":     c.BaseURL,
				"correct_base_url":     correctBaseURL,
				"customer_data_region": customerDataRegion,
				"current_apim_region":  currentAPIMRegion,
				"location":             location,
				"transaction_id":       transactionID,
			})

			regionInfo := ""
			if customerDataRegion != "" && currentAPIMRegion != "" {
				regionInfo = fmt.Sprintf(" Your account data is in region %q but you are connecting via the %q region endpoint.", customerDataRegion, currentAPIMRegion)
			} else if customerDataRegion != "" {
				regionInfo = fmt.Sprintf(" Your account data is in region %q.", customerDataRegion)
			}

			return nil, fmt.Errorf(
				"API endpoint has moved (307 Temporary Redirect).%s\n"+
					"Update your provider configuration:\n"+
					"  Current base_url: %s\n"+
					"  Correct base_url: %s\n"+
					"Detail: %s (transaction ID: %s)",
				regionInfo,
				c.BaseURL,
				correctBaseURL,
				detail,
				transactionID,
			)
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			// Track success/failure based on status code
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				c.Metrics.SuccessfulRequests.Add(1)
				if rateLimitHitsThisRequest > 0 {
					c.Metrics.RateLimitRetryOK.Add(1)
				}
			} else {
				c.Metrics.FailedRequests.Add(1)
				c.Metrics.RecordError(resp.StatusCode)
			}
			// Log per-request timing at Debug level to avoid noise during normal runs
			tflog.Debug(ctx, "spa-terraform-provider: request completed", map[string]any{
				"method":              method,
				"url":                 fullURL,
				"status":              resp.StatusCode,
				"duration_ms":         time.Since(requestStart).Milliseconds(),
				"rate_limit_429_hits": rateLimitHitsThisRequest,
				"transaction_id":      transactionID,
			})
			c.Metrics.LogPeriodicSummary(ctx)
			return resp, nil
		}

		// Handle 429 Too Many Requests
		tflog.Warn(ctx, "spa-terraform-provider: 429 rate limit hit", map[string]any{
			"attempt":        attempt + 1,
			"max_retries":    maxRetries,
			"method":         method,
			"url":            fullURL,
			"transaction_id": transactionID,
		})

		// Drain and close the body so the transport can reuse the connection.
		rateLimitHitsThisRequest++
		c.Metrics.RateLimitHits.Add(1)
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		if attempt == maxRetries {
			c.Metrics.FailedRequests.Add(1)
			c.Metrics.RateLimitRetryFail.Add(1)
			c.Metrics.RecordError(http.StatusTooManyRequests)
			return nil, fmt.Errorf("API rate limit exceeded after %d retries (transaction ID: %s)", maxRetries, transactionID)
		}

		// Parse Retry-After header.
		retryAfter := resp.Header.Get("Retry-After")
		const maxRetryWait = 30 * time.Second
		waitDuration := maxRetryWait // Default wait if header is missing

		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				waitDuration = time.Duration(seconds) * time.Second
			} else if retryTime, err := http.ParseTime(retryAfter); err == nil {
				waitDuration = time.Until(retryTime)
				if waitDuration < 0 {
					waitDuration = 0
				}
			} else {
				tflog.Warn(ctx, "spa-terraform-provider: Received unparseable Retry-After header, using default retry delay", map[string]any{
					"retry_after":    retryAfter,
					"default_wait":   waitDuration.Seconds(),
					"attempt":        attempt + 1,
					"max_retries":    maxRetries,
					"method":         method,
					"url":            fullURL,
					"transaction_id": transactionID,
				})
			}
		}

		// Cap wait duration to avoid excessively long sleeps
		if waitDuration > maxRetryWait {
			tflog.Warn(ctx, "spa-terraform-provider: Retry-After exceeds maximum wait, capping to maxRetryWait", map[string]any{
				"retry_after_requested": waitDuration.Seconds(),
				"max_retry_wait":        maxRetryWait.Seconds(),
				"attempt":               attempt + 1,
				"max_retries":           maxRetries,
				"method":                method,
				"url":                   fullURL,
				"transaction_id":        transactionID,
			})
			waitDuration = maxRetryWait
		}

		tflog.Debug(ctx, "spa-terraform-provider: Received 429 Too Many Requests, retrying after delay", map[string]any{
			"retry_after_seconds": waitDuration.Seconds(),
			"attempt":             attempt + 1,
			"max_retries":         maxRetries,
			"method":              method,
			"url":                 fullURL,
			"transaction_id":      transactionID,
		})

		retryWaitStart := time.Now()
		timer := time.NewTimer(waitDuration)
		select {
		case <-timer.C:
			c.Metrics.TotalRetryDelay.Add(int64(time.Since(retryWaitStart)))
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			c.Metrics.FailedRequests.Add(1)
			return nil, fmt.Errorf("context cancelled while waiting for rate limit retry (transaction ID: %s): %w", transactionID, ctx.Err())
		}
	}

	// This should not be reached, but just in case
	c.Metrics.FailedRequests.Add(1)
	return nil, fmt.Errorf("unexpected state in rate limit retry loop (transaction ID: %s)", transactionID)
}

// ErrNotFound is the sentinel every HTTP 404 from the SPA API unwraps to. Use
// IsNotFound (or errors.Is) to test for it — never match on the error text: the
// message embeds a random transaction ID and the raw response body, either of
// which can contain "404" for a response that is not a 404 at all.
var ErrNotFound = errors.New("resource not found")

// APIError is a non-2xx response from the SPA API. It carries the parsed status
// code so callers never have to inspect the message text to classify a failure.
type APIError struct {
	StatusCode    int
	TransactionID string
	Body          string
}

// Error renders the message verbatim as it has always been rendered, so
// diagnostics, logs and existing assertions are unaffected by the type change.
func (e *APIError) Error() string {
	return fmt.Sprintf("API request failed with status %d (transaction ID: %s): %s", e.StatusCode, e.TransactionID, e.Body)
}

// Unwrap exposes ErrNotFound for 404s, so errors.Is works on wrapped errors.
func (e *APIError) Unwrap() error {
	if e.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	return nil
}

// IsNotFound reports whether err is an HTTP 404 from the SPA API. Transport
// failures, and non-404 responses whose transaction ID or body happens to
// contain "404", are correctly reported as false.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInternalServerError reports whether err is an HTTP 500 from the SPA API.
// Like IsNotFound it classifies on the parsed status code and never on the
// message text: that text embeds a random transaction ID and a verbatim echo of
// the caller's own payload, so it can read "status 500" for a response that is
// not a 500. Only 500 counts — other 5xx statuses come from the gateway rather
// than the service, and callers here care about partial work the service itself
// may have done.
func IsInternalServerError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusInternalServerError
}

// handleResponse processes API response and handles common error cases
func (c *APIClient) handleResponse(ctx context.Context, resp *http.Response, target any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	url := resp.Request.URL.String()
	txid := resp.Request.Header.Get("Citrix-TransactionId")

	// Log the response for debugging
	tflog.Debug(ctx, "spa-terraform-provider: SPA API response", map[string]any{
		"url":            url,
		"transaction_id": txid,
		"status":         resp.StatusCode,
		"body":           string(body),
		// "headers":     resp.Header,
		// "content_len": len(body),
	})

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode:    resp.StatusCode,
			TransactionID: txid,
			Body:          string(body),
		}
	}

	if target != nil && len(body) > 0 && string(body) != "null" {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Application API methods

// BrowserMode represents browser mode configuration
type BrowserMode struct {
	BrowserMode string `json:"browserMode"` // CEB or CEP
}

// HybridConfig represents hybrid configuration
type HybridConfig struct {
	FirstTime bool `json:"firstTime"`
	IsHybrid  bool `json:"isHybrid"`
}

// LastActivity represents last activity timestamp
type LastActivity struct {
	LastActivity float64 `json:"lastActivity"`
}

// TerminateMachineAccess represents machine access termination
type TerminateMachineAccess struct {
	ID          string `json:"id,omitempty"`
	AccountName string `json:"accountName,omitempty"`
	Name        string `json:"name,omitempty"`
	DNSHostName string `json:"dnsHostName,omitempty"`
	DomainName  string `json:"domainName,omitempty"`
	ObjectID    string `json:"objectId,omitempty"`
	IDPType     string `json:"idpType,omitempty"`
	// CreatedTime string `json:"createdTime,omitempty"`
	Duration int `json:"duration,omitempty"`
}

// TerminateUserAccess represents a user access termination record
type TerminateUserAccess struct {
	ID          string `json:"id,omitempty"`
	AccountName string `json:"accountName,omitempty"`
	Email       string `json:"email,omitempty"`
	DomainName  string `json:"domainName,omitempty"`
	ObjectID    string `json:"objectId,omitempty"`
	IDPType     string `json:"idpType,omitempty"`
	// CreatedTime string `json:"createdTime,omitempty"`
	Duration int `json:"duration,omitempty"`
}

// Policy represents a policy in the application
type Policy struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

// Application represents an application in the SPA system (for individual application queries)
type Application struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	// description and category enforce a backend minimum length and reject ""
	// ("is too short"); the backend also retains the prior value when the field is
	// omitted or null, so it cannot be cleared once set. An empty value is omitted
	// rather than sent; the schema uses UseStateForUnknown so an omitted value keeps
	// the prior one instead of drifting.
	Description string `json:"description,omitempty"`
	// url is required by the backend and always carries a non-empty value, so it is
	// sent unconditionally.
	URL      string `json:"url"`
	Category string `json:"category,omitempty"`
	// Boolean fields must NOT use omitempty: a false value is the Go zero value
	// and omitempty would drop it from the request body, leaving the backend's
	// previous value unchanged and causing "inconsistent result after apply".
	Hidden               bool           `json:"hidden"`
	AgentlessAccess      bool           `json:"agentlessAccess"`
	MobileSecurity       bool           `json:"mobileSecurity"`
	SbsOnlyLaunch        bool           `json:"sbsOnlyLaunch"`
	UsingTemplate        bool           `json:"usingTemplate"`
	TemplateName         string         `json:"templateName,omitempty"`
	Icon                 string         `json:"icon,omitempty"`
	IconURL              string         `json:"iconURL,omitempty"`
	RelatedURLs          []string       `json:"relatedURLs,omitempty"`
	Keywords             []string       `json:"keywords,omitempty"`
	Locations            []Location     `json:"locations,omitempty"`
	Policies             []Policy       `json:"policies,omitempty"`
	Destination          []Destination  `json:"destination,omitempty"`
	CustomProperties     map[string]any `json:"customProperties,omitempty"`
	CustomerDomainFields map[string]any `json:"customerDomainFields"`
	SSO                  map[string]any `json:"sso,omitempty"`
	CreatedTime          string         `json:"createdTime,omitempty"`
	State                string         `json:"state,omitempty"`
	PolicyCount          string         `json:"policyCount,omitempty"`
}

// ApplicationListItem represents an application in the applications listing response (where SSO is a string)
type ApplicationListItem struct {
	ID                   string         `json:"id,omitempty"`
	Name                 string         `json:"name"`
	Type                 string         `json:"type"`
	Description          string         `json:"description,omitempty"`
	URL                  string         `json:"url,omitempty"`
	Category             string         `json:"category,omitempty"`
	Hidden               bool           `json:"hidden,omitempty"`
	AgentlessAccess      bool           `json:"agentlessAccess,omitempty"`
	MobileSecurity       bool           `json:"mobileSecurity,omitempty"`
	SbsOnlyLaunch        bool           `json:"sbsOnlyLaunch,omitempty"`
	UsingTemplate        bool           `json:"usingTemplate,omitempty"`
	TemplateName         string         `json:"templateName,omitempty"`
	Icon                 string         `json:"icon,omitempty"`
	IconURL              string         `json:"iconURL,omitempty"`
	RelatedURLs          []string       `json:"relatedURLs,omitempty"`
	Keywords             []string       `json:"keywords,omitempty"`
	Locations            []Location     `json:"locations,omitempty"`
	Policies             []Policy       `json:"policies,omitempty"`
	Destination          []Destination  `json:"destination,omitempty"`
	CustomProperties     map[string]any `json:"customProperties,omitempty"`
	CustomerDomainFields map[string]any `json:"customerDomainFields,omitempty"`
	SSO                  string         `json:"sso,omitempty"`
	CreatedTime          string         `json:"createdTime,omitempty"`
	State                string         `json:"state,omitempty"`
	PolicyCount          string         `json:"policyCount,omitempty"`
}

// Location represents a location object with name and uuid
type Location struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

// Destination represents a destination configuration for ZTNA apps
type Destination struct {
	Destination string `json:"destination,omitempty"`
	Port        string `json:"port,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Subtype     string `json:"subtype,omitempty"`
}

// ApplicationsResponse represents the response from listing applications
type ApplicationsResponse struct {
	Applications []ApplicationListItem `json:"items,omitempty"`
	Total        int                   `json:"total,omitempty"`
	Count        int                   `json:"count,omitempty"`
	Offset       int                   `json:"offset,omitempty"`
}

// GetApplications retrieves a list of applications
func (c *APIClient) GetApplications(ctx context.Context, offset, limit int, name, appType string) (*ApplicationsResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if name != "" {
		params.Add("name", name)
	}
	if appType != "" {
		params.Add("type", appType)
	}

	path := "/applications"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetApplications calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result ApplicationsResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetApplicationsDetailed retrieves applications with optional detailed information
// When detailed=true OR when FetchDetailsOnList is enabled, it fetches full application details
// This reduces the number of API calls and avoids rate limiting issues by reusing the same auth token
func (c *APIClient) GetApplicationsDetailed(ctx context.Context, offset, limit int, name, appType string, detailed bool) (*ApplicationsResponse, error) {
	// Check if we should fetch details based on the flag or explicit request
	shouldFetchDetails := detailed || c.FetchDetailsOnList

	if !shouldFetchDetails {
		// Use standard listing
		return c.GetApplications(ctx, offset, limit, name, appType)
	}

	tflog.Info(ctx, "spa-terraform-provider: Fetching detailed applications list", map[string]interface{}{
		"offset":                offset,
		"limit":                 limit,
		"name":                  name,
		"appType":               appType,
		"detailed":              detailed,
		"fetch_details_on_list": c.FetchDetailsOnList,
		"should_fetch_details":  shouldFetchDetails,
	})

	// First get the list of applications
	apps, err := c.GetApplications(ctx, offset, limit, name, appType)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications list: %w", err)
	}

	// Convert ApplicationListItem to Application with full details
	// We'll fetch each application's full details in batch
	detailedApps := make([]ApplicationListItem, len(apps.Applications))

	// Use a channel to limit concurrent requests and avoid overwhelming the API
	const maxConcurrent = 5
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fetchErrors []error

	for i, app := range apps.Applications {
		wg.Add(1)
		go func(index int, appItem ApplicationListItem) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Get detailed application info using the same client
			fullApp, err := c.GetApplication(ctx, appItem.ID)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, fmt.Errorf("failed to get details for app %s (%s): %w", appItem.Name, appItem.ID, err))
				mu.Unlock()
				// Use the basic info we already have
				detailedApps[index] = appItem
				return
			}

			// if fullApp.RelatedURLs != nil {
			// 	slices.Sort(fullApp.RelatedURLs) // Ensure related URLs are sorted for consistency
			// }

			// Convert Application to ApplicationListItem format with full details
			mu.Lock()
			detailedApps[index] = ApplicationListItem{
				ID:                   fullApp.ID,
				Name:                 fullApp.Name,
				Type:                 fullApp.Type,
				Description:          fullApp.Description,
				URL:                  fullApp.URL,
				Category:             fullApp.Category,
				Hidden:               fullApp.Hidden,
				AgentlessAccess:      fullApp.AgentlessAccess,
				MobileSecurity:       fullApp.MobileSecurity,
				SbsOnlyLaunch:        fullApp.SbsOnlyLaunch,
				UsingTemplate:        fullApp.UsingTemplate,
				TemplateName:         fullApp.TemplateName,
				Icon:                 fullApp.Icon,
				IconURL:              fullApp.IconURL,
				RelatedURLs:          fullApp.RelatedURLs,
				Keywords:             fullApp.Keywords,
				Locations:            fullApp.Locations,
				Policies:             fullApp.Policies,
				Destination:          fullApp.Destination,
				CustomProperties:     fullApp.CustomProperties,
				CustomerDomainFields: fullApp.CustomerDomainFields,
				State:                fullApp.State,
				PolicyCount:          fullApp.PolicyCount,
				// Prefer value from individual GET; fall back to list value (e.g. ztna apps
				// return createdTime in list responses but not in individual GET responses).
				CreatedTime: func() string {
					if fullApp.CreatedTime != "" {
						return fullApp.CreatedTime
					}
					return appItem.CreatedTime
				}(),
			}

			// Convert SSO from map to string for ApplicationListItem
			if len(fullApp.SSO) > 0 {
				// Try to convert SSO map to a reasonable string representation
				if ssoBytes, err := json.Marshal(fullApp.SSO); err == nil {
					detailedApps[index].SSO = string(ssoBytes)
				}
			}
			mu.Unlock()
		}(i, app)
	}

	wg.Wait()

	// Log any fetch errors but don't fail the entire operation
	if len(fetchErrors) > 0 {
		tflog.Warn(ctx, "Some application details could not be fetched", map[string]interface{}{
			"error_count": len(fetchErrors),
			"errors":      fmt.Sprintf("%v", fetchErrors),
		})
	}

	// Return the response with detailed applications
	return &ApplicationsResponse{
		Applications: detailedApps,
		Total:        apps.Total,
		Count:        apps.Count,
		Offset:       apps.Offset,
	}, nil
}

// GetApplication retrieves a specific application by ID
func (c *APIClient) GetApplication(ctx context.Context, id string) (*Application, error) {
	path := fmt.Sprintf("/applications/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetApplication calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result Application
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	// if result.RelatedURLs != nil {
	// 	slices.Sort(result.RelatedURLs) // Ensure related URLs are sorted for consistency
	// }
	return &result, nil
}

// ssoReadbackRetries and ssoReadbackBackoff bound the re-fetch performed by
// GetApplicationAwaitSSO. They are package vars so tests can shorten the wait.
var (
	ssoReadbackRetries = 3
	ssoReadbackBackoff = 300 * time.Millisecond
)

// GetApplicationAwaitSSO fetches an application and, when the response omits the
// SSO object, re-fetches a bounded number of times until it appears. The GET
// that immediately follows a create/update — and the GET issued by
// `terraform import`, which reads in that same window — can transiently omit the
// SSO object due to backend eventual consistency; without this, Read/Import
// would record a configured SSO as null. Applications that genuinely have no SSO
// exhaust the small budget and return unchanged with an empty SSO.
func (c *APIClient) GetApplicationAwaitSSO(ctx context.Context, id string) (*Application, error) {
	app, err := c.GetApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	for attempt := 0; len(app.SSO) == 0 && attempt < ssoReadbackRetries; attempt++ {
		select {
		case <-time.After(ssoReadbackBackoff):
		case <-ctx.Done():
			return app, ctx.Err()
		}
		refetched, rerr := c.GetApplication(ctx, id)
		if rerr != nil {
			// Propagate the error: inside this loop app.SSO is always empty, so
			// returning it as success would persist sso=null and defeat the guard.
			// Failing lets Read preserve prior state and retry.
			return nil, rerr
		}
		app = refetched
	}
	return app, nil
}

// CreateApplication creates a new application
func (c *APIClient) CreateApplication(ctx context.Context, app *Application) (*Application, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateApplication calling API", map[string]any{
		"method": "POST",
		"path":   "/applications",
		"app":    app.Name,
	})

	resp, err := c.makeRequest(ctx, "POST", "/applications", app)
	if err != nil {
		return nil, err
	}

	var result Application
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateApplication updates an existing application
func (c *APIClient) UpdateApplication(ctx context.Context, id string, app *Application) error {
	path := fmt.Sprintf("/applications/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UpdateApplication calling API", map[string]any{
		"method": "PUT",
		"path":   path,
		"app":    app.Name,
	})

	resp, err := c.makeRequest(ctx, "PUT", path, app)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// CompleteApplication completes an application (transitions state from incomplete to complete)
func (c *APIClient) CompleteApplication(ctx context.Context, id string) error {
	path := fmt.Sprintf("/applications/%s?action=complete", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CompleteApplication calling API", map[string]any{
		"method": "PUT",
		"path":   path,
		"app_id": id,
	})

	resp, err := c.makeRequest(ctx, "PUT", path, map[string]interface{}{})
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// DeleteApplication deletes an application
func (c *APIClient) DeleteApplication(ctx context.Context, id string) error {
	path := fmt.Sprintf("/applications/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteApplication calling API", map[string]any{
		"method": "DELETE",
		"path":   path,
		"app_id": id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Application already deleted (404), treating as success", map[string]any{
			"app_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// Access Policy API methods

// AccessPolicy represents an access policy
type AccessPolicy struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"` // Required field for create/update
	// Priority is a pointer with omitempty so an omitted priority is absent from
	// the request body and the backend assigns one; an explicit value (including
	// 0) is sent verbatim.
	Priority    *int         `json:"priority,omitempty"`
	Modified    string       `json:"modified,omitempty"`
	Apps        []string     `json:"apps,omitempty"`
	AccessRules []AccessRule `json:"accessRules,omitempty"`
}

// AccessRule represents an access rule within an access policy
type AccessRule struct {
	ID               string            `json:"id,omitempty"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description"` // No omitempty - schema defaults to "" and must be sent to clear
	Priority         int               `json:"priority"`
	Active           bool              `json:"active"`                 // Required field - no omitempty
	Access           string            `json:"access,omitempty"`       // ACCESS_DENY, ACCESS_ALLOW
	AccessNative     string            `json:"accessNative,omitempty"` // ACCESS_DENY, ACCESS_ALLOW
	AdvancedSettings *AdvancedSettings `json:"advancedSettings,omitempty"`
	Conditions       []Condition       `json:"conditions,omitempty"`
	Restrictions     *Restrictions     `json:"restrictions,omitempty"`
	Rules            []Rule            `json:"rules,omitempty"`
}

// AdvancedSettings represents advanced settings for access rules
type AdvancedSettings struct {
	DomainOverrides []DomainOverride `json:"domainOverrides,omitempty"`
}

// DomainOverride represents a domain override setting
type DomainOverride struct {
	FQDN        string   `json:"fqdn"`
	LocationIDs []string `json:"locationIds"`
	Type        string   `json:"type"`
}

// Condition represents a condition for access rules.
//
// User and group scope belongs in a Rule with Type "TYPE_USERGROUP". The
// service's retired conditions[].userAndGroups field is neither sent nor
// decoded; encoding/json ignores it in responses.
type Condition struct {
	PlatformFilter string `json:"platformFilter,omitempty"` // PLATFORM_FILTER_MOBILE, PLATFORM_FILTER_PC, PLATFORM_FILTER_ANY
}

// Restrictions represents access rule restrictions
type Restrictions struct {
	// No omitempty: false is the Go zero value and would be dropped, leaving the
	// backend's previous value and causing "inconsistent result after apply".
	RedirectSBS              bool                   `json:"redirectSBS"`
	EnhancedSecuritySettings map[string]interface{} `json:"enhancedSecuritySettings,omitempty"`
}

// Rule represents a rule within an access rule
type Rule struct {
	Type      string                 `json:"type,omitempty"`     // TYPE_TAG, TYPE_USERGROUP, TYPE_PLATFORM, TYPE_MACHINEGROUP, TYPE_MULTIURLDOMAIN
	Operator  string                 `json:"operator,omitempty"` // OPERATOR_EQ, OPERATOR_IN, OPERATOR_CONTAINS, OPERATOR_LTE, OPERATOR_GTE, OPERATOR_NOT, OPERATOR_RANGE
	TagSource string                 `json:"tagSource"`          // NLS, CAS, EPA, ITM, ThirdPartyDevicePosture, CONTEXTUAL
	TagKey    string                 `json:"tagKey"`
	Values    []string               `json:"values,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AccessPoliciesResponse represents the response from listing access policies
type AccessPoliciesResponse struct {
	Policies []AccessPolicy `json:"items,omitempty"`
	Total    int            `json:"total,omitempty"`
	Count    int            `json:"count,omitempty"`
	Offset   int            `json:"offset,omitempty"`
}

// GetAccessPolicies retrieves a list of access policies
func (c *APIClient) GetAccessPolicies(ctx context.Context, offset, limit int, name, orderBy string) (*AccessPoliciesResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if orderBy != "" {
		params.Add("orderby", orderBy)
	} else {
		params.Add("orderby", "name")
	}
	if name != "" {
		params.Add("name", name)
	}

	path := "/accessPolicy"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetAccessPolicies calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result AccessPoliciesResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	// Debug logging after JSON unmarshaling
	tflog.Debug(ctx, "spa-terraform-provider: GetAccessPolicies unmarshaled result", map[string]any{
		"total_policies": len(result.Policies),
	})

	// for i, policy := range result.Policies {
	// 	slices.Sort(policy.Apps) // Ensure apps are sorted for consistency
	// 	tflog.Debug(ctx, "spa-terraform-provider: Policy after unmarshaling", map[string]any{
	// 		"index":        i,
	// 		"policy_id":    policy.ID,
	// 		"policy_name":  policy.Name,
	// 		"access_rules": len(policy.AccessRules),
	// 	})
	// }

	return &result, nil
}

// GetAccessPoliciesDetailed retrieves access policies with optional detailed information
// When detailed=true OR when FetchDetailsOnList is enabled, it fetches full policy details
// This reduces the number of API calls and avoids rate limiting issues by reusing the same auth token
func (c *APIClient) GetAccessPoliciesDetailed(ctx context.Context, offset, limit int, name, orderBy string, detailed bool) (*AccessPoliciesResponse, error) {
	// Check if we should fetch details based on the flag or explicit request
	shouldFetchDetails := detailed || c.FetchDetailsOnList

	if !shouldFetchDetails {
		// Use standard listing
		return c.GetAccessPolicies(ctx, offset, limit, name, orderBy)
	}

	tflog.Info(ctx, "spa-terraform-provider: Fetching detailed access policies list", map[string]interface{}{
		"offset":                offset,
		"limit":                 limit,
		"name":                  name,
		"orderBy":               orderBy,
		"detailed":              detailed,
		"fetch_details_on_list": c.FetchDetailsOnList,
		"should_fetch_details":  shouldFetchDetails,
	})

	// First get the list of access policies
	policies, err := c.GetAccessPolicies(ctx, offset, limit, name, orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get access policies list: %w", err)
	}

	// Convert basic policy list to detailed policies
	// We'll fetch each policy's full details in batch
	detailedPolicies := make([]AccessPolicy, len(policies.Policies))

	// Use a channel to limit concurrent requests and avoid overwhelming the API
	const maxConcurrent = 5
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fetchErrors []error

	for i, policy := range policies.Policies {
		wg.Add(1)
		go func(index int, policyItem AccessPolicy) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Get detailed policy info using the same client
			fullPolicy, err := c.GetAccessPolicy(ctx, policyItem.ID)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, fmt.Errorf("failed to get details for policy %s (%s): %w", policyItem.Name, policyItem.ID, err))
				mu.Unlock()
				// Use the basic info we already have
				detailedPolicies[index] = policyItem
				return
			}

			// Use the full policy details
			mu.Lock()
			detailedPolicies[index] = *fullPolicy
			mu.Unlock()
		}(i, policy)
	}

	wg.Wait()

	// Log any fetch errors but don't fail the entire operation
	if len(fetchErrors) > 0 {
		tflog.Warn(ctx, "Some access policy details could not be fetched", map[string]interface{}{
			"error_count": len(fetchErrors),
			"errors":      fmt.Sprintf("%v", fetchErrors),
		})
	}

	// Return the response with detailed policies
	return &AccessPoliciesResponse{
		Policies: detailedPolicies,
		Total:    policies.Total,
		Count:    policies.Count,
		Offset:   policies.Offset,
	}, nil
}

// GetAccessPolicy retrieves a specific access policy by ID
func (c *APIClient) GetAccessPolicy(ctx context.Context, id string) (*AccessPolicy, error) {
	path := fmt.Sprintf("/accessPolicy/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetAccessPolicy calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result AccessPolicy
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	// slices.Sort(result.Apps)
	return &result, nil
}

// CreateAccessPolicy creates a new access policy
func (c *APIClient) CreateAccessPolicy(ctx context.Context, policy *AccessPolicy) (*AccessPolicy, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateAccessPolicy calling API", map[string]any{
		"method": "POST",
		"path":   "/accessPolicy",
		"policy": policy.Name,
	})

	// Body is logged (redacted) by makeRequest; don't add a raw payload dump here.
	resp, err := c.makeRequest(ctx, "POST", "/accessPolicy", policy)
	if err != nil {
		return nil, err
	}

	// Extract policy ID from Location header only if the request was successful (201 Created)
	// Format: /accessSecurity/accessPolicy/2edf72b2-90a0-4ccb-ac58-38f964694f70
	if resp.StatusCode == http.StatusCreated {
		location := resp.Header.Get("Location")
		if location != "" {
			// Extract ID from the last segment of the path
			parts := strings.Split(location, "/")
			if len(parts) > 0 {
				policy.ID = parts[len(parts)-1]
				tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateAccessPolicy extracted ID from Location header", map[string]any{
					"location": location,
					"id":       policy.ID,
				})
			}
		}
	}

	var result AccessPolicy
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	// If ID was extracted from Location header, set it on the result
	if policy.ID != "" {
		result.ID = policy.ID
	}
	// slices.Sort(result.Apps)

	return &result, nil
}

// UpdateAccessPolicy updates an existing access policy
func (c *APIClient) UpdateAccessPolicy(ctx context.Context, id string, policy *AccessPolicy) error {
	path := fmt.Sprintf("/accessPolicy/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UpdateAccessPolicy calling API", map[string]any{
		"method": "PUT",
		"path":   path,
		"policy": policy.Name,
	})
	resp, err := c.makeRequest(ctx, "PUT", path, policy)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// DeleteAccessPolicy deletes an access policy
func (c *APIClient) DeleteAccessPolicy(ctx context.Context, id string) error {
	path := fmt.Sprintf("/accessPolicy/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteAccessPolicy calling API", map[string]any{
		"method":    "DELETE",
		"path":      path,
		"policy_id": id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Access policy already deleted (404), treating as success", map[string]any{
			"policy_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// Security Group API methods

// SecurityGroup represents a security group
type SecurityGroup struct {
	ID             string                `json:"id,omitempty"`
	Name           string                `json:"name"`
	AppIds         []string              `json:"appIds"`
	System         ConfigurationSettings `json:"system"`
	UnpublishedApp ConfigurationSettings `json:"unpublishedApp"`
	Modified       int64                 `json:"modified,omitempty"`
}

// ConfigurationSettings represents the data in/out configuration
type ConfigurationSettings struct {
	DataIn  string `json:"dataIn"`
	DataOut string `json:"dataOut"`
}

// SecurityGroupsResponse represents the response from listing security groups
type SecurityGroupsResponse struct {
	SecurityGroups []SecurityGroup `json:"items,omitempty"`
	Total          int             `json:"total,omitempty"`
	Count          int             `json:"count,omitempty"`
	Offset         int             `json:"offset,omitempty"`
}

// GetSecurityGroups retrieves a list of security groups
func (c *APIClient) GetSecurityGroups(ctx context.Context, offset, limit int) (*SecurityGroupsResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	path := "/securityGroup"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetSecurityGroups calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result SecurityGroupsResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSecurityGroup retrieves a specific security group by ID
func (c *APIClient) GetSecurityGroup(ctx context.Context, id string) (*SecurityGroup, error) {
	path := fmt.Sprintf("/securityGroup/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetSecurityGroup calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result SecurityGroup
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateSecurityGroup creates a new security group
func (c *APIClient) CreateSecurityGroup(ctx context.Context, sg *SecurityGroup) (*SecurityGroup, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateSecurityGroup calling API", map[string]any{
		"method": "POST",
		"path":   "/securityGroup",
		"sg":     sg.Name,
	})
	resp, err := c.makeRequest(ctx, "POST", "/securityGroup", sg)
	if err != nil {
		return nil, err
	}

	var result SecurityGroup
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateSecurityGroup updates an existing security group
func (c *APIClient) UpdateSecurityGroup(ctx context.Context, id string, sg *SecurityGroup) error {
	path := fmt.Sprintf("/securityGroup/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UpdateSecurityGroup calling API", map[string]any{
		"method": "PUT",
		"path":   path,
		"sg":     sg.Name,
	})
	resp, err := c.makeRequest(ctx, "PUT", path, sg)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// DeleteSecurityGroup deletes a security group
func (c *APIClient) DeleteSecurityGroup(ctx context.Context, id string) error {
	path := fmt.Sprintf("/securityGroup/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteSecurityGroup calling API", map[string]any{
		"method": "DELETE",
		"path":   path,
		"sg_id":  id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Security group already deleted (404), treating as success", map[string]any{
			"sg_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// Routing Domain API methods

// RoutingDomain represents a routing domain
type RoutingDomain struct {
	FQDN        string   `json:"fqdn"`
	Type        string   `json:"type"`
	AppType     string   `json:"appType,omitempty"`
	Comment     string   `json:"comment"` // Remove omitempty since API requires it
	Flag        string   `json:"flag,omitempty"`
	Error       string   `json:"error,omitempty"` // Keep omitempty for computed field
	IP          bool     `json:"ip"`              // Remove omitempty since false (zero value) must be sent explicitly
	LocationIds []string `json:"locationIds"`     // Remove omitempty since API requires array
}

// RoutingDomainsResponse represents the response from listing routing domains
type RoutingDomainsResponse struct {
	RoutingDomains []RoutingDomain `json:"items,omitempty"`
	Total          int             `json:"totalNum,omitempty"`
	Count          int             `json:"count,omitempty"`
	Offset         int             `json:"offset,omitempty"`
}

// GetRoutingDomains retrieves a list of routing domains
func (c *APIClient) GetRoutingDomains(ctx context.Context, offset, limit int) (*RoutingDomainsResponse, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetRoutingDomains called directly", map[string]any{
		"offset": offset,
		"limit":  limit,
	})
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	path := "/routingDomains"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetRoutingDomains calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result RoutingDomainsResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *APIClient) encodeFQDN(ctx context.Context, fqdn string) string {
	newFQDN := url.PathEscape(fqdn)

	if !strings.EqualFold(newFQDN, fqdn) {
		tflog.Trace(ctx, "spa-terraform-provider: APIClient.encodeFQDN encoded FQDN", map[string]any{
			"original": fqdn,
			"encoded":  newFQDN,
		})
	}
	return newFQDN

}

// GetRoutingDomain retrieves a specific routing domain by FQDN
func (c *APIClient) GetRoutingDomain(ctx context.Context, fqdn string) (*RoutingDomain, error) {
	fqdn = c.encodeFQDN(ctx, fqdn)
	path := fmt.Sprintf("/routingDomains/%s", fqdn)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetRoutingDomain calling API", map[string]any{
		"method": "GET",
		"path":   path,
		"fqdn":   fqdn,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result RoutingDomain
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateRoutingDomain creates a new routing domain
func (c *APIClient) CreateRoutingDomain(ctx context.Context, rd *RoutingDomain) (*RoutingDomain, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateRoutingDomain calling API", map[string]any{
		"method": "POST",
		"path":   "/routingDomains",
		"fqdn":   rd.FQDN,
	})
	resp, err := c.makeRequest(ctx, "POST", "/routingDomains", rd)
	if err != nil {
		return nil, err
	}

	var result RoutingDomain
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateRoutingDomain updates an existing routing domain
func (c *APIClient) UpdateRoutingDomain(ctx context.Context, fqdn string, rd *RoutingDomain) error {
	fqdn = c.encodeFQDN(ctx, fqdn)
	path := fmt.Sprintf("/routingDomains/%s", fqdn)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UpdateRoutingDomain calling API", map[string]any{
		"method": "PUT",
		"path":   path,
		"fqdn":   fqdn,
	})
	resp, err := c.makeRequest(ctx, "PUT", path, rd)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// DeleteRoutingDomain deletes a routing domain
func (c *APIClient) DeleteRoutingDomain(ctx context.Context, fqdn string) error {
	fqdn = c.encodeFQDN(ctx, fqdn)
	path := fmt.Sprintf("/routingDomains/%s", fqdn)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteRoutingDomain calling API", map[string]any{
		"method": "DELETE",
		"path":   path,
		"fqdn":   fqdn,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Routing domain already deleted (404), treating as success", map[string]any{
			"fqdn": fqdn,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// Certificate API methods

// Certificate represents a certificate
type Certificate struct {
	ID                  string `json:"id,omitempty"`
	CertificateID       string `json:"certificateId,omitempty"`
	CertificateName     string `json:"certificateName"`
	Certificate         string `json:"certificate,omitempty"`
	CertificatePassword string `json:"certificatePassword,omitempty"`
	ApplicationID       string `json:"applicationId,omitempty"`
	Domain              string `json:"domain,omitempty"`
}

// CertificatesResponse represents the response from listing certificates
type CertificatesResponse struct {
	Certificates []Certificate `json:"items,omitempty"`
	Total        int           `json:"total,omitempty"`
	Count        int           `json:"count,omitempty"`
	Offset       int           `json:"offset,omitempty"`
}

// GetCertificates retrieves a list of certificates
func (c *APIClient) GetCertificates(ctx context.Context, offset, limit int) (*CertificatesResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	path := "/certificate"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetCertificates calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result CertificatesResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateCertificate creates a new certificate
func (c *APIClient) CreateCertificate(ctx context.Context, cert *Certificate) (*Certificate, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateCertificate calling API", map[string]any{
		"method": "POST",
		"path":   "/certificate",
		"cert":   cert.CertificateName,
	})
	resp, err := c.makeRequest(ctx, "POST", "/certificate", cert)
	if err != nil {
		return nil, err
	}

	var result Certificate
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteCertificate deletes a certificate
func (c *APIClient) DeleteCertificate(ctx context.Context, id string) error {
	path := fmt.Sprintf("/certificate/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteCertificate calling API", map[string]any{
		"method":  "DELETE",
		"path":    path,
		"cert_id": id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Certificate already deleted (404), treating as success", map[string]any{
			"cert_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// AssignCertificateToApplication assigns a certificate to an application domain
func (c *APIClient) AssignCertificateToApplication(ctx context.Context, applicationID, domain string, cert *Certificate) error {
	path := fmt.Sprintf("/certificate/application/%s/domain/%s", applicationID, domain)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.AssignCertificateToApplication calling API", map[string]any{
		"method":         "POST",
		"path":           path,
		"cert":           cert.CertificateName,
		"application_id": applicationID,
	})
	resp, err := c.makeRequest(ctx, "POST", path, cert)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// UnassignCertificateFromApplication removes a certificate from an application domain
func (c *APIClient) UnassignCertificateFromApplication(ctx context.Context, applicationID, domain string) error {
	path := fmt.Sprintf("/certificate/application/%s/domain/%s", applicationID, domain)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UnassignCertificateFromApplication calling API", map[string]any{
		"method":         "DELETE",
		"path":           path,
		"application_id": applicationID,
		"domain":         domain,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// Browser Mode API methods

// GetBrowserMode retrieves browser mode configuration
func (c *APIClient) GetBrowserMode(ctx context.Context) (*BrowserMode, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetBrowserMode called directly")

	resp, err := c.makeRequest(ctx, "GET", "/browserMode", nil)
	if err != nil {
		return nil, err
	}

	var result BrowserMode
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Hybrid Configuration API methods

// GetHybridConfig retrieves hybrid configuration
func (c *APIClient) GetHybridConfig(ctx context.Context) (*HybridConfig, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetHybridConfig called directly")
	resp, err := c.makeRequest(ctx, "GET", "/hybridConfig", nil)
	if err != nil {
		return nil, err
	}

	var result HybridConfig
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Last Activity API methods

// GetLastActivity retrieves last activity timestamp
func (c *APIClient) GetLastActivity(ctx context.Context) (*LastActivity, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetLastActivity called directly")
	resp, err := c.makeRequest(ctx, "GET", "/lastActivity", nil)
	if err != nil {
		return nil, err
	}

	var result LastActivity
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Terminate Machine Access API methods

// TerminateMachineAccessResponse represents the response for listing terminate machine access
type TerminateMachineAccessResponse struct {
	Items  []TerminateMachineAccess `json:"items,omitempty"`
	Total  int                      `json:"total,omitempty"`
	Count  int                      `json:"count,omitempty"`
	Offset int                      `json:"offset,omitempty"`
}

// TerminateUserAccessResponse represents the response for listing terminate user access
type TerminateUserAccessResponse struct {
	Items  []TerminateUserAccess `json:"items,omitempty"`
	Total  int                   `json:"total,omitempty"`
	Count  int                   `json:"count,omitempty"`
	Offset int                   `json:"offset,omitempty"`
}

// GetTerminateMachineAccess retrieves a list of machine access termination records
func (c *APIClient) GetTerminateMachineAccess(ctx context.Context, offset, limit int) (*TerminateMachineAccessResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	path := "/terminateAccess/machine"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetTerminateMachineAccess calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result TerminateMachineAccessResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTerminateUserAccess retrieves a list of user access termination records
func (c *APIClient) GetTerminateUserAccess(ctx context.Context, offset, limit int) (*TerminateUserAccessResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	path := "/terminateAccess/user"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetTerminateUserAccess calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result TerminateUserAccessResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTerminateMachineAccessByID retrieves a specific machine access termination record by ID
// Since the API doesn't support individual GET by ID, we fetch all records and find the matching one
func (c *APIClient) GetTerminateMachineAccessByID(ctx context.Context, id string) (*TerminateMachineAccess, error) {
	// Get all machine access termination records
	machines, err := c.GetTerminateMachineAccess(ctx, 0, -1)
	if err != nil {
		return nil, err
	}

	// Find the machine with matching ID
	for _, machine := range machines.Items {
		if machine.ID == id {
			return &machine, nil
		}
	}

	return nil, fmt.Errorf("terminate machine access with ID %s not found", id)
}

// CreateTerminateMachineAccess creates a new machine access termination record
func (c *APIClient) CreateTerminateMachineAccess(ctx context.Context, machine *TerminateMachineAccess) (*TerminateMachineAccess, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateTerminateMachineAccess calling API", map[string]any{
		"method":  "POST",
		"path":    "/terminateAccess/machine",
		"machine": machine.Name,
	})
	resp, err := c.makeRequest(ctx, "POST", "/terminateAccess/machine", machine)
	if err != nil {
		return nil, err
	}

	var result TerminateMachineAccess
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteTerminateMachineAccess deletes a machine access termination record
func (c *APIClient) DeleteTerminateMachineAccess(ctx context.Context, id string) error {
	path := fmt.Sprintf("/terminateAccess/machine/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteTerminateMachineAccess calling API", map[string]any{
		"method":     "DELETE",
		"path":       path,
		"machine_id": id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Terminate machine access already deleted (404), treating as success", map[string]any{
			"machine_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// Terminate User Access API methods

// GetTerminateUserAccessByID retrieves a specific user access termination record by ID
// Since the API doesn't support individual GET by ID, we fetch all records and find the matching one
func (c *APIClient) GetTerminateUserAccessByID(ctx context.Context, id string) (*TerminateUserAccess, error) {
	// Get all user access termination records
	users, err := c.GetTerminateUserAccess(ctx, 0, -1)
	if err != nil {
		return nil, err
	}

	// Find the user with matching ID
	for _, user := range users.Items {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("terminate user access with ID %s not found", id)
}

// CreateTerminateUserAccess creates a new user access termination record
func (c *APIClient) CreateTerminateUserAccess(ctx context.Context, user *TerminateUserAccess) (*TerminateUserAccess, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateTerminateUserAccess calling API", map[string]any{
		"method": "POST",
		"path":   "/terminateAccess/user",
		"user":   user.Email,
	})
	resp, err := c.makeRequest(ctx, "POST", "/terminateAccess/user", user)
	if err != nil {
		return nil, err
	}

	var result TerminateUserAccess
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateTerminateUserAccess updates an existing user access termination record
func (c *APIClient) UpdateTerminateUserAccess(ctx context.Context, id string, user *TerminateUserAccess) error {
	path := fmt.Sprintf("/terminateAccess/user/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UpdateTerminateUserAccess calling API", map[string]any{
		"method": "PUT",
		"path":   path,
		"user":   user.Email,
	})
	resp, err := c.makeRequest(ctx, "PUT", path, user)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// DeleteTerminateUserAccess deletes a user access termination record
func (c *APIClient) DeleteTerminateUserAccess(ctx context.Context, id string) error {
	path := fmt.Sprintf("/terminateAccess/user/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteTerminateUserAccess calling API", map[string]any{
		"method":  "DELETE",
		"path":    path,
		"user_id": id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Terminate user access already deleted (404), treating as success", map[string]any{
			"user_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// Session Policy API methods

// SessionPolicyCondition represents a condition within a session policy rule (spec §4.1)
type SessionPolicyCondition struct {
	Type      string                 `json:"type"`
	Operator  string                 `json:"operator"`
	TagSource string                 `json:"tagSource"`
	TagKey    string                 `json:"tagKey"`
	Values    []string               `json:"values"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SessionPolicyAction represents the actions block of a session policy rule (spec §4.2)
type SessionPolicyAction struct {
	Routing               string `json:"routing,omitempty"`
	DisableSecurityGroups string `json:"disableSecurityGroups,omitempty"`
	LocalLanAccess        string `json:"localLanAccess,omitempty"`
}

// SessionPolicyRule represents one rule within a session policy (spec §4.3)
type SessionPolicyRule struct {
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Priority    int                      `json:"priority"`
	Active      bool                     `json:"active"`
	Actions     SessionPolicyAction      `json:"actions"`
	Conditions  []SessionPolicyCondition `json:"conditions"`
}

// SessionPolicy represents a session policy for create/update/read (spec §4.4–4.6)
type SessionPolicy struct {
	ID           string              `json:"id,omitempty"`
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	Active       bool                `json:"active"`
	Priority     *int                `json:"priority,omitempty"`
	GenericRules []SessionPolicyRule `json:"genericRules"`
}

// SessionPoliciesResponse is the paginated list response (spec §4.7)
type SessionPoliciesResponse struct {
	Items    []SessionPolicy `json:"items"`
	TotalNum int             `json:"totalNum"`
}

// GetSessionPolicies retrieves a list of session policies
func (c *APIClient) GetSessionPolicies(ctx context.Context, offset, limit int, name, orderBy string) (*SessionPoliciesResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if orderBy != "" {
		params.Add("orderby", orderBy)
	} else {
		params.Add("orderby", "name")
	}
	if name != "" {
		params.Add("name", name)
	}

	path := "/sessionPolicy"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetSessionPolicies calling API", map[string]any{
		"method": "GET",
		"path":   path,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result SessionPoliciesResponse
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSessionPolicy retrieves a specific session policy by ID
func (c *APIClient) GetSessionPolicy(ctx context.Context, id string) (*SessionPolicy, error) {
	path := fmt.Sprintf("/sessionPolicy/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.GetSessionPolicy calling API", map[string]any{
		"method":    "GET",
		"path":      path,
		"policy_id": id,
	})
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result SessionPolicy
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateSessionPolicy creates a new session policy
func (c *APIClient) CreateSessionPolicy(ctx context.Context, policy *SessionPolicy) (*SessionPolicy, error) {
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateSessionPolicy calling API", map[string]any{
		"method": "POST",
		"path":   "/sessionPolicy",
		"policy": policy.Name,
	})

	// Body is logged (redacted) by makeRequest; don't add a raw payload dump here.
	resp, err := c.makeRequest(ctx, "POST", "/sessionPolicy", policy)
	if err != nil {
		return nil, err
	}

	// Extract policy ID from Location header (POST returns 201 with no body)
	// Format: /accessSecurity/sessionPolicy/<uuid>
	if resp.StatusCode == http.StatusCreated {
		location := resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("session policy created but Location header was missing or empty")
		}
		parts := strings.Split(location, "/")
		id := parts[len(parts)-1]
		if id == "" {
			return nil, fmt.Errorf("session policy created but could not extract ID from Location header: %s", location)
		}
		policy.ID = id
		tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateSessionPolicy extracted ID from Location header", map[string]any{
			"location": location,
			"id":       policy.ID,
		})
	}

	var result SessionPolicy
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	if policy.ID != "" {
		result.ID = policy.ID
	}

	return &result, nil
}

// UpdateSessionPolicy updates an existing session policy (full replacement)
func (c *APIClient) UpdateSessionPolicy(ctx context.Context, id string, policy *SessionPolicy) error {
	path := fmt.Sprintf("/sessionPolicy/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.UpdateSessionPolicy calling API", map[string]any{
		"method":    "PUT",
		"path":      path,
		"policy_id": id,
	})
	resp, err := c.makeRequest(ctx, "PUT", path, policy)
	if err != nil {
		return err
	}

	return c.handleResponse(ctx, resp, nil)
}

// DeleteSessionPolicy deletes a session policy
func (c *APIClient) DeleteSessionPolicy(ctx context.Context, id string) error {
	path := fmt.Sprintf("/sessionPolicy/%s", id)

	tflog.Debug(ctx, "spa-terraform-provider: APIClient.DeleteSessionPolicy calling API", map[string]any{
		"method":    "DELETE",
		"path":      path,
		"policy_id": id,
	})
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	// Treat 404 as success — the resource is already gone (desired state)
	if resp.StatusCode == 404 {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		tflog.Info(ctx, "spa-terraform-provider: Session policy already deleted (404), treating as success", map[string]any{
			"policy_id": id,
		})
		return nil
	}

	return c.handleResponse(ctx, resp, nil)
}

// API Gateway Revision API methods

// GetAPIGatewayRevision retrieves API gateway revision information
func (c *APIClient) GetAPIGatewayRevision(ctx context.Context) (map[string]any, error) {
	resp, err := c.makeRequest(ctx, "GET", "/CitrixAPIGatewayRevision", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := c.handleResponse(ctx, resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
