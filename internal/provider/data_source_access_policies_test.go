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

// testAccAccessPoliciesDataSourceConfig returns a config that reads all access
// policies without any filters.
func testAccAccessPoliciesDataSourceConfig() string {
	return `
data "spa_access_policies" "all" {}
`
}

// testAccAccessPoliciesDataSourceBasicConfig creates one policy (with the
// required routing domain + app prerequisites) and reads all access policies
// so the list is guaranteed to be non-empty.
func testAccAccessPoliciesDataSourceBasicConfig(policyName string) string {
	const rdResourceName = "basic_rd"
	const appResourceName = "basic_app"
	const fqdn = "tf-acc-ds-policies-basic.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policies basic routing domain", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policies-basic-app",
		appType:      "web",
		description:  "DS policies basic acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})
	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "basic_policy",
		name:            policyName,
		description:     "DS policies basic acceptance test",
		active:          false,
		appResourceName: appResourceName,
		priority:        975,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})
	return rdConfig + appConfig + policyConfig + `
data "spa_access_policies" "all" {
  depends_on = [spa_access_policy.basic_policy]
}
`
}

// testAccAccessPoliciesDataSourcePaginationConfig creates one policy and reads
// it back with offset=0, limit=1 so that exactly 1 item is guaranteed.
func testAccAccessPoliciesDataSourcePaginationConfig(policyName string) string {
	const rdResourceName = "pagination_rd"
	const appResourceName = "pagination_app"
	const fqdn = "tf-acc-ds-policies-pagination.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policies pagination routing domain", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policies-pagination-app",
		appType:      "web",
		description:  "DS policies pagination acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})
	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "pagination_policy",
		name:            policyName,
		description:     "DS policies pagination acceptance test",
		active:          false,
		appResourceName: appResourceName,
		priority:        974,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})
	return rdConfig + appConfig + policyConfig + `
data "spa_access_policies" "paged" {
  offset     = 0
  limit      = 1
  depends_on = [spa_access_policy.pagination_policy]
}
`
}

// testAccAccessPoliciesDataSourceOffsetConfig creates two policies (sharing one
// routing domain and app) and exposes two overlapping paginated data source
// reads so the sliding-window relationship between offset=0/limit=2 and
// offset=1/limit=2 can be verified.
func testAccAccessPoliciesDataSourceOffsetConfig(policy1Name, policy2Name string) string {
	const rdResourceName = "offset_rd"
	const appResourceName = "offset_app"
	const fqdn = "tf-acc-ds-policies-offset.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policies offset routing domain", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policies-offset-app",
		appType:      "web",
		description:  "DS policies offset acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})
	policy1Config := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "offset_policy_1",
		name:            policy1Name,
		description:     "DS policies offset acceptance test 1",
		active:          false,
		appResourceName: appResourceName,
		priority:        973,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})
	policy2Config := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "offset_policy_2",
		name:            policy2Name,
		description:     "DS policies offset acceptance test 2",
		active:          false,
		appResourceName: appResourceName,
		priority:        972,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})
	return rdConfig + appConfig + policy1Config + policy2Config + `
data "spa_access_policies" "page1" {
  offset     = 0
  limit      = 2
  depends_on = [spa_access_policy.offset_policy_1, spa_access_policy.offset_policy_2]
}
data "spa_access_policies" "page2" {
  offset     = 1
  limit      = 2
  depends_on = [spa_access_policy.offset_policy_1, spa_access_policy.offset_policy_2]
}
`
}

// testAccAccessPoliciesDataSourceRequiredFieldsConfig creates one policy and
// reads up to 5 items so required fields on the first item can be verified.
func testAccAccessPoliciesDataSourceRequiredFieldsConfig(policyName string) string {
	const rdResourceName = "fields_rd"
	const appResourceName = "fields_app"
	const fqdn = "tf-acc-ds-policies-fields.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policies fields routing domain", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policies-fields-app",
		appType:      "web",
		description:  "DS policies fields acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})
	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "fields_policy",
		name:            policyName,
		description:     "DS policies fields acceptance test",
		active:          false,
		appResourceName: appResourceName,
		priority:        971,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})
	return rdConfig + appConfig + policyConfig + `
data "spa_access_policies" "paged" {
  offset     = 0
  limit      = 5
  depends_on = [spa_access_policy.fields_policy]
}
`
}

// testAccAccessPoliciesDataSourceConfigWithPagination returns a config that
// reads access policies with explicit offset/limit values.
func testAccAccessPoliciesDataSourceConfigWithPagination(offset, limit int) string {
	return fmt.Sprintf(`
data "spa_access_policies" "paged" {
  offset = %d
  limit  = %d
}
`, offset, limit)
}

