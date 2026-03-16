package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

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
		if rs.Type != "spa_routing_domain" {
			continue
		}

		fqdn := rs.Primary.Attributes["fqdn"]
		_, err := client.GetRoutingDomain(ctx, fqdn)
		if err == nil {
			return fmt.Errorf("routing domain %s still exists in the API after destroy", fqdn)
		}
		if !strings.Contains(err.Error(), "404") {
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
resource "spa_routing_domain" "%s" {
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain("tf-acc-test-internal.example.com") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainConfig("test_internal", "tf-acc-test-internal.example.com", "internal", "web", "Terraform acceptance test - internal", "enabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.test_internal"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "fqdn", "tf-acc-test-internal.example.com"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "type", "internal"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "app_type", "web"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "comment", "Terraform acceptance test - internal"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "flag", "enabled"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "ip", "false"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "location_ids.#", "0"),
				),
			},
			{
				Config: testAccRoutingDomainConfig("test_internal", "tf-acc-test-internal.example.com", "internal", "web", "Terraform acceptance test - internal UPDATED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.test_internal"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "fqdn", "tf-acc-test-internal.example.com"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "type", "internal"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "app_type", "web"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "comment", "Terraform acceptance test - internal UPDATED"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "flag", "disabled"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "ip", "false"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_internal", "location_ids.#", "0"),
				),
			},
			{
				ResourceName:                         "spa_routing_domain.test_internal",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "tf-acc-test-internal.example.com",
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
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.test_external"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "fqdn", fqdn),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "type", "external"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "app_type", "saas"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "comment", "Terraform acceptance test - external"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "flag", "enabled"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "ip", "false"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "location_ids.#", "0"),
				),
			},
			{
				Config: testAccRoutingDomainConfig("test_external", fqdn, "external", "saas", "Terraform acceptance test - external UPDATED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.test_external"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "fqdn", fqdn),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "type", "external"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "app_type", "saas"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "comment", "Terraform acceptance test - external UPDATED"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "flag", "disabled"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "ip", "false"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_external", "location_ids.#", "0"),
				),
			},
			{
				ResourceName:                         "spa_routing_domain.test_external",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fqdn,
				ImportStateVerifyIdentifierAttribute: "fqdn",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRoutingDomain_externalViaConnector(t *testing.T) {
	fqdn := "tf-acc-test-ext-connector.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainConfig("test_ext_connector", fqdn, "external_via_connector", "web", "Terraform acceptance test - external via connector", "enabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.test_ext_connector"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "fqdn", fqdn),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "type", "external_via_connector"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "app_type", "web"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "comment", "Terraform acceptance test - external via connector"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "flag", "enabled"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "ip", "false"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "location_ids.#", "0"),
				),
			},
			{
				Config: testAccRoutingDomainConfig("test_ext_connector", fqdn, "external_via_connector", "web", "Terraform acceptance test - external via connector UPDATED", "disabled", "false", "[]"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.test_ext_connector"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "fqdn", fqdn),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "type", "external_via_connector"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "app_type", "web"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "comment", "Terraform acceptance test - external via connector UPDATED"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "flag", "disabled"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "ip", "false"),
					resource.TestCheckResourceAttr("spa_routing_domain.test_ext_connector", "location_ids.#", "0"),
				),
			},
			{
				ResourceName:                         "spa_routing_domain.test_ext_connector",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        fqdn,
				ImportStateVerifyIdentifierAttribute: "fqdn",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
