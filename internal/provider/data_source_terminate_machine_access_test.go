package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccTerminateMachineAccessDataSourceConfig returns a config that reads all
// machine access termination records without pagination arguments.
func testAccTerminateMachineAccessDataSourceConfig() string {
	return `
data "citrixspa_terminate_machine_access" "all" {}
`
}

// =============================================================================
// Tests — citrixspa_terminate_machine_access data source
// =============================================================================

// TestAccTerminateMachineAccessDataSource_basic reads the full machine
// termination list and verifies the computed list attribute is present in
// state. The list may be empty, so the count is asserted as set.
func TestAccTerminateMachineAccessDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTerminateMachineAccessDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.citrixspa_terminate_machine_access.all", "machines.#"),
				),
			},
		},
	})
}
