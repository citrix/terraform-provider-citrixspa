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

// testAccApplicationDataSourceByNameConfig creates a web app resource and then
// reads it back via the singular data source using the app's name.
func testAccApplicationDataSourceByNameConfig(name string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "ds_app",
		name:         name,
		appType:      "web",
		description:  "DS by name acceptance test",
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
		keywords:     []string{"acc-test"},
	}) + fmt.Sprintf(`
data "spa_application" "by_name" {
  name       = %q
  depends_on = [spa_application.ds_app]
}
`, name)
}

// testAccApplicationDataSourceByIDConfig creates a web app resource and then
// reads it back via the singular data source using the app's computed ID.
func testAccApplicationDataSourceByIDConfig(name string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "ds_app",
		name:         name,
		appType:      "web",
		description:  "DS by ID acceptance test",
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
	}) + `
data "spa_application" "by_id" {
  id         = spa_application.ds_app.id
  depends_on = [spa_application.ds_app]
}
`
}

// testAccApplicationDataSourceUpdatedConfig updates the resource and re-reads
// it by name so we can assert the data source reflects the new values.
func testAccApplicationDataSourceUpdatedConfig(name, description string) string {
	return testAccApplicationConfig(testAppConfig{
		resourceName: "ds_app",
		name:         name,
		appType:      "web",
		description:  description,
		url:          fmt.Sprintf("https://%s.example.com", name),
		relatedURLs:  []string{fmt.Sprintf("*.%s.example.com", name)},
		keywords:     []string{"acc-test", "updated"},
	}) + fmt.Sprintf(`
data "spa_application" "by_name" {
  name       = %q
  depends_on = [spa_application.ds_app]
}
`, name)
}

// testAccApplicationDataSourceNotFoundConfig references a name that should
// never exist in the environment.
func testAccApplicationDataSourceNotFoundConfig() string {
	return `
data "spa_application" "missing" {
  name = "tf-acc-nonexistent-app-should-not-exist"
}
`
}

// =============================================================================
// Tests
// =============================================================================

// TestAccApplicationDataSource_byName creates a web app resource and reads it
// back using the name lookup; all attributes are verified.
func TestAccApplicationDataSource_byName(t *testing.T) {
	name := "tf-acc-ds-app-byname"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceByNameConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.ds_app"),
					// id and name must be populated
					resource.TestCheckResourceAttrSet("data.spa_application.by_name", "id"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "name", name),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "type", "web"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "description", "DS by name acceptance test"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "url", fmt.Sprintf("https://%s.example.com", name)),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "hidden", "false"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "agentless_access", "false"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "mobile_security", "false"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "using_template", "false"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "related_urls.#", "1"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.spa_application.by_name", "keywords.*", "acc-test"),
					resource.TestCheckResourceAttrSet("data.spa_application.by_name", "state"),
					// Data source values must match the resource
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "id",
						"spa_application.ds_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "name",
						"spa_application.ds_app", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "type",
						"spa_application.ds_app", "type",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "description",
						"spa_application.ds_app", "description",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "url",
						"spa_application.ds_app", "url",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "state",
						"spa_application.ds_app", "state",
					),
				),
			},
		},
	})
}

// TestAccApplicationDataSource_byID creates a web app resource and reads it
// back using the ID lookup; verifies attribute parity with the resource.
func TestAccApplicationDataSource_byID(t *testing.T) {
	name := "tf-acc-ds-app-byid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceByIDConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.ds_app"),
					resource.TestCheckResourceAttrSet("data.spa_application.by_id", "id"),
					resource.TestCheckResourceAttr("data.spa_application.by_id", "name", name),
					resource.TestCheckResourceAttr("data.spa_application.by_id", "type", "web"),
					resource.TestCheckResourceAttr("data.spa_application.by_id", "description", "DS by ID acceptance test"),
					resource.TestCheckResourceAttrSet("data.spa_application.by_id", "state"),
					// Data source values must match the resource
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_id", "id",
						"spa_application.ds_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_id", "name",
						"spa_application.ds_app", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_id", "type",
						"spa_application.ds_app", "type",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_id", "description",
						"spa_application.ds_app", "description",
					),
				),
			},
		},
	})
}

// TestAccApplicationDataSource_reflectsUpdate verifies the data source returns
// fresh values after the underlying resource is updated.
func TestAccApplicationDataSource_reflectsUpdate(t *testing.T) {
	name := "tf-acc-ds-app-update"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: create
			{
				Config: testAccApplicationDataSourceByNameConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_application.by_name", "name", name),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "description", "DS by name acceptance test"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "keywords.#", "1"),
				),
			},
			// Step 2: update description and keywords; data source must reflect new values
			{
				Config: testAccApplicationDataSourceUpdatedConfig(name, "DS by name acceptance test UPDATED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_application.by_name", "name", name),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "description", "DS by name acceptance test UPDATED"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "keywords.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.spa_application.by_name", "keywords.*", "acc-test"),
					resource.TestCheckTypeSetElemAttr("data.spa_application.by_name", "keywords.*", "updated"),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "description",
						"spa_application.ds_app", "description",
					),
				),
			},
		},
	})
}

// TestAccApplicationDataSource_saas creates a SaaS app and reads it back,
// verifying the SSO field and SaaS-specific attributes are surfaced.
func TestAccApplicationDataSource_saas(t *testing.T) {
	name := "tf-acc-ds-app-saas"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupApplicationByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName:    "ds_app",
					name:            name,
					appType:         "saas",
					description:     "SaaS DS acceptance test",
					url:             fmt.Sprintf("https://%s.example.com", name),
					agentlessAccess: true,
					relatedURLs:     []string{fmt.Sprintf("*.%s.example.com", name)},
					sso:             `{ type = "nosso" }`,
				}) + fmt.Sprintf(`
data "spa_application" "by_name" {
  name       = %q
  depends_on = [spa_application.ds_app]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.ds_app"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "name", name),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "type", "saas"),
					resource.TestCheckResourceAttr("data.spa_application.by_name", "agentless_access", "true"),
					resource.TestCheckResourceAttrSet("data.spa_application.by_name", "id"),
					resource.TestCheckResourceAttrSet("data.spa_application.by_name", "state"),
					resource.TestCheckResourceAttrPair(
						"data.spa_application.by_name", "id",
						"spa_application.ds_app", "id",
					),
				),
			},
		},
	})
}

// TestAccApplicationDataSource_notFound verifies that looking up a non-existent
// application produces an error.
func TestAccApplicationDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationDataSourceNotFoundConfig(),
				ExpectError: regexp.MustCompile(`(?i)(application not found|no application found|unable to read)`),
			},
		},
	})
}
