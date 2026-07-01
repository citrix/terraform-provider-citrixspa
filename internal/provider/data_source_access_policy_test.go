package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccAccessPolicyDataSourceByNameConfig creates a policy resource (with its
// required routing domain + application) and reads it back via name lookup.
func testAccAccessPolicyDataSourceByNameConfig(policyName string) string {
	const rdResourceName = "ds_policy_rd"
	const appResourceName = "ds_policy_app"
	const fqdn = "tf-acc-ds-policy-byname.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policy by-name routing domain", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policy-app-byname",
		appType:      "web",
		description:  "DS policy acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "ds_policy",
		name:            policyName,
		description:     "DS by-name acceptance test",
		active:          false,
		appResourceName: appResourceName,
		priority:        990,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	dataSourceConfig := fmt.Sprintf(`
data "spa_access_policy" "by_name" {
  name       = %q
  depends_on = [spa_access_policy.ds_policy]
}
`, policyName)

	return rdConfig + appConfig + policyConfig + dataSourceConfig
}

// testAccAccessPolicyDataSourceByIDConfig creates a policy resource and reads
// it back via ID lookup (using the resource's computed ID).
func testAccAccessPolicyDataSourceByIDConfig(policyName string) string {
	const rdResourceName = "ds_policy_rd"
	const appResourceName = "ds_policy_app"
	const fqdn = "tf-acc-ds-policy-byid.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policy by-ID routing domain", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policy-app-byid",
		appType:      "web",
		description:  "DS policy by-ID acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "ds_policy",
		name:            policyName,
		description:     "DS by-ID acceptance test",
		active:          false,
		appResourceName: appResourceName,
		priority:        989,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	dataSourceConfig := `
data "spa_access_policy" "by_id" {
  id         = spa_access_policy.ds_policy.id
  depends_on = [spa_access_policy.ds_policy]
}
`
	return rdConfig + appConfig + policyConfig + dataSourceConfig
}

// testAccAccessPolicyDataSourceUpdatedConfig updates the policy and re-reads it
// by name so the data source must reflect the new values.
func testAccAccessPolicyDataSourceUpdatedConfig(policyName, description string, active bool, priority int) string {
	const rdResourceName = "ds_policy_rd"
	const appResourceName = "ds_policy_app"
	const fqdn = "tf-acc-ds-policy-byname.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policy by-name routing domain", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policy-app-byname",
		appType:      "web",
		description:  "DS policy acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "ds_policy",
		name:            policyName,
		description:     description,
		active:          active,
		appResourceName: appResourceName,
		priority:        priority,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	dataSourceConfig := fmt.Sprintf(`
data "spa_access_policy" "by_name" {
  name       = %q
  depends_on = [spa_access_policy.ds_policy]
}
`, policyName)

	return rdConfig + appConfig + policyConfig + dataSourceConfig
}

// testAccAccessPolicyDataSourceNotFoundConfig tries to look up a policy that
// should never exist.
func testAccAccessPolicyDataSourceNotFoundConfig() string {
	return `
data "spa_access_policy" "missing" {
  name = "tf-acc-nonexistent-policy-should-not-exist"
}
`
}

// =============================================================================
// Cleanup helper
// =============================================================================

// testAccCleanupAccessPolicyByName deletes any access policy matching the
// given name left over from a previous failed test run. Errors are ignored.
func testAccCleanupAccessPolicyByName(policyName string) {
	client, err := testAccCreateClient()
	if err != nil {
		return
	}
	ctx := context.Background()

	result, err := client.GetAccessPolicies(ctx, 0, -1, policyName, "name")
	if err != nil {
		return
	}
	for _, p := range result.Policies {
		if strings.EqualFold(p.Name, policyName) {
			_ = client.DeleteAccessPolicy(ctx, p.ID)
		}
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccAccessPolicyDataSource_byName creates a policy and reads it back via
// name lookup, asserting all top-level scalar attributes match the resource.
func TestAccAccessPolicyDataSource_byName(t *testing.T) {
	policyName := "tf-acc-ds-policy-byname"
	appFQDN := "tf-acc-ds-policy-byname.example.com"
	appName := "tf-acc-ds-policy-app-byname"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupAccessPolicyByName(policyName)
			testAccCleanupRoutingDomain(appFQDN)
			testAccCleanupApplicationByName(appName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyDataSourceByNameConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.ds_policy"),
					// id must be populated
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_name", "id"),
					// top-level scalar attributes
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "name", policyName),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "description", "DS by-name acceptance test"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "active", "false"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "priority", "990"),
					// apps set must be non-empty
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_name", "apps.#"),
					// access_rules list must be present
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_name", "access_rules.#"),
					// modified must be populated
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_name", "modified"),
					// data source must match the resource
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "id",
						"spa_access_policy.ds_policy", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "name",
						"spa_access_policy.ds_policy", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "description",
						"spa_access_policy.ds_policy", "description",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "active",
						"spa_access_policy.ds_policy", "active",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "priority",
						"spa_access_policy.ds_policy", "priority",
					),
				),
			},
		},
	})
}

