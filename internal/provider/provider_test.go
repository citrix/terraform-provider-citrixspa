package provider

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"citrixspa": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance test: TF_ACC not set")
	}

	if os.Getenv("CITRIX_CUSTOMER_ID") == "" {
		t.Fatal("CITRIX_CUSTOMER_ID must be set for acceptance tests")
	}
	if os.Getenv("CITRIX_CLIENT_ID") == "" && os.Getenv("CITRIX_AUTH_TOKEN") == "" {
		t.Fatal("Either CITRIX_CLIENT_ID/CITRIX_CLIENT_SECRET or CITRIX_AUTH_TOKEN must be set for acceptance tests")
	}
	if os.Getenv("CITRIX_CLIENT_ID") != "" && os.Getenv("CITRIX_CLIENT_SECRET") == "" {
		t.Fatal("CITRIX_CLIENT_SECRET must be set when using CITRIX_CLIENT_ID")
	}
}

// testAccCreateClient creates an API client from environment variables for
// verifying resource existence/destruction directly against the backend API.
func testAccCreateClient() (*APIClient, error) {
	baseURL := "https://api.cloud.com/accessSecurity"
	if v := os.Getenv("SPA_BASE_URL"); v != "" {
		baseURL = v
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	tokenURL := fmt.Sprintf("%s://%s", base.Scheme, base.Host)

	customerID := os.Getenv("CITRIX_CUSTOMER_ID")
	authToken := os.Getenv("CITRIX_AUTH_TOKEN")
	clientID := os.Getenv("CITRIX_CLIENT_ID")
	clientSecret := os.Getenv("CITRIX_CLIENT_SECRET")

	var tp TokenProvider
	if clientID != "" && clientSecret != "" {
		tp = NewAuthenticatedClient(tokenURL, customerID, clientID, clientSecret, false)
	}

	client := NewAPIClient(baseURL, customerID, authToken, nil, 0, false, false, tp, "terraform-provider-citrixspa/test")
	return client, nil
}

// TestConfigure exercises the provider Configure method's validation and setup
// branches: base/token URL handling, customer-id and authentication validation
// (direct token vs service principal, missing, and conflicting), max_concurrent
// validation, rate-limit configuration, and the optional feature flags.
func TestConfigure(t *testing.T) {
	// Clear auth-related env vars so the test config is the sole source of truth
	// (Configure falls back to CITRIX_* env vars otherwise).
	t.Setenv("CITRIX_CUSTOMER_ID", "")
	t.Setenv("CITRIX_AUTH_TOKEN", "")
	t.Setenv("CITRIX_CLIENT_ID", "")
	t.Setenv("CITRIX_CLIENT_SECRET", "")

	ctx := context.Background()

	schemaResp := &provider.SchemaResponse{}
	New("test")().Schema(ctx, provider.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}
	configType := schemaResp.Schema.Type().TerraformType(ctx)

	str := func(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
	num := func(v int64) tftypes.Value { return tftypes.NewValue(tftypes.Number, big.NewFloat(float64(v))) }
	boolean := func(v bool) tftypes.Value { return tftypes.NewValue(tftypes.Bool, v) }

	// build produces a provider config; every attribute defaults to null and is
	// overridden by the supplied map.
	build := func(overrides map[string]tftypes.Value) tfsdk.Config {
		attrs := map[string]tftypes.Value{
			"base_url":                   tftypes.NewValue(tftypes.String, nil),
			"token_url":                  tftypes.NewValue(tftypes.String, nil),
			"customer_id":                tftypes.NewValue(tftypes.String, nil),
			"auth_token":                 tftypes.NewValue(tftypes.String, nil),
			"client_id":                  tftypes.NewValue(tftypes.String, nil),
			"client_secret":              tftypes.NewValue(tftypes.String, nil),
			"rate_limit":                 tftypes.NewValue(tftypes.Number, nil),
			"max_concurrent":             tftypes.NewValue(tftypes.Number, nil),
			"fetch_details_on_list":      tftypes.NewValue(tftypes.Bool, nil),
			"enable_token_cache":         tftypes.NewValue(tftypes.Bool, nil),
			"suppress_asb_notifications": tftypes.NewValue(tftypes.Bool, nil),
		}
		for k, v := range overrides {
			attrs[k] = v
		}
		return tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(configType, attrs)}
	}

	tests := []struct {
		name    string
		cfg     map[string]tftypes.Value
		wantErr string // substring of an expected error Summary; "" means expect success
	}{
		{
			name: "direct auth with defaults",
			cfg:  map[string]tftypes.Value{"customer_id": str("cust"), "auth_token": str("tok")},
		},
		{
			name: "service principal auth",
			cfg:  map[string]tftypes.Value{"customer_id": str("cust"), "client_id": str("cid"), "client_secret": str("sec")},
		},
		{
			name:    "missing customer id",
			cfg:     map[string]tftypes.Value{"auth_token": str("tok")},
			wantErr: "Unable to find customer ID",
		},
		{
			name:    "missing authentication",
			cfg:     map[string]tftypes.Value{"customer_id": str("cust")},
			wantErr: "Unable to find authentication credentials",
		},
		{
			name: "conflicting authentication methods",
			cfg: map[string]tftypes.Value{
				"customer_id": str("cust"), "auth_token": str("tok"),
				"client_id": str("cid"), "client_secret": str("sec"),
			},
			wantErr: "Conflicting authentication methods",
		},
		{
			name:    "invalid base url",
			cfg:     map[string]tftypes.Value{"base_url": str("https://exa\x7fmple.com")},
			wantErr: "Invalid Base URL",
		},
		{
			name:    "invalid token url",
			cfg:     map[string]tftypes.Value{"customer_id": str("cust"), "auth_token": str("tok"), "token_url": str("https://exa\x7fmple.com")},
			wantErr: "Invalid Token URL",
		},
		{
			name: "custom base and token url",
			cfg: map[string]tftypes.Value{
				"customer_id": str("cust"), "auth_token": str("tok"),
				"base_url":  str("https://dev.example.com/accessSecurity"),
				"token_url": str("https://dev.example.com"),
			},
		},
		{
			name:    "negative max_concurrent",
			cfg:     map[string]tftypes.Value{"customer_id": str("cust"), "auth_token": str("tok"), "max_concurrent": num(-1)},
			wantErr: "Invalid max_concurrent value",
		},
		{
			name: "zero max_concurrent (unlimited)",
			cfg:  map[string]tftypes.Value{"customer_id": str("cust"), "auth_token": str("tok"), "max_concurrent": num(0)},
		},
		{
			name: "positive max_concurrent",
			cfg:  map[string]tftypes.Value{"customer_id": str("cust"), "auth_token": str("tok"), "max_concurrent": num(3)},
		},
		{
			name: "rate limit and optional flags set",
			cfg: map[string]tftypes.Value{
				"customer_id": str("cust"), "auth_token": str("tok"),
				"rate_limit": num(5), "max_concurrent": num(3),
				"enable_token_cache":         boolean(false),
				"fetch_details_on_list":      boolean(true),
				"suppress_asb_notifications": boolean(true),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New("test")()
			resp := &provider.ConfigureResponse{}
			p.Configure(ctx, provider.ConfigureRequest{Config: build(tc.cfg)}, resp)

			if tc.wantErr == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("expected no error, got: %v", resp.Diagnostics)
				}
				return
			}
			found := false
			for _, d := range resp.Diagnostics.Errors() {
				if strings.Contains(d.Summary(), tc.wantErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, resp.Diagnostics)
			}
		})
	}
}
