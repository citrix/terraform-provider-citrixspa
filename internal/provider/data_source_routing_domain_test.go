package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccRoutingDomainDataSourceConfig creates a routing domain resource and
// reads it back via the data source, so all computed fields can be verified.
func testAccRoutingDomainDataSourceConfig(fqdn, rdType, appType, comment, flag string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "ds_single" {
  fqdn         = %q
  type         = %q
  app_type     = %q
  comment      = %q
  flag         = %q
  ip           = false
  location_ids = []
}

data "spa_routing_domain" "lookup" {
  fqdn       = spa_routing_domain.ds_single.fqdn
  depends_on = [spa_routing_domain.ds_single]
}
`, fqdn, rdType, appType, comment, flag)
}

// testAccRoutingDomainDataSourceUpdatedConfig updates the resource and re-reads
// it to assert the data source reflects the new values.
func testAccRoutingDomainDataSourceUpdatedConfig(fqdn, comment, flag string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "ds_single" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = %q
  flag         = %q
  ip           = false
  location_ids = []
}

data "spa_routing_domain" "lookup" {
  fqdn       = spa_routing_domain.ds_single.fqdn
  depends_on = [spa_routing_domain.ds_single]
}
`, fqdn, comment, flag)
}

// testAccRoutingDomainDataSourceNonExistentConfig tries to read a domain that
// should not exist, expecting a plan-time or apply-time error.
func testAccRoutingDomainDataSourceNonExistentConfig() string {
	return `
data "spa_routing_domain" "missing" {
  fqdn = "tf-acc-nonexistent-should-not-exist.example.com"
}
`
}

// =============================================================================
// Tests
// =============================================================================

// TestAccRoutingDomainDataSource_basic creates a routing domain resource and
// reads it back via the singular data source, verifying every attribute.
func TestAccRoutingDomainDataSource_basic(t *testing.T) {
	fqdn := "tf-acc-ds-single.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainDataSourceConfig(fqdn, "internal", "web", "DS acceptance test", "enabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source attributes must match the resource
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "fqdn", fqdn),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "type", "internal"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "app_type", "web"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "comment", "DS acceptance test"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "flag", "enabled"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "ip", "false"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "location_ids.#", "0"),
					// error field is always computed by the API
					resource.TestCheckResourceAttrSet("data.spa_routing_domain.lookup", "error"),
					// Data source values must match the resource values exactly
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "fqdn",
						"spa_routing_domain.ds_single", "fqdn",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "type",
						"spa_routing_domain.ds_single", "type",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "app_type",
						"spa_routing_domain.ds_single", "app_type",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "comment",
						"spa_routing_domain.ds_single", "comment",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "flag",
						"spa_routing_domain.ds_single", "flag",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "ip",
						"spa_routing_domain.ds_single", "ip",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "error",
						"spa_routing_domain.ds_single", "error",
					),
				),
			},
		},
	})
}

// TestAccRoutingDomainDataSource_reflectsUpdate verifies the data source reads
// fresh values after the resource is updated.
func TestAccRoutingDomainDataSource_reflectsUpdate(t *testing.T) {
	fqdn := "tf-acc-ds-single-upd.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			// Step 1: create
			{
				Config: testAccRoutingDomainDataSourceConfig(fqdn, "internal", "web", "initial comment", "enabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "fqdn", fqdn),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "comment", "initial comment"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "flag", "enabled"),
				),
			},
			// Step 2: update comment and flag; data source must return new values
			{
				Config: testAccRoutingDomainDataSourceUpdatedConfig(fqdn, "updated comment", "disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "fqdn", fqdn),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "comment", "updated comment"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "flag", "disabled"),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "comment",
						"spa_routing_domain.ds_single", "comment",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domain.lookup", "flag",
						"spa_routing_domain.ds_single", "flag",
					),
				),
			},
		},
	})
}

// TestAccRoutingDomainDataSource_external verifies the data source works with
// an external routing domain type.
func TestAccRoutingDomainDataSource_external(t *testing.T) {
	fqdn := "tf-acc-ds-ext.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainDataSourceConfig(fqdn, "external", "saas", "External DS test", "enabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "fqdn", fqdn),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "type", "external"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "app_type", "saas"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "comment", "External DS test"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "flag", "enabled"),
					resource.TestCheckResourceAttr("data.spa_routing_domain.lookup", "ip", "false"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domain.lookup", "error"),
				),
			},
		},
	})
}

// TestAccRoutingDomainDataSource_notFound verifies that looking up a domain
// that does not exist produces an error.
func TestAccRoutingDomainDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoutingDomainDataSourceNonExistentConfig(),
				ExpectError: regexp.MustCompile(`(?i)(unable to read|404|not found)`),
			},
		},
	})
}
