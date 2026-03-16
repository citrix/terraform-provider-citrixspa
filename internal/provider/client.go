package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/time/rate"
)

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
	GetSecurityGroups(ctx context.Context, offset, limit int, name string) (*SecurityGroupsResponse, error)
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
}

type TokenProvider interface {
	// getToken returns a valid auth token for the API client
	GetToken(ctx context.Context) (string, error)
}

// APIClient is a client for the SPA API
type APIClient struct {
	BaseURL            string
	CustomerID         string
	AuthToken          string
	HTTPClient         *http.Client
	Limiter            *rate.Limiter // Rate limiter for API requests
	tokenProvider      TokenProvider // Token provider for getting auth tokens
	FetchDetailsOnList bool          // When true, detailed listing methods will fetch individual item details
}

// Ensure APIClient implements SPAClient
var _ SPAClient = (*APIClient)(nil)

func NewAPIClient(baseURL, customerID, authToken string, limiter *rate.Limiter, fetchDetailsOnList bool, tp TokenProvider) *APIClient {
	p := &APIClient{
		BaseURL:    strings.TrimSuffix(baseURL, "/"), // Ensure no trailing slash
		CustomerID: customerID,
		AuthToken:  authToken,
		HTTPClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		Limiter:            limiter,
		FetchDetailsOnList: fetchDetailsOnList, // Set the flag for detailed listing
		tokenProvider:      tp,                 // Set the token provider for dynamic token management
	}

	if p.tokenProvider == nil {
		p.tokenProvider = p // Fallback to self if no provider is set
	}
	return p
}

func (c *APIClient) GetToken(ctx context.Context) (string, error) {
	return c.AuthToken, nil
}

// makeRequest performs an HTTP request with proper headers and error handling
func (c *APIClient) makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	// Get valid token before making the request
	token, err := c.tokenProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	var reqBody io.Reader
	var bodyContent string
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyBytes)
		bodyContent = string(bodyBytes)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
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

	// Rate limit the request if a limiter is configured
	if c.Limiter != nil {
		if err := c.Limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit exceeded (transaction ID: %s): %w", transactionID, err)
		}
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request (transaction ID: %s): %w", transactionID, err)
	}

	return resp, nil
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
	// CreatedTime          string         `json:"createdTime,omitempty"`
	State       string `json:"state,omitempty"`
	PolicyCount string `json:"policyCount,omitempty"`
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
	// CreatedTime          string         `json:"createdTime,omitempty"`
	State       string `json:"state,omitempty"`
	PolicyCount string `json:"policyCount,omitempty"`
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

	return c.handleResponse(ctx, resp, nil)
}

// Access Policy API methods

// AccessPolicy represents an access policy
type AccessPolicy struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Active      bool         `json:"active"` // Required field for create/update
	Priority    int          `json:"priority,omitempty"`
	Apps        []string     `json:"apps,omitempty"`
	AccessRules []AccessRule `json:"accessRules,omitempty"`
}

// AccessRule represents an access rule within an access policy
type AccessRule struct {
	ID               string            `json:"id,omitempty"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
	Priority         int               `json:"priority,omitempty"`
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
func (c *APIClient) GetSecurityGroups(ctx context.Context, offset, limit int, name string) (*SecurityGroupsResponse, error) {
	params := url.Values{}
	params.Add("offset", fmt.Sprintf("%d", offset))
	// Only add limit parameter if it's not negative (negative means no limit)
	if limit >= 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if name != "" {
		params.Add("name", name)
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
	IP          bool     `json:"ip,omitempty"`
	LocationIds []string `json:"locationIds"` // Remove omitempty since API requires array
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
