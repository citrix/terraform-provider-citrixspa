package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccSecurityGroupDataSourceByIDConfig creates a prerequisite web app and a
// security group resource, then reads the group back via the singular data
// source using the resource's computed ID.
func testAccSecurityGroupDataSourceByIDConfig(appName, sgName string) string {
	return testAccSecurityGroupPrereqAppConfig("ds_sg_app", appName) +
		testAccSecurityGroupConfig(testSecurityGroupConfig{
			resourceName:   "ds_sg",
			name:           sgName,
			appRefs:        []string{"spa_application.ds_sg_app.id"},
			systemIn:       "enabled",
			systemOut:      "disabled",
			unpublishedIn:  "disabled",
			unpublishedOut: "disabled",
		}) + `
data "spa_security_group" "by_id" {
  id         = spa_security_group.ds_sg.id
  depends_on = [spa_security_group.ds_sg]
}
`
}

// testAccSecurityGroupDataSourceNotFoundConfig references a non-existent ID.
func testAccSecurityGroupDataSourceNotFoundConfig() string {
	return `
data "spa_security_group" "missing" {
  id = "00000000-0000-0000-0000-000000000000"
}
`
}

// =============================================================================
// Tests
// =============================================================================

// TestAccSecurityGroupDataSource_byID creates a security group resource and
// reads it back via the data source; verifies all attributes match.
func TestAccSecurityGroupDataSource_byID(t *testing.T) {
	appName := "tf-acc-ds-sg-app"
	sgName := "tf-acc-ds-sg-byid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSecurityGroupByName(sgName)
			testAccCleanupApplicationByName(appName)
			testAccCleanupRoutingDomain(appName + ".example.com")
			testAccCleanupRoutingDomain("api." + appName + ".example.com")
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupDataSourceByIDConfig(appName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI("spa_security_group.ds_sg"),
					// Verify data source fields are populated
					resource.TestCheckResourceAttrSet("data.spa_security_group.by_id", "id"),
					resource.TestCheckResourceAttr("data.spa_security_group.by_id", "name", sgName),
					resource.TestCheckResourceAttr("data.spa_security_group.by_id", "system.data_in", "enabled"),
					resource.TestCheckResourceAttr("data.spa_security_group.by_id", "system.data_out", "disabled"),
					resource.TestCheckResourceAttr("data.spa_security_group.by_id", "unpublished_app.data_in", "disabled"),
					resource.TestCheckResourceAttr("data.spa_security_group.by_id", "unpublished_app.data_out", "disabled"),
					resource.TestCheckResourceAttrSet("data.spa_security_group.by_id", "modified"),
					// Data source values must match the resource
					resource.TestCheckResourceAttrPair(
						"data.spa_security_group.by_id", "id",
						"spa_security_group.ds_sg", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_security_group.by_id", "name",
						"spa_security_group.ds_sg", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_security_group.by_id", "app_ids.#",
						"spa_security_group.ds_sg", "app_ids.#",
					),
				),
			},
		},
	})
}

// TestAccSecurityGroupDataSource_notFound verifies that looking up a
// non-existent security group produces an error.
func TestAccSecurityGroupDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecurityGroupDataSourceNotFoundConfig(),
				ExpectError: regexp.MustCompile(`(?i)(not found|unable to read|error)`),
			},
		},
	})
}
