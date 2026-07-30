package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

// tokenRetryBaseDelay is the base backoff for token-request retries (delay =
// tokenRetryBaseDelay << attempt: 1s, 2s, 4s ...). It is a package-level var so
// tests can shrink it to avoid real-time sleeps; production behavior is unchanged.
var tokenRetryBaseDelay = time.Second

// retryAfterWait resolves the delay before the next 429 retry. It honors a
// Retry-After header — either delta-seconds ("120") or an HTTP-date — and falls
// back to base when the header is absent or unparseable. The result is clamped
// to [0, max]: a negative Retry-After (e.g. "-1") or a past HTTP-date never
// yields a negative duration (which would fire the timer immediately), and an
// oversized value is capped at max. now is passed in so the logic is testable
// without real time.
func retryAfterWait(header string, now time.Time, base, max time.Duration) time.Duration {
	wait := base
	if header != "" {
		if secs, err := strconv.Atoi(header); err == nil {
			// Cap before multiplying: time.Duration(secs) * time.Second would
			// overflow int64 (and wrap negative) for values above ~9.2e9 seconds.
			if int64(secs) > int64(max/time.Second) {
				wait = max
			} else {
				wait = time.Duration(secs) * time.Second
			}
		} else if t, err := http.ParseTime(header); err == nil {
			wait = t.Sub(now)
		}
	}
	if wait < 0 {
		wait = 0
	}
	if wait > max {
		wait = max
	}
	return wait
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
	encoded := data.Encode()

	// The token endpoint is rate-limited independently of the data-plane API and
	// occasionally returns transient 429/5xx responses under load (e.g. parallel
	// acceptance tests). Retry those with backoff (honoring Retry-After on 429),
	// mirroring the data-plane retry loop, so a single transient blip doesn't fail
	// the whole operation. Non-retryable statuses (401/403/400) fail immediately.
	const maxRetries = 3
	const maxRetryWait = 30 * time.Second
	var lastStatus int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(encoded))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := a.HTTPClient.Do(req)

		var wait time.Duration
		switch {
		case err != nil:
			// Do may return a non-nil resp alongside err (e.g. redirect-policy
			// failures); drain/close it so the connection is not leaked.
			if resp != nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			// If the context was cancelled/expired while the request was in flight,
			// Do surfaces that error here — return it directly instead of logging a
			// misleading "network error, retrying" and computing a pointless backoff.
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to obtain token: %w", err)
			}
			wait = tokenRetryBaseDelay << attempt
			tflog.Warn(ctx, "spa-terraform-provider: token request failed (network error), retrying", map[string]interface{}{
				"error": err.Error(), "attempt": attempt + 1, "max_retries": maxRetries, "wait_seconds": wait.Seconds(),
			})

		case resp.StatusCode == http.StatusOK:
			var tokenResponse OAuth2TokenResponse
			decodeErr := json.NewDecoder(resp.Body).Decode(&tokenResponse)
			resp.Body.Close()
			if decodeErr != nil {
				return nil, fmt.Errorf("failed to parse token response: %w", decodeErr)
			}
			return &tokenResponse, nil

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			// Transient server-side condition: retry. For 429, honor Retry-After
			// (delta-seconds or HTTP-date), clamping negatives and capping at
			// maxRetryWait; see retryAfterWait.
			lastStatus = resp.StatusCode
			wait = tokenRetryBaseDelay << attempt
			if resp.StatusCode == http.StatusTooManyRequests {
				wait = retryAfterWait(resp.Header.Get("Retry-After"), time.Now(), wait, maxRetryWait)
			} else if wait > maxRetryWait {
				wait = maxRetryWait
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if attempt == maxRetries {
				return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
			}
			tflog.Warn(ctx, "spa-terraform-provider: token request rate-limited/transient error, retrying", map[string]interface{}{
				"status": resp.StatusCode, "attempt": attempt + 1, "max_retries": maxRetries, "wait_seconds": wait.Seconds(),
			})

		default:
			// Non-retryable status (e.g. 400/401/403): fail immediately.
			status := resp.StatusCode
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("token request failed with status %d", status)
		}

		// Wait before the next attempt, respecting context cancellation.
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	// Unreachable in practice: every switch arm returns on the final attempt
	// (attempt == maxRetries). This return exists only because a for loop is not a
	// terminating statement in Go, so the compiler requires one here.
	return nil, fmt.Errorf("token request failed with status %d", lastStatus)
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
