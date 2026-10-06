package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// CheckDestroy function — verify routing domains are deleted from backend after destroy
// =============================================================================

func testAccCheckRoutingDomainDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "citrixspa_routing_domain" {
			continue
		}

		fqdn := rs.Primary.Attributes["fqdn"]
		_, err := client.GetRoutingDomain(ctx, fqdn)
		if err == nil {
			return fmt.Errorf("routing domain %s still exists in the API after destroy", fqdn)
		}
		if !IsNotFound(err) {
			return fmt.Errorf("unexpected error checking routing domain %s: %s", fqdn, err)
		}
	}
	return nil
}

// =============================================================================
// Exists-in-API check function — verify routing domain exists in the backend
// =============================================================================

func testAccCheckRoutingDomainExistsInAPI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		fqdn := rs.Primary.Attributes["fqdn"]
		if fqdn == "" {
			return fmt.Errorf("no FQDN set for resource %s", resourceName)
		}

		client, err := testAccCreateClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		rd, err := client.GetRoutingDomain(context.Background(), fqdn)
		if err != nil {
			return fmt.Errorf("routing domain %s not found in API: %s", fqdn, err)
		}

		if rd.FQDN != fqdn {
			return fmt.Errorf("routing domain FQDN mismatch: API=%q, state=%q", rd.FQDN, fqdn)
		}

		return nil
	}
}

// =============================================================================
// Pre-test cleanup helper
// =============================================================================

// testAccCleanupRoutingDomain attempts to delete a routing domain that may be
// left over from a previous failed test run. It disables the domain first
// (required by the API), then deletes it. Errors are ignored — if the domain
// does not exist, this is a no-op.
func testAccCleanupRoutingDomain(fqdn string) {
	client, err := testAccCreateClient()
	if err != nil {
		return
	}
	ctx := context.Background()

	// Check if the routing domain exists
	rd, err := client.GetRoutingDomain(ctx, fqdn)
	if err != nil {
		// Domain doesn't exist or API error — nothing to clean up
		return
	}

	// If not already disabled, disable it first (API requires this before deletion)
	if rd.Flag != "disabled" {
		updateRD := &RoutingDomain{
			FQDN:        rd.FQDN,
			Type:        rd.Type,
			AppType:     rd.AppType,
			Comment:     rd.Comment,
			Flag:        "disabled",
			Error:       rd.Error,
			IP:          rd.IP,
			LocationIds: rd.LocationIds,
		}
		_ = client.UpdateRoutingDomain(ctx, fqdn, updateRD)
	}

	// Delete the routing domain
	_ = client.DeleteRoutingDomain(ctx, fqdn)
}

// =============================================================================
// Routing Domain Tests
// =============================================================================

// testAccRoutingDomainConfig generates a dynamic routing domain configuration
// with customizable parameters, reducing the need for multiple similar config functions.
func testAccRoutingDomainConfig(resourceName, fqdn, rdType, appType, comment, flag, ip, locationIds string) string {
	return fmt.Sprintf(`
resource "citrixspa_routing_domain" "%s" {
  fqdn         = %q
  type         = %q
  app_type     = %q
  comment      = %q
  flag         = %q
  ip           = %s
  location_ids = %s
}
`, resourceName, fqdn, rdType, appType, comment, flag, ip, locationIds)
}

