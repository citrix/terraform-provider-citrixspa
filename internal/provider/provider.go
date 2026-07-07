package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/time/rate"
)

// Ensure SPAProvider satisfies various provider interfaces.
var _ provider.Provider = &SPAProvider{}

// SPA service has a rate limit of 1000 requests per minute, which is approximately 16.67 requests per second.
// To avoid hitting the rate limit, we set a maximum of 15 requests per second. (900 requests per minute)
var maxRateLimitRequestPerSecond = 15 // Maximum rate limit requests per second for the provider.

// SPAProvider defines the provider implementation.
type SPAProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	limiter *rate.Limiter
}

// SPAProviderModel describes the provider data model.
type SPAProviderModel struct {
	BaseURL                  types.String `tfsdk:"base_url"`
	TokenURL                 types.String `tfsdk:"token_url"`
	CustomerID               types.String `tfsdk:"customer_id"`
	AuthToken                types.String `tfsdk:"auth_token"`
	ClientID                 types.String `tfsdk:"client_id"`
	ClientSecret             types.String `tfsdk:"client_secret"`
	RateLimit                types.Int64  `tfsdk:"rate_limit"` // Rate limit in requests per second
	FetchDetailsOnList       types.Bool   `tfsdk:"fetch_details_on_list"`
	EnableTokenCache         types.Bool   `tfsdk:"enable_token_cache"`
	SuppressASBNotifications types.Bool   `tfsdk:"suppress_asb_notifications"`
}

func (p *SPAProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: Metadata - Setting provider metadata")
	resp.TypeName = "spa"
	resp.Version = p.version
}

func (p *SPAProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: Schema - Defining provider schema")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for the SPA API. Should include the /accessSecurity path. Default: https://api.cloud.com/accessSecurity",
				Optional:            true,
			},
			"token_url": schema.StringAttribute{
				MarkdownDescription: "Token URL for the CC API. Default is the base URL scheme and host.",
				Optional:            true,
			},
			"customer_id": schema.StringAttribute{
				MarkdownDescription: "Citrix Cloud Customer ID",
				Optional:            true,
			},
			"auth_token": schema.StringAttribute{
				MarkdownDescription: "Citrix Cloud authorization token (CWSAuth Bearer format). Alternative to client_id/client_secret.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Citrix Cloud Service Principal Client ID. Used with client_secret for OAuth2 authentication.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Citrix Cloud Service Principal Client Secret. Used with client_id for OAuth2 authentication.",
				Optional:            true,
				Sensitive:           true,
			},
			"rate_limit": schema.Int64Attribute{
				MarkdownDescription: "Rate limit for the provider in requests per second. Default is 15 requests per second (900 requests per minute). " +
					"Set to 0 for no rate limiting, or a negative value for infinite rate limiting.",
				Optional: true,
			},
			"fetch_details_on_list": schema.BoolAttribute{
				MarkdownDescription: "When true, automatically fetch detailed information for each resource during list operations. " +
					"This makes additional API calls but reuses cached tokens to avoid rate limiting. Useful for tools like spa_manager.ps1 that need complete resource details.",
				Optional: true,
			},
			"enable_token_cache": schema.BoolAttribute{
				MarkdownDescription: "Enable encrypted disk-based token caching to reuse authentication tokens across terraform operations. " +
					"Tokens are encrypted using AES-256-GCM and stored in ~/.terraform.d/spa-cache/. Default: true.",
				Optional: true,
			},
			"suppress_asb_notifications": schema.BoolAttribute{
				MarkdownDescription: "When true, suppress ASB notifications during API operations. " +
					"Recommended when applying 10 or more resource changes to avoid intermittent errors. " +
					"Note: when enabled, synchronization of changes may be delayed. Default: false.",
				Optional: true,
			},
		},
	}
}

