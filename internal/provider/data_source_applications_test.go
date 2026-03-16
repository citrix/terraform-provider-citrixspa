package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccApplicationsDataSourceConfig returns a config that reads all
// applications without pagination arguments.
func testAccApplicationsDataSourceConfig() string {
	return `
data "spa_applications" "all" {}
`
}

// testAccApplicationsDataSourceBasicConfig creates one web application resource
// and reads all applications so the list is guaranteed to be non-empty.
func testAccApplicationsDataSourceBasicConfig(name string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "basic_test",
		name:         name,
		appType:      "web",
		description:  "Terraform acceptance test basic data source",
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
	}) + `
data "spa_applications" "all" {
  depends_on = [spa_application.basic_test]
}
`
}

// testAccApplicationsDataSourceConfigWithPagination returns a config that
// reads applications with explicit offset/limit values.
func testAccApplicationsDataSourceConfigWithPagination(offset, limit int) string {
	return fmt.Sprintf(`
data "spa_applications" "paged" {
  offset = %d
  limit  = %d
}
`, offset, limit)
}

// testAccApplicationsDataSourcePaginationConfig creates one web application
// resource and reads it back with offset=0, limit=1 so that the pagination
// limit is verified against a known-present item.
func testAccApplicationsDataSourcePaginationConfig(name string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "pagination_test",
		name:         name,
		appType:      "web",
		description:  "Terraform acceptance test pagination",
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
	}) + `
data "spa_applications" "paged" {
  offset     = 0
  limit      = 1
  depends_on = [spa_application.pagination_test]
}
`
}

// testAccApplicationsDataSourceRequiredFieldsConfig creates one web application
// resource and reads up to 5 items so that required fields on each item can be
// verified against a known-present entry.
func testAccApplicationsDataSourceRequiredFieldsConfig(name string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "fields_test",
		name:         name,
		appType:      "web",
		description:  "Terraform acceptance test required fields",
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
	}) + `
data "spa_applications" "paged" {
  offset     = 0
  limit      = 5
  depends_on = [spa_application.fields_test]
}
`
}

// testAccApplicationsDataSourceOffsetConfig creates two web application
// resources and exposes two overlapping paginated data source reads so the
// sliding-window relationship between offset=0/limit=2 and offset=1/limit=2
// can be verified.
func testAccApplicationsDataSourceOffsetConfig(name1, name2 string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "offset_1",
		name:         name1,
		appType:      "web",
		description:  "Terraform acceptance test offset app 1",
		url:          fmt.Sprintf("https://%s.example.com", name1),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name1)},
	}) + testAccApplicationConfig(testAppConfig{
		resourceName: "offset_2",
		name:         name2,
		appType:      "web",
		description:  "Terraform acceptance test offset app 2",
		url:          fmt.Sprintf("https://%s.example.com", name2),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name2)},
	}) + `
data "spa_applications" "page1" {
  offset     = 0
  limit      = 2
  depends_on = [spa_application.offset_1, spa_application.offset_2]
}
data "spa_applications" "page2" {
  offset     = 1
  limit      = 2
  depends_on = [spa_application.offset_1, spa_application.offset_2]
}
`
}

// testAccApplicationsDataSourceWithResourceConfig creates a web application
// resource and reads all applications so we can verify the created app appears
// in the list.
func testAccApplicationsDataSourceWithResourceConfig(name string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "ds_test",
		name:         name,
		appType:      "web",
		description:  "Terraform acceptance test data source",
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
	}) + `
data "spa_applications" "all" {
  depends_on = [spa_application.ds_test]
}
`
}

// =============================================================================
// Custom check helpers
// =============================================================================

// testAccCleanupApplicationByName deletes any application with the given name
// left over from a previous failed test run. Errors are silently ignored.
func testAccCleanupApplicationByName(appName string) {
	client, err := testAccCreateClient()
	if err != nil {
		return
	}
	ctx := context.Background()

	result, err := client.GetApplicationsDetailed(ctx, 0, -1, appName, "", false)
	if err != nil {
		return
	}
	for _, app := range result.Applications {
		if app.Name == appName {
			_ = client.DeleteApplication(ctx, app.ID)
			return
		}
	}
}

// testAccCheckApplicationsContainsName verifies that the given application name
// appears in the applications list returned by a spa_applications data source.
func testAccCheckApplicationsContainsName(dataSourceName, appName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		countStr, ok := rs.Primary.Attributes["applications.#"]
		if !ok {
			return fmt.Errorf("applications.# not found in state for %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("could not parse applications.# value %q: %w", countStr, err)
		}

		for i := 0; i < count; i++ {
			key := fmt.Sprintf("applications.%d.name", i)
			if rs.Primary.Attributes[key] == appName {
				return nil
			}
		}

		return fmt.Errorf("application with name %q not found in %s (checked %d items)", appName, dataSourceName, count)
	}
}

