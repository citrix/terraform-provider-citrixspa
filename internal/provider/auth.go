package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// OAuth2TokenResponse represents the response from the OAuth2 token endpoint
type OAuth2TokenResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// AuthClient handles authentication with Citrix Cloud
type AuthClient struct {
	BaseURL    string
	CustomerID string
	HTTPClient *http.Client
}

// GetBearerToken obtains a bearer token using OAuth 2.0 Client Credentials Grant
func (a *AuthClient) GetBearerToken(ctx context.Context, clientID, clientSecret string) (*OAuth2TokenResponse, error) {
	// Construct the token endpoint URL
	tokenURL := fmt.Sprintf("%s/cctrustoauth2/%s/tokens/clients", a.BaseURL, a.CustomerID)

	tflog.Info(ctx, "spa-terraform-provider: Request bearer token from CC", map[string]interface{}{
		"token_url": tokenURL,
		"client_id": clientID,
	})

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain token: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var tokenResponse OAuth2TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResponse, nil
}

// TokenCache represents a cached token with expiration
type TokenCache struct {
	Token     string
	ExpiresAt time.Time
}

// AuthenticatedClient wraps APIClient with automatic token management
type AuthenticatedClient struct {
	AuthClient       *AuthClient
	ClientID         string
	ClientSecret     string
	CachedToken      string
	TokenCache       *TokenCache
	TokenPersistence *TokenPersistence
	EnableTokenCache bool
	mu               sync.Mutex // Protects token cache access
}

// Ensure AuthenticatedClient implements SPAClient
var _ TokenProvider = (*AuthenticatedClient)(nil)

func (p *TokenCache) IsValid() bool {
	// Check if the token is still valid (not expired)
	return p != nil && time.Now().Before(p.ExpiresAt)
}

// NewAuthenticatedClient creates a new authenticated client
func NewAuthenticatedClient(authBaseURL, customerID, clientID, clientSecret string, enableTokenCache bool) *AuthenticatedClient {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	var tokenPersistence *TokenPersistence
	if enableTokenCache {
		tokenPersistence = NewTokenPersistence(customerID, clientID)
	}

	p := &AuthenticatedClient{
		AuthClient: &AuthClient{
			BaseURL:    authBaseURL,
			CustomerID: customerID,
			HTTPClient: httpClient,
		},
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		TokenPersistence: tokenPersistence,
		EnableTokenCache: enableTokenCache,
	}
	return p
}

func (ac *AuthenticatedClient) GetToken(ctx context.Context) (string, error) {
	if err := ac.EnsureValidToken(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure valid token: %w", err)
	}

	ac.mu.Lock()
	token := ac.CachedToken
	ac.mu.Unlock()

	if token == "" {
		return "", fmt.Errorf("no valid token available")
	}
	return token, nil
}

func (ac *AuthenticatedClient) EnsureValidToken(ctx context.Context) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// First check in-memory cache
	if ac.TokenCache != nil && ac.TokenCache.IsValid() {
		ac.CachedToken = ac.TokenCache.Token
		tflog.Debug(ctx, "spa-terraform-provider: Using cached token", map[string]interface{}{
			"expires_at": ac.TokenCache.ExpiresAt.Format(time.RFC3339),
		})
		return nil
	}

	// If token cache is enabled, try to load from disk
	if ac.EnableTokenCache && ac.TokenPersistence != nil {
		if cachedToken, err := ac.TokenPersistence.LoadToken(ac.AuthClient.CustomerID, ac.ClientID); err == nil && cachedToken != nil {
			tflog.Info(ctx, "spa-terraform-provider: Loaded valid token from disk cache")
			// Update in-memory cache
			ac.TokenCache = &TokenCache{
				Token:     cachedToken.Token,
				ExpiresAt: cachedToken.ExpiresAt,
			}
			ac.CachedToken = cachedToken.Token
			return nil
		}
	}

	// Get a new token
	token, err := ac.AuthClient.GetBearerToken(ctx, ac.ClientID, ac.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to get bearer token: %w", err)
	}

	expiresIn, err := time.ParseDuration(token.ExpiresIn + "s")
	if err != nil {
		expiresIn = time.Duration(3600 * time.Second) // Default to 1 hour if parsing fails
	}

	expiresAt := time.Now().Add(expiresIn).Add(-5 * time.Minute) // 5 minutes buffer

	tflog.Info(ctx, "spa-terraform-provider: Obtained new bearer token from CC", map[string]interface{}{
		"token_type": token.TokenType,
		"expires_in": token.ExpiresIn,
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	// Cache the token in memory
	ac.TokenCache = &TokenCache{
		Token:     token.AccessToken,
		ExpiresAt: expiresAt,
	}

	// Save to disk if enabled
	if ac.EnableTokenCache && ac.TokenPersistence != nil {
		if err := ac.TokenPersistence.SaveToken(ac.AuthClient.CustomerID, ac.ClientID, token.AccessToken, expiresAt); err != nil {
			tflog.Warn(ctx, "Failed to save token to disk cache", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			tflog.Debug(ctx, "spa-terraform-provider: Token saved to disk cache")
		}
	}

	// Update the API client
	ac.CachedToken = token.AccessToken

	return nil
}