func TestAccRoutingDomain_internal(t *testing.T) {
	fqdnOriginal := "tf-acc-test-internal.example.com"
	fqdnRenamed := "tf-acc-test-internal-renamed.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdnOriginal)
			testAccCleanupRoutingDomain(fqdnRenamed)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create the routing domain
			{
				Config: testAccRoutingDomainConfig("test_internal", fqdnOriginal, "internal", "web", "Terraform acceptance test - internal", "enabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "fqdn", fqdnOriginal),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "type", "internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "app_type", "web"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "comment", "Terraform acceptance test - internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "flag", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "location_ids.#", "0"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan
			{
				Config:   testAccRoutingDomainConfig("test_internal", fqdnOriginal, "internal", "web", "Terraform acceptance test - internal", "enabled", "false", "[]"),
				PlanOnly: true,
			},
			// Step 3: Update comment and flag (same FQDN)
			{
				Config: testAccRoutingDomainConfig("test_internal", fqdnOriginal, "internal", "web", "Terraform acceptance test - internal UPDATED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "fqdn", fqdnOriginal),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "type", "internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "app_type", "web"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "comment", "Terraform acceptance test - internal UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "flag", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "location_ids.#", "0"),
				),
			},
			// Step 4: Change the FQDN — exercises the oldFQDN-from-state path in Update
			{
				Config: testAccRoutingDomainConfig("test_internal", fqdnRenamed, "internal", "web", "Terraform acceptance test - internal RENAMED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "fqdn", fqdnRenamed),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "type", "internal"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "app_type", "web"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "comment", "Terraform acceptance test - internal RENAMED"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "flag", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_internal", "location_ids.#", "0"),
				),
			},
			{
				ResourceName:                         "citrixspa_routing_domain.test_internal",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fqdnRenamed,
				ImportStateVerifyIdentifierAttribute: "fqdn",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRoutingDomain_external(t *testing.T) {
	fqdn := "tf-acc-test-external.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainConfig("test_external", fqdn, "external", "saas", "Terraform acceptance test - external", "enabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_external"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "fqdn", fqdn),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "type", "external"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "app_type", "saas"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "comment", "Terraform acceptance test - external"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "flag", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "location_ids.#", "0"),
				),
			},
			{
				Config: testAccRoutingDomainConfig("test_external", fqdn, "external", "saas", "Terraform acceptance test - external UPDATED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_external"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "fqdn", fqdn),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "type", "external"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "app_type", "saas"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "comment", "Terraform acceptance test - external UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "flag", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_external", "location_ids.#", "0"),
				),
			},
			{
				ResourceName:                         "citrixspa_routing_domain.test_external",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fqdn,
				ImportStateVerifyIdentifierAttribute: "fqdn",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRoutingDomain_internalViaGateway(t *testing.T) {
	fqdn := "tf-acc-test-int-gateway.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainConfig("test_int_gateway", fqdn, "internal_via_gateway", "web", "Terraform acceptance test - internal via gateway", "enabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_int_gateway"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "fqdn", fqdn),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "type", "internal_via_gateway"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "app_type", "web"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "comment", "Terraform acceptance test - internal via gateway"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "flag", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "location_ids.#", "0"),
				),
			},
			{
				Config: testAccRoutingDomainConfig("test_int_gateway", fqdn, "internal_via_gateway", "web", "Terraform acceptance test - internal via gateway UPDATED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.test_int_gateway"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "fqdn", fqdn),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "type", "internal_via_gateway"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "app_type", "web"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "comment", "Terraform acceptance test - internal via gateway UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "flag", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "ip", "false"),
					resource.TestCheckResourceAttr("citrixspa_routing_domain.test_int_gateway", "location_ids.#", "0"),
				),
			},
			{
				ResourceName:                         "citrixspa_routing_domain.test_int_gateway",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fqdn,
				ImportStateVerifyIdentifierAttribute: "fqdn",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// =============================================================================
// Unit tests — drift warning when a routing domain is deleted outside Terraform
// =============================================================================

type mockRoutingDomainClient struct {
	SPAClient
	getRoutingDomain    func(ctx context.Context, fqdn string) (*RoutingDomain, error)
	createRoutingDomain func(ctx context.Context, rd *RoutingDomain) (*RoutingDomain, error)
}

func (m *mockRoutingDomainClient) GetRoutingDomain(ctx context.Context, fqdn string) (*RoutingDomain, error) {
	return m.getRoutingDomain(ctx, fqdn)
}

func (m *mockRoutingDomainClient) CreateRoutingDomain(ctx context.Context, rd *RoutingDomain) (*RoutingDomain, error) {
	return m.createRoutingDomain(ctx, rd)
}

func testRoutingDomainState() RoutingDomainResourceModel {
	return RoutingDomainResourceModel{
		FQDN:        types.StringValue("intranet.example.com"),
		Type:        types.StringValue("internal"),
		AppType:     types.StringValue("web"),
		Comment:     types.StringNull(),
		Flag:        types.StringValue("enabled"),
		Error:       types.StringValue("none"),
		IP:          types.BoolValue(false),
		LocationIds: types.ListNull(types.StringType),
	}
}

func runRoutingDomainRead(t *testing.T, client SPAClient, model RoutingDomainResourceModel) (tfsdk.State, diag.Diagnostics) {
	t.Helper()
	ctx := context.Background()
	r := &RoutingDomainResource{client: client}

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", schemaResp.Diagnostics.Errors())
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("failed to seed prior state: %v", diags.Errors())
	}

	readResp := &fwresource.ReadResponse{State: state}
	r.Read(ctx, fwresource.ReadRequest{State: state}, readResp)

	return readResp.State, readResp.Diagnostics
}

