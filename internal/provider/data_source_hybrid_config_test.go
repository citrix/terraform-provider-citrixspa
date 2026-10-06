package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccHybridConfigDataSourceConfig returns a config that reads the tenant's
// hybrid configuration. The data source takes no arguments.
func testAccHybridConfigDataSourceConfig() string {
	return `
data "citrixspa_hybrid_config" "test" {}
`
}

// =============================================================================
// Tests — citrixspa_hybrid_config data source
// =============================================================================

// TestAccHybridConfigDataSource_basic reads the hybrid configuration and
// verifies the computed boolean attributes are populated in state.
func TestAccHybridConfigDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHybridConfigDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrixspa_hybrid_config.test", "first_time"),
					resource.TestCheckResourceAttrSet("data.citrixspa_hybrid_config.test", "is_hybrid"),
				),
			},
		},
	})
}
