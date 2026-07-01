package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccSecurityGroupsDataSourceConfig creates a prerequisite app and security
// group, then reads all security groups via the plural data source.
func testAccSecurityGroupsDataSourceConfig(appName, sgName string) string {
	return testAccSecurityGroupPrereqAppConfig("sgs_app", appName) +
		testAccSecurityGroupConfig(testSecurityGroupConfig{
			resourceName:   "sgs_sg",
			name:           sgName,
			appRefs:        []string{"spa_application.sgs_app.id"},
			systemIn:       "enabled",
			systemOut:      "enabled",
			unpublishedIn:  "disabled",
			unpublishedOut: "disabled",
		}) + `
data "spa_security_groups" "all" {
  depends_on = [spa_security_group.sgs_sg]
}
`
}

// =============================================================================
// Custom check helpers
// =============================================================================

// testAccCheckSecurityGroupsListNonEmpty verifies the security_groups list has
// at least one element.
func testAccCheckSecurityGroupsListNonEmpty(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		countStr, ok := rs.Primary.Attributes["security_groups.#"]
		if !ok {
			return fmt.Errorf("security_groups.# not found in state for %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("could not parse security_groups.# value %q: %w", countStr, err)
		}

		if count < 1 {
			return fmt.Errorf("expected at least 1 security group, got %d", count)
		}
		return nil
	}
}

// testAccCheckSecurityGroupsContainsName verifies that the given security group
// name appears in the list returned by the data source.
func testAccCheckSecurityGroupsContainsName(dataSourceName, sgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		countStr, ok := rs.Primary.Attributes["security_groups.#"]
		if !ok {
			return fmt.Errorf("security_groups.# not found in state for %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("could not parse security_groups.# value %q: %w", countStr, err)
		}

		for i := 0; i < count; i++ {
			key := fmt.Sprintf("security_groups.%d.name", i)
			if rs.Primary.Attributes[key] == sgName {
				return nil
			}
		}

		return fmt.Errorf("security group %q not found in %s (checked %d items)", sgName, dataSourceName, count)
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccSecurityGroupsDataSource_basic creates a security group and verifies
// the plural data source returns a non-empty list.
func TestAccSecurityGroupsDataSource_basic(t *testing.T) {
	appName := "tf-acc-ds-sgs-app"
	sgName := "tf-acc-ds-sgs-basic"

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
				Config: testAccSecurityGroupsDataSourceConfig(appName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI("spa_security_group.sgs_sg"),
					testAccCheckSecurityGroupsListNonEmpty("data.spa_security_groups.all"),
				),
			},
		},
	})
}

// TestAccSecurityGroupsDataSource_containsCreated creates a security group and
// verifies it appears by name in the plural data source list.
func TestAccSecurityGroupsDataSource_containsCreated(t *testing.T) {
	appName := "tf-acc-ds-sgs-contains-app"
	sgName := "tf-acc-ds-sgs-contains"

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
				Config: testAccSecurityGroupsDataSourceConfig(appName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI("spa_security_group.sgs_sg"),
					testAccCheckSecurityGroupsContainsName("data.spa_security_groups.all", sgName),
				),
			},
		},
	})
}