func TestRoutingDomainRead404WarnsAndRemovesFromState(t *testing.T) {
	client := &mockRoutingDomainClient{
		getRoutingDomain: func(ctx context.Context, fqdn string) (*RoutingDomain, error) {
			return nil, &APIError{StatusCode: 404, TransactionID: "test-txid", Body: "routing domain not found"}
		},
	}

	state, diags := runRoutingDomainRead(t, client, testRoutingDomainState())

	if diags.HasError() {
		t.Fatalf("expected no error diagnostics on a 404 read, got: %v", diags.Errors())
	}

	warnings := diags.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning on a 404 read, got %d: %v", len(warnings), warnings)
	}

	detail := warnings[0].Detail()

	if !strings.Contains(detail, "intranet.example.com") {
		t.Errorf("warning does not name the missing fqdn: %q", detail)
	}
	if !strings.Contains(detail, "Note 3") {
		t.Errorf("warning does not point at the routing domain documentation note: %q", detail)
	}
	if strings.Contains(detail, "terraform state rm") {
		t.Errorf("warning must not tell the user to run terraform state rm: %q", detail)
	}
	if strings.Contains(detail, "removed from the Terraform state") {
		t.Errorf("warning must not claim the state file was rewritten — a plan refresh does not persist state: %q", detail)
	}

	if !strings.Contains(detail, "remove its resource block from the configuration") {
		t.Errorf("warning does not carry the remove-from-configuration remedy: %q", detail)
	}
	// A 404 read also happens while the resource block is being removed, where
	// Terraform drops the object and plans no create. The warning must therefore
	// offer recreation conditionally, never predict it.
	if strings.Contains(detail, "Terraform will plan to recreate") {
		t.Errorf("warning must not predict a recreate — none is planned when the resource block is being removed: %q", detail)
	}
	if !strings.Contains(detail, "If it is still required, the next terraform apply recreates it") {
		t.Errorf("warning does not offer recreation conditionally: %q", detail)
	}
	// The cleanup is gated per tenant, and a routing domain can also be deleted
	// straight from the console, so the cascade must be offered as the likely
	// cause rather than asserted as the cause.
	if !strings.Contains(detail, "The most common cause is") {
		t.Errorf("warning states the application cascade as definitive rather than likely: %q", detail)
	}

	if !state.Raw.IsNull() {
		t.Errorf("expected the routing domain to be removed from state after a 404 read")
	}
}

// Create refreshes the routing domain straight after creating it. A 404 on that
// read means the service has not converged yet — it must not be reported as an
// application cascade, and the resource must stay in state, or the apply fails
// with "Missing Resource State After Create".
func TestRoutingDomainCreate404DoesNotBlameApplicationCascade(t *testing.T) {
	ctx := context.Background()
	client := &mockRoutingDomainClient{
		createRoutingDomain: func(ctx context.Context, rd *RoutingDomain) (*RoutingDomain, error) {
			return &RoutingDomain{FQDN: rd.FQDN, Type: rd.Type, Error: "none"}, nil
		},
		getRoutingDomain: func(ctx context.Context, fqdn string) (*RoutingDomain, error) {
			return nil, &APIError{StatusCode: 404, TransactionID: "test-txid", Body: "routing domain not found"}
		},
	}

	r := &RoutingDomainResource{client: client}
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", schemaResp.Diagnostics.Errors())
	}

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	model := testRoutingDomainState()
	if diags := plan.Set(ctx, &model); diags.HasError() {
		t.Fatalf("failed to seed plan: %v", diags.Errors())
	}

	createResp := &fwresource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, fwresource.CreateRequest{Plan: plan}, createResp)

	if createResp.Diagnostics.HasError() {
		t.Fatalf("expected no error diagnostics on a post-create 404, got: %v", createResp.Diagnostics.Errors())
	}

	warnings := createResp.Diagnostics.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning on a post-create 404, got %d: %v", len(warnings), warnings)
	}

	if summary := warnings[0].Summary(); summary != "Routing Domain Not Readable After Creation" {
		t.Errorf("post-create 404 must not reuse the drift warning, got summary %q", summary)
	}
	if detail := warnings[0].Detail(); strings.Contains(detail, "deletes the routing domains") {
		t.Errorf("post-create 404 must not blame an application deletion: %q", detail)
	}

	if createResp.State.Raw.IsNull() {
		t.Error("a just-created routing domain must stay in state after a failed read-back")
	}
}

func TestRoutingDomainReadNon404ErrorsWithoutWarning(t *testing.T) {
	// Both shapes must behave identically: a typed API failure, and a bare
	// error that never came from handleResponse at all (a transport failure,
	// say). The latter is the only coverage that IsNotFound stays false for
	// errors it cannot classify, so keep it alongside the typed one.
	errs := []struct {
		name string
		err  error
	}{
		{"typed API error", &APIError{
			StatusCode:    http.StatusInternalServerError,
			TransactionID: "cccccccc-dddd-eeee-ffff-000000000000",
			Body:          `{"detail":"internal server error"}`,
		}},
		{"untyped error", fmt.Errorf("API request failed with status 500: internal server error")},
	}

	for _, tc := range errs {
		t.Run(tc.name, func(t *testing.T) {
			client := &mockRoutingDomainClient{
				getRoutingDomain: func(ctx context.Context, fqdn string) (*RoutingDomain, error) {
					return nil, tc.err
				},
			}

			state, diags := runRoutingDomainRead(t, client, testRoutingDomainState())

			if !diags.HasError() {
				t.Fatalf("expected an error diagnostic when the read fails with a non-404 error")
			}
			if got := len(diags.Warnings()); got != 0 {
				t.Errorf("expected no warnings on a non-404 read failure, got %d: %v", got, diags.Warnings())
			}
			if state.Raw.IsNull() {
				t.Errorf("expected state to be retained when the read fails with a non-404 error")
			}
		})
	}
}
