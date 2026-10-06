package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccLastActivityDataSourceConfig returns a config that reads the tenant's
// last activity timestamp. The data source takes no arguments.
func testAccLastActivityDataSourceConfig() string {
	return `
data "citrixspa_last_activity" "test" {}
`
}

// =============================================================================
// Tests — citrixspa_last_activity data source
// =============================================================================

// TestAccLastActivityDataSource_basic reads the last activity timestamp and
// verifies the computed attribute is populated in state.
func TestAccLastActivityDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLastActivityDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrixspa_last_activity.test", "last_activity"),
				),
			},
		},
	})
}