func (p *SPAProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: Configure - Configuring provider")
	var data SPAProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// Create and configure the client here.

	// Default values
	baseURL := "https://api.cloud.com/accessSecurity"
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Base URL",
			"The provider cannot create the SPA API client as the base URL is invalid: "+err.Error(),
		)
		return
	}

	// Set default token URL to base scheme and host
	tokenURL := fmt.Sprintf("%s://%s", base.Scheme, base.Host)

	tflog.Debug(ctx, "spa-terraform-provider: SPA Provider configuration data", map[string]any{
		"base_url":    baseURL,
		"token_url":   data.TokenURL.String(),
		"customer_id": data.CustomerID.String(),
	})

	if !data.TokenURL.IsNull() {
		// Use the provided token URL
		providedTokenURL := data.TokenURL.ValueString()
		_, err := url.Parse(providedTokenURL)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Token URL",
				"The provider cannot create the SPA API client as the token URL is invalid: "+err.Error(),
			)
			return
		}
		// Use the full provided token URL
		tokenURL = providedTokenURL
		tflog.Debug(ctx, "spa-terraform-provider: Using custom token endpoint", map[string]any{
			"token_endpoint": tokenURL,
		})
	} else {
		tflog.Debug(ctx, "spa-terraform-provider: Using default token endpoint", map[string]any{
			"token_endpoint": tokenURL,
		})
	}

	customerID := os.Getenv("CITRIX_CUSTOMER_ID")
	if !data.CustomerID.IsNull() {
		customerID = data.CustomerID.ValueString()
	}

	// Get authentication parameters
	authToken := os.Getenv("CITRIX_AUTH_TOKEN")
	if !data.AuthToken.IsNull() {
		authToken = data.AuthToken.ValueString()
	}

	clientID := os.Getenv("CITRIX_CLIENT_ID")
	if !data.ClientID.IsNull() {
		clientID = data.ClientID.ValueString()
	}

	clientSecret := os.Getenv("CITRIX_CLIENT_SECRET")
	if !data.ClientSecret.IsNull() {
		clientSecret = data.ClientSecret.ValueString()
	}

	// Validate required customer ID
	if customerID == "" {
		resp.Diagnostics.AddError(
			"Unable to find customer ID",
			"The provider cannot create the SPA API client as there is a missing or empty value for the customer ID. "+
				"Set the customer_id value in the configuration or use the CITRIX_CUSTOMER_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Validate authentication method
	hasDirectAuth := authToken != ""
	hasServicePrincipal := clientID != "" && clientSecret != ""

	if !hasDirectAuth && !hasServicePrincipal {
		resp.Diagnostics.AddError(
			"Unable to find authentication credentials",
			"The provider requires either:\n"+
				"1. A direct auth token (set auth_token in configuration or CITRIX_AUTH_TOKEN environment variable), OR\n"+
				"2. Service Principal credentials (set client_id and client_secret in configuration or CITRIX_CLIENT_ID and CITRIX_CLIENT_SECRET environment variables).",
		)
	}

	if hasDirectAuth && hasServicePrincipal {
		resp.Diagnostics.AddError(
			"Conflicting authentication methods",
			"The provider has both direct auth token and service principal credentials configured. "+
				"Please use only one authentication method.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.RateLimit.IsNull() && p.limiter != nil {
		rps := data.RateLimit.ValueInt64()
		r, b := calculateRateLimit(ctx, int(rps))
		p.limiter.SetLimit(r)
		p.limiter.SetBurst(b)
		tflog.Debug(ctx, "spa-terraform-provider: Rate limit configured", map[string]any{
			"rate_limit":            rps,
			"rate_limit_per_second": r,
			"bucket_size":           b,
		})
	}

	// Parse provider configuration flags
	enableTokenCache := true // Default to enabled
	if !data.EnableTokenCache.IsNull() {
		enableTokenCache = data.EnableTokenCache.ValueBool()
	}

	fetchDetailsOnList := false // Default to disabled
	if !data.FetchDetailsOnList.IsNull() {
		fetchDetailsOnList = data.FetchDetailsOnList.ValueBool()
	}

	suppressASBNotifications := false // Default to disabled
	if !data.SuppressASBNotifications.IsNull() {
		suppressASBNotifications = data.SuppressASBNotifications.ValueBool()
	}

	tflog.Debug(ctx, "spa-terraform-provider: Provider configuration", map[string]any{
		"enable_token_cache":         enableTokenCache,
		"fetch_details_on_list":      fetchDetailsOnList,
		"suppress_asb_notifications": suppressASBNotifications,
	})

	// Create the appropriate client based on authentication method
	var client SPAClient

	userAgent := fmt.Sprintf("terraform-provider-citrixspa/%s", p.version)

	var tp TokenProvider
	if !hasDirectAuth {
		// Use service principal authentication
		tp = NewAuthenticatedClient(tokenURL, customerID, clientID, clientSecret, enableTokenCache)
	}
	client = NewAPIClient(baseURL, customerID, authToken, p.limiter, fetchDetailsOnList, suppressASBNotifications, tp, userAgent)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SPAProvider) Resources(ctx context.Context) []func() resource.Resource {
	tflog.Debug(ctx, "spa-terraform-provider: Resources - Registering provider resources")
	return []func() resource.Resource{
		NewApplicationResource,
		NewAccessPolicyResource,
		NewSecurityGroupResource,
		NewRoutingDomainResource,
		NewCertificateResource,
		NewBrowserModeResource,
		NewTerminateMachineAccessResource,
		NewTerminateUserAccessResource,
		NewSessionPolicyResource,
	}
}

func (p *SPAProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	tflog.Debug(ctx, "spa-terraform-provider: DataSources - Registering provider data sources")
	return []func() datasource.DataSource{
		NewApplicationDataSource,
		NewApplicationsDataSource,
		NewAccessPolicyDataSource,
		NewAccessPoliciesDataSource,
		NewSecurityGroupDataSource,
		NewSecurityGroupsDataSource,
		NewRoutingDomainDataSource,
		NewRoutingDomainsDataSource,
		NewBrowserModeDataSource,
		NewHybridConfigDataSource,
		NewLastActivityDataSource,
		NewTerminateMachineAccessDataSource,
		NewTerminateUserAccessDataSource,
		NewCertificatesDataSource,
		NewSessionPolicyDataSource,
		NewSessionPoliciesDataSource,
	}
}

// calculateRateLimit calculates the rate limit and bucket size based on the provided requests per second (rps).
// If the rps is zero or exceeds the maximum allowed, it defaults to the maximum rate limit.
// If the rps is less than 0, it returns an infinite rate limit.
// Returns the rate.Limit and the bucket size.
func calculateRateLimit(ctx context.Context, rps int) (rate.Limit, int) {
	tflog.Debug(ctx, "spa-terraform-provider: calculateRateLimit - Calculating rate limit", map[string]any{
		"requested_rps": rps,
	})
	if rps < 0 {
		return rate.Inf, 0
	} else if rps > maxRateLimitRequestPerSecond {
		tflog.Warn(ctx, "Rate limit exceeds maximum allowed, using maximum rate limit", map[string]any{
			"max_rate_limit":       maxRateLimitRequestPerSecond,
			"requested_rate_limit": rps,
		})
		rps = maxRateLimitRequestPerSecond
	} else if rps == 0 {
		tflog.Warn(ctx, "Rate limit is zero, using maximum rate limit", map[string]any{
			"max_rate_limit":       maxRateLimitRequestPerSecond,
			"requested_rate_limit": rps,
		})
		rps = maxRateLimitRequestPerSecond
	}

	// Calculate the bucket size based on the rate limit
	return rate.Limit(rps), rps
}

func New(version string) func() provider.Provider {
	tflog.Debug(context.Background(), "spa-terraform-provider: New - Creating new provider instance", map[string]any{
		"version": version,
	})
	r, b := calculateRateLimit(context.Background(), maxRateLimitRequestPerSecond)
	return func() provider.Provider {
		return &SPAProvider{
			version: version,
			limiter: rate.NewLimiter(rate.Limit(r), b),
		}
	}
}
