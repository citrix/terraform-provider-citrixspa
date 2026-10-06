package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccCertificatesDataSourceConfig returns a config that reads all
// certificates without pagination arguments.
func testAccCertificatesDataSourceConfig() string {
	return `
data "citrixspa_certificates" "all" {}
`
}

// =============================================================================
// Tests — citrixspa_certificates data source
// =============================================================================

// TestAccCertificatesDataSource_basic reads the full certificate list and
// verifies the computed list attribute is present in state. The list may be
// empty on a fresh tenant, so the count is asserted as set rather than positive.
func TestAccCertificatesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificatesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrixspa_certificates.all", "certificates.#"),
				),
			},
		},
	})
}