// testAccCheckApplicationsListNonEmpty verifies that the applications list
// contains at least one item.
func testAccCheckApplicationsListNonEmpty(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(rs.Primary.Attributes["applications.#"], "%d", &count); err != nil {
			return fmt.Errorf("could not parse applications.#: %w", err)
		}

		if count <= 0 {
			return fmt.Errorf("expected at least one application, got %d", count)
		}
		return nil
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccApplicationsDataSource_basic creates one application and reads all
// applications with no pagination args, verifying defaults and a non-empty list.
func TestAccApplicationsDataSource_basic(t *testing.T) {
	name := "tf-acc-ds-apps-basic"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourceBasicConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// the resource must exist in the API
					testAccCheckApplicationExistsInAPI("spa_application.basic_test"),
					// pagination defaults are reflected in state
					resource.TestCheckResourceAttr("data.spa_applications.all", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_applications.all", "limit", "-1"),
					// list must be present and non-empty
					resource.TestCheckResourceAttrSet("data.spa_applications.all", "applications.#"),
					testAccCheckApplicationsListNonEmpty("data.spa_applications.all"),
				),
			},
		},
	})
}

// TestAccApplicationsDataSource_withResource creates a web application resource
// and verifies it appears in the list returned by the data source.
func TestAccApplicationsDataSource_withResource(t *testing.T) {
	name := "tf-acc-ds-apps-list"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourceWithResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// the resource must exist in the API
					testAccCheckApplicationExistsInAPI("spa_application.ds_test"),
					// the data source must include the newly created app
					testAccCheckApplicationsContainsName("data.spa_applications.all", name),
					// list must be non-empty
					testAccCheckApplicationsListNonEmpty("data.spa_applications.all"),
					resource.TestCheckResourceAttrSet("data.spa_applications.all", "applications.#"),
				),
			},
		},
	})
}

// TestAccApplicationsDataSource_pagination creates one application, requests
// a single page of limit=1 and verifies exactly 1 application is returned.
func TestAccApplicationsDataSource_pagination(t *testing.T) {
	name := "tf-acc-ds-apps-pagination"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourcePaginationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// the resource must exist in the API
					testAccCheckApplicationExistsInAPI("spa_application.pagination_test"),
					resource.TestCheckResourceAttr("data.spa_applications.paged", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_applications.paged", "limit", "1"),
					// exactly 1 item should come back
					resource.TestCheckResourceAttr("data.spa_applications.paged", "applications.#", "1"),
					// that one item must have required fields
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.id"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.name"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.type"),
				),
			},
		},
	})
}

// TestAccApplicationsDataSource_offset creates two applications, then fetches
// two overlapping pages and confirms the sliding-window relationship:
// page1[offset=0,limit=2][1] must equal page2[offset=1,limit=2][0].
// Creating the applications explicitly makes the test self-contained and
// prevents a false positive when the environment has zero existing apps.
func TestAccApplicationsDataSource_offset(t *testing.T) {
	name1 := "tf-acc-ds-offset-1"
	name2 := "tf-acc-ds-offset-2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name1)
			testAccCleanupApplicationByName(name2)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourceOffsetConfig(name1, name2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// both resources must exist
					testAccCheckApplicationExistsInAPI("spa_application.offset_1"),
					testAccCheckApplicationExistsInAPI("spa_application.offset_2"),
					// pagination attributes reflected in state
					resource.TestCheckResourceAttr("data.spa_applications.page1", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_applications.page1", "limit", "2"),
					resource.TestCheckResourceAttr("data.spa_applications.page2", "offset", "1"),
					resource.TestCheckResourceAttr("data.spa_applications.page2", "limit", "2"),
					// page1 must return exactly 2 items (limit=2 and >=2 apps exist)
					resource.TestCheckResourceAttr("data.spa_applications.page1", "applications.#", "2"),
					// page2 must return at least 1 item (offset=1 with >=2 apps)
					testAccCheckApplicationsListNonEmpty("data.spa_applications.page2"),
					// sliding-window: page1[1] and page2[0] must be the same application
					resource.TestCheckResourceAttrPair(
						"data.spa_applications.page1", "applications.1.id",
						"data.spa_applications.page2", "applications.0.id",
					),
				),
			},
		},
	})
}

// TestAccApplicationsDataSource_eachItemHasRequiredFields verifies that the
// first item in the list has all mandatory computed fields populated.
// TestAccApplicationsDataSource_eachItemHasRequiredFields creates one application
// and verifies that the first item in the list has all mandatory computed fields
// populated.
func TestAccApplicationsDataSource_eachItemHasRequiredFields(t *testing.T) {
	name := "tf-acc-ds-apps-fields"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourceRequiredFieldsConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.fields_test"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.#"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.id"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.name"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.type"),
					resource.TestCheckResourceAttrSet("data.spa_applications.paged", "applications.0.state"),
				),
			},
		},
	})
}

// TestAccApplicationsDataSource_apiDirectVerification creates an application
// via the resource, then calls GetApplicationsDetailed directly to confirm the
// API also returns it.
func TestAccApplicationsDataSource_apiDirectVerification(t *testing.T) {
	name := "tf-acc-ds-apps-api-verify"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationsDataSourceWithResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.ds_test"),
					func(s *terraform.State) error {
						client, err := testAccCreateClient()
						if err != nil {
							return fmt.Errorf("failed to create API client: %w", err)
						}
						ctx := context.Background()
						result, err := client.GetApplicationsDetailed(ctx, 0, -1, "", "", false)
						if err != nil {
							return fmt.Errorf("GetApplicationsDetailed API call failed: %w", err)
						}
						for _, app := range result.Applications {
							if app.Name == name {
								return nil
							}
						}
						return fmt.Errorf("application %q not found in GetApplicationsDetailed response (%d items)", name, len(result.Applications))
					},
				),
			},
		},
	})
}