// testAccAccessPoliciesDataSourceWithResourceConfig creates a full policy
// (routing domain + app + policy) and a data source that reads all policies,
// so we can verify the created policy appears in the list.
func testAccAccessPoliciesDataSourceWithResourceConfig(policyName string) string {
	const rdResourceName = "ds_policies_rd"
	const appResourceName = "ds_policies_app"
	const fqdn = "tf-acc-ds-policies-list.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policies list routing domain", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policies-app",
		appType:      "web",
		description:  "DS policies list acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "ds_policies",
		name:            policyName,
		description:     "DS policies list acceptance test",
		active:          false,
		appResourceName: appResourceName,
		priority:        980,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig + `
data "spa_access_policies" "all" {
  depends_on = [spa_access_policy.ds_policies]
}
`
}

// testAccAccessPoliciesDataSourceWithNameFilterConfig creates a policy and
// reads only that policy via the name filter.
func testAccAccessPoliciesDataSourceWithNameFilterConfig(policyName string) string {
	const rdResourceName = "ds_policies_rd"
	const appResourceName = "ds_policies_app"
	const fqdn = "tf-acc-ds-policies-list.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"DS policies list routing domain", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-ds-policies-app",
		appType:      "web",
		description:  "DS policies list acceptance test app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "ds_policies",
		name:            policyName,
		description:     "DS policies name filter test",
		active:          false,
		appResourceName: appResourceName,
		priority:        979,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig + fmt.Sprintf(`
data "spa_access_policies" "filtered" {
  name       = %q
  depends_on = [spa_access_policy.ds_policies]
}
`, policyName)
}

// =============================================================================
// Custom check helpers
// =============================================================================

// testAccCheckAccessPoliciesListNonEmpty verifies that the access_policies
// list contains at least one item.
func testAccCheckAccessPoliciesListNonEmpty(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(rs.Primary.Attributes["access_policies.#"], "%d", &count); err != nil {
			return fmt.Errorf("could not parse access_policies.#: %w", err)
		}

		if count <= 0 {
			return fmt.Errorf("expected at least one access policy, got %d", count)
		}
		return nil
	}
}

// testAccCheckAccessPoliciesContainsName verifies that the given policy name
// appears in the access_policies list returned by a spa_access_policies data
// source.
func testAccCheckAccessPoliciesContainsName(dataSourceName, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		countStr, ok := rs.Primary.Attributes["access_policies.#"]
		if !ok {
			return fmt.Errorf("access_policies.# not found in state for %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("could not parse access_policies.# value %q: %w", countStr, err)
		}

		for i := 0; i < count; i++ {
			key := fmt.Sprintf("access_policies.%d.name", i)
			if rs.Primary.Attributes[key] == policyName {
				return nil
			}
		}

		return fmt.Errorf("policy with name %q not found in %s (checked %d items)", policyName, dataSourceName, count)
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccAccessPoliciesDataSource_basic creates one policy and reads all access
// policies with no arguments, verifying pagination defaults and a non-empty list.
func TestAccAccessPoliciesDataSource_basic(t *testing.T) {
	policyName := "tf-acc-ds-policies-basic"
	appFQDN := "tf-acc-ds-policies-basic.example.com"
	appName := "tf-acc-ds-policies-basic-app"

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
				Config: testAccAccessPoliciesDataSourceBasicConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.basic_policy"),
					// pagination defaults are reflected in state
					resource.TestCheckResourceAttr("data.spa_access_policies.all", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_access_policies.all", "limit", "-1"),
					// list must be present and non-empty
					resource.TestCheckResourceAttrSet("data.spa_access_policies.all", "access_policies.#"),
					testAccCheckAccessPoliciesListNonEmpty("data.spa_access_policies.all"),
				),
			},
		},
	})
}

// TestAccAccessPoliciesDataSource_withResource creates a policy resource and
// verifies it appears in the list returned by the data source.
func TestAccAccessPoliciesDataSource_withResource(t *testing.T) {
	policyName := "tf-acc-ds-policies-list"
	appFQDN := "tf-acc-ds-policies-list.example.com"
	appName := "tf-acc-ds-policies-app"

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
				Config: testAccAccessPoliciesDataSourceWithResourceConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.ds_policies"),
					// newly created policy must appear in the list
					testAccCheckAccessPoliciesContainsName("data.spa_access_policies.all", policyName),
					// list must be non-empty
					testAccCheckAccessPoliciesListNonEmpty("data.spa_access_policies.all"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.all", "access_policies.#"),
				),
			},
		},
	})
}

// TestAccAccessPoliciesDataSource_pagination creates one policy, requests a
// single page of limit=1 and verifies exactly 1 policy is returned.
func TestAccAccessPoliciesDataSource_pagination(t *testing.T) {
	policyName := "tf-acc-ds-policies-pagination"
	appFQDN := "tf-acc-ds-policies-pagination.example.com"
	appName := "tf-acc-ds-policies-pagination-app"

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
				Config: testAccAccessPoliciesDataSourcePaginationConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.pagination_policy"),
					resource.TestCheckResourceAttr("data.spa_access_policies.paged", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_access_policies.paged", "limit", "1"),
					// exactly 1 item should come back
					resource.TestCheckResourceAttr("data.spa_access_policies.paged", "access_policies.#", "1"),
					// that one item must have required fields
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.id"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.name"),
				),
			},
		},
	})
}

