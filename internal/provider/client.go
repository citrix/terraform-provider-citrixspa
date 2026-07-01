package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	Limiter                  *rate.Limiter // Rate limiter for API requests
	tokenProvider            TokenProvider // Token provider for getting auth tokens
	FetchDetailsOnList       bool          // When true, detailed listing methods will fetch individual item details
	SuppressASBNotifications bool          // When true, suppress ASB notifications on API requests
	UserAgent                string        // Custom User-Agent header for API requests
	Metrics                  *PerformanceMetrics
}

// Ensure APIClient implements SPAClient
var _ SPAClient = (*APIClient)(nil)

func NewAPIClient(baseURL, customerID, authToken string, limiter *rate.Limiter, fetchDetailsOnList bool, suppressASBNotifications bool, tp TokenProvider, userAgent string) *APIClient {
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

// makeRequest performs an HTTP request with proper headers and error handling
func (c *APIClient) makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
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

	// Log the request for debugging
	tflog.Debug(ctx, "spa-terraform-provider: SPA API request", map[string]any{
		"method":         method,
		"url":            fullURL,
		"transaction_id": transactionID,
		"body":           bodyContent,
		// "headers":        req.Header,
	})

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

		// Handle 429 Too Many Requests - drain and close the body so the transport can reuse the connection.
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
		return fmt.Errorf("API request failed with status %d (transaction ID: %s): %s", resp.StatusCode, txid, string(body))
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
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Active      bool         `json:"active"` // Required field for create/update
	Priority    int          `json:"priority"`
	Modified    string       `json:"modified,omitempty"`
	Apps        []string     `json:"apps,omitempty"`
	AccessRules []AccessRule `json:"accessRules,omitempty"`
}

// AccessRule represents an access rule within an access policy
type AccessRule struct {
	ID               string            `json:"id,omitempty"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
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

// Condition represents a condition for access rules
type Condition struct {
	PlatformFilter string                 `json:"platformFilter,omitempty"` // PLATFORM_FILTER_MOBILE, PLATFORM_FILTER_PC, PLATFORM_FILTER_ANY
	UserAndGroups  map[string]interface{} `json:"userAndGroups,omitempty"`
}

// Restrictions represents access rule restrictions
type Restrictions struct {
	RedirectSBS              bool                   `json:"redirectSBS,omitempty"`
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

	// Debug: marshal policy to JSON to see what's being sent
	policyJSON, _ := json.MarshalIndent(policy, "", "  ")
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateAccessPolicy payload", map[string]any{
		"json": string(policyJSON),
	})

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

	policyJSON, _ := json.MarshalIndent(policy, "", "  ")
	tflog.Debug(ctx, "spa-terraform-provider: APIClient.CreateSessionPolicy payload", map[string]any{
		"json": string(policyJSON),
	})

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