// TestAccAccessPolicyDataSource_byID creates a policy and reads it back via ID
// lookup, asserting attribute parity with the resource.
func TestAccAccessPolicyDataSource_byID(t *testing.T) {
	policyName := "tf-acc-ds-policy-byid"
	appFQDN := "tf-acc-ds-policy-byid.example.com"
	appName := "tf-acc-ds-policy-app-byid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupAccessPolicyByName(policyName)
			testAccCleanupRoutingDomain(appFQDN)
			testAccCleanupApplicationByName(appName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyDataSourceByIDConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.ds_policy"),
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_id", "id"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_id", "name", policyName),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_id", "description", "DS by-ID acceptance test"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_id", "active", "false"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_id", "priority", "989"),
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_id", "apps.#"),
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_id", "access_rules.#"),
					resource.TestCheckResourceAttrSet("data.spa_access_policy.by_id", "modified"),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_id", "id",
						"spa_access_policy.ds_policy", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_id", "name",
						"spa_access_policy.ds_policy", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_id", "active",
						"spa_access_policy.ds_policy", "active",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_id", "priority",
						"spa_access_policy.ds_policy", "priority",
					),
				),
			},
		},
	})
}

// TestAccAccessPolicyDataSource_reflectsUpdate creates a policy then updates it,
// verifying the data source returns the fresh values.
func TestAccAccessPolicyDataSource_reflectsUpdate(t *testing.T) {
	policyName := "tf-acc-ds-policy-upd"
	appFQDN := "tf-acc-ds-policy-byname.example.com"
	appName := "tf-acc-ds-policy-app-byname"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupAccessPolicyByName(policyName)
			testAccCleanupRoutingDomain(appFQDN)
			testAccCleanupApplicationByName(appName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: create
			{
				Config: testAccAccessPolicyDataSourceUpdatedConfig(policyName, "initial description", false, 985),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "name", policyName),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "description", "initial description"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "active", "false"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "priority", "985"),
				),
			},
			// Step 2: update description, active and priority; data source must reflect changes
			{
				Config: testAccAccessPolicyDataSourceUpdatedConfig(policyName, "updated description", true, 984),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "name", policyName),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "description", "updated description"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "active", "true"),
					resource.TestCheckResourceAttr("data.spa_access_policy.by_name", "priority", "984"),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "description",
						"spa_access_policy.ds_policy", "description",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "active",
						"spa_access_policy.ds_policy", "active",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policy.by_name", "priority",
						"spa_access_policy.ds_policy", "priority",
					),
				),
			},
		},
	})
}

// TestAccAccessPolicyDataSource_notFound verifies that looking up a non-existent
// policy produces an error.
func TestAccAccessPolicyDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAccessPolicyDataSourceNotFoundConfig(),
				ExpectError: regexp.MustCompile(`(?i)(access policy not found|no access policy found|unable to read)`),
			},
		},
	})
}