// TestAccAccessPoliciesDataSource_offset creates two policies, then fetches
// two overlapping pages and confirms the sliding-window relationship:
// page1[offset=0,limit=2][1] must equal page2[offset=1,limit=2][0].
// Creating policies explicitly makes the test self-contained and prevents
// a false positive when the environment has fewer than 2 existing policies.
func TestAccAccessPoliciesDataSource_offset(t *testing.T) {
	policy1Name := "tf-acc-ds-policies-offset-1"
	policy2Name := "tf-acc-ds-policies-offset-2"
	appFQDN := "tf-acc-ds-policies-offset.example.com"
	appName := "tf-acc-ds-policies-offset-app"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupAccessPolicyByName(policy1Name)
			testAccCleanupAccessPolicyByName(policy2Name)
			testAccCleanupRoutingDomain(appFQDN)
			testAccCleanupApplicationByName(appName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPoliciesDataSourceOffsetConfig(policy1Name, policy2Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					// both policies must exist
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.offset_policy_1"),
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.offset_policy_2"),
					// pagination attributes reflected in state
					resource.TestCheckResourceAttr("data.spa_access_policies.page1", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_access_policies.page1", "limit", "2"),
					resource.TestCheckResourceAttr("data.spa_access_policies.page2", "offset", "1"),
					resource.TestCheckResourceAttr("data.spa_access_policies.page2", "limit", "2"),
					// page1 must return exactly 2 items (limit=2 and >=2 policies exist)
					resource.TestCheckResourceAttr("data.spa_access_policies.page1", "access_policies.#", "2"),
					// page2 must return at least 1 item
					testAccCheckAccessPoliciesListNonEmpty("data.spa_access_policies.page2"),
					// sliding-window: page1[1] and page2[0] must be the same policy
					resource.TestCheckResourceAttrPair(
						"data.spa_access_policies.page1", "access_policies.1.id",
						"data.spa_access_policies.page2", "access_policies.0.id",
					),
				),
			},
		},
	})
}

// TestAccAccessPoliciesDataSource_eachItemHasRequiredFields creates one policy
// and checks that the first item in the first page has all mandatory computed
// fields populated.
func TestAccAccessPoliciesDataSource_eachItemHasRequiredFields(t *testing.T) {
	policyName := "tf-acc-ds-policies-fields"
	appFQDN := "tf-acc-ds-policies-fields.example.com"
	appName := "tf-acc-ds-policies-fields-app"

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
				Config: testAccAccessPoliciesDataSourceRequiredFieldsConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.fields_policy"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.#"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.id"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.name"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.active"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.priority"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.apps.#"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.paged", "access_policies.0.modified"),
				),
			},
		},
	})
}

// TestAccAccessPoliciesDataSource_nameFilter creates a policy and reads it
// back using the name filter, verifying exactly that policy is returned.
func TestAccAccessPoliciesDataSource_nameFilter(t *testing.T) {
	policyName := "tf-acc-ds-policies-filter"
	appFQDN := "tf-acc-ds-policies-list.example.com"
	appName := "tf-acc-ds-policies-app"

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
				Config: testAccAccessPoliciesDataSourceWithNameFilterConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.ds_policies"),
					// filtered list must contain our policy
					testAccCheckAccessPoliciesContainsName("data.spa_access_policies.filtered", policyName),
					resource.TestCheckResourceAttr("data.spa_access_policies.filtered", "name", policyName),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.filtered", "access_policies.#"),
					resource.TestCheckResourceAttrSet("data.spa_access_policies.filtered", "access_policies.0.id"),
					resource.TestCheckResourceAttr("data.spa_access_policies.filtered", "access_policies.0.name", policyName),
				),
			},
		},
	})
}

// TestAccAccessPoliciesDataSource_apiDirectVerification creates a policy via
// the resource, then calls GetAccessPolicies directly to confirm the API also
// returns it, comparing count with the data source.
func TestAccAccessPoliciesDataSource_apiDirectVerification(t *testing.T) {
	policyName := "tf-acc-ds-policies-api-verify"
	appFQDN := "tf-acc-ds-policies-list.example.com"
	appName := "tf-acc-ds-policies-app"

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
				Config: testAccAccessPoliciesDataSourceWithResourceConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.ds_policies"),
					func(s *terraform.State) error {
						client, err := testAccCreateClient()
						if err != nil {
							return fmt.Errorf("failed to create API client: %w", err)
						}
						ctx := context.Background()
						result, err := client.GetAccessPolicies(ctx, 0, -1, "", "name")
						if err != nil {
							return fmt.Errorf("GetAccessPolicies API call failed: %w", err)
						}
						for _, p := range result.Policies {
							if p.Name == policyName {
								return nil
							}
						}
						return fmt.Errorf("policy %q not found in GetAccessPolicies response (%d items)", policyName, len(result.Policies))
					},
				),
			},
		},
	})
}
