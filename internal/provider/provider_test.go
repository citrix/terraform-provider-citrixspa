package provider

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"spa": providerserver.NewProtocol6WithError(New("test")()),
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

	client := NewAPIClient(baseURL, customerID, authToken, nil, false, tp)
	return client, nil
}
