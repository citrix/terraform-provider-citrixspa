package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccTerminateUserAccessDataSourceConfig returns a config that reads all
// user access termination records without pagination arguments.
func testAccTerminateUserAccessDataSourceConfig() string {
	return `
data "citrixspa_terminate_user_access" "all" {}
`
}

// =============================================================================
// Tests — citrixspa_terminate_user_access data source
// =============================================================================

// TestAccTerminateUserAccessDataSource_basic reads the full user termination
// list and verifies the computed list attribute is present in state. The list
// may be empty, so the count is asserted as set.
func TestAccTerminateUserAccessDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTerminateUserAccessDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrixspa_terminate_user_access.all", "users.#"),
				),
			},
		},
	})
}
