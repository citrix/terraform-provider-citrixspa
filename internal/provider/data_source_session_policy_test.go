package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Config helpers — single session policy data source
// =============================================================================

// testAccSessionPolicyDataSourceByNameConfig creates a session policy resource
// and reads it back via name lookup.
func testAccSessionPolicyDataSourceByNameConfig(policyName string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "ds_source" {
  name        = %q
  description = "DS by-name acceptance test"
  active      = false

  generic_rules = [
    {
      name     = "DS Rule"
      priority = 1
      active   = true

      actions = {
        routing          = "default"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        }
      ]
    }
  ]
}

data "spa_session_policy" "by_name" {
  name       = %q
  depends_on = [spa_session_policy.ds_source]
}
`, policyName, policyName)
}

// testAccSessionPolicyDataSourceByIDConfig creates a session policy resource and
// reads it back via ID lookup.
func testAccSessionPolicyDataSourceByIDConfig(policyName string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "ds_source" {
  name        = %q
  description = "DS by-ID acceptance test"
  active      = false

  generic_rules = [
    {
      name     = "DS Rule"
      priority = 1
      active   = true

      actions = {
        routing          = "default"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        }
      ]
    }
  ]
}

data "spa_session_policy" "by_id" {
  id         = spa_session_policy.ds_source.id
  depends_on = [spa_session_policy.ds_source]
}
`, policyName)
}

// testAccSessionPolicyDataSourceNotFoundConfig tries to look up a session policy
// that does not exist and expects the data source to return an error.
func testAccSessionPolicyDataSourceNotFoundConfig() string {
	return `
data "spa_session_policy" "missing" {
  name = "tf-acc-nonexistent-session-policy-should-not-exist"
}
`
}

// =============================================================================
// Config helpers — session policies list data source
// =============================================================================

// testAccSessionPoliciesDataSourceConfig creates a session policy resource and
// reads the list so the new item is guaranteed to be present.
func testAccSessionPoliciesDataSourceConfig(policyName string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "list_source" {
  name        = %q
  description = "DS list acceptance test"
  active      = false

  generic_rules = [
    {
      name     = "List Rule"
      priority = 1
      active   = true

      actions = {
        routing          = "default"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        }
      ]
    }
  ]
}

data "spa_session_policies" "all" {
  depends_on = [spa_session_policy.list_source]
}
`, policyName)
}

// testAccSessionPoliciesDataSourceNameFilterConfig creates a session policy and
// reads the list with a name filter so exactly that policy is returned.
func testAccSessionPoliciesDataSourceNameFilterConfig(policyName string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "filter_source" {
  name        = %q
  description = "DS name-filter acceptance test"
  active      = false

  generic_rules = [
    {
      name     = "Filter Rule"
      priority = 1
      active   = true

      actions = {
        routing          = "default"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        }
      ]
    }
  ]
}

data "spa_session_policies" "filtered" {
  name       = %q
  depends_on = [spa_session_policy.filter_source]
}
`, policyName, policyName)
}

// testAccSessionPoliciesDataSourcePaginationConfig creates one session policy and
// reads it with explicit offset/limit parameters.
func testAccSessionPoliciesDataSourcePaginationConfig(policyName string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "paged_source" {
  name        = %q
  description = "DS pagination acceptance test"
  active      = false

  generic_rules = [
    {
      name     = "Paged Rule"
      priority = 1
      active   = true

      actions = {
        routing          = "default"
        local_lan_access = "enabled"
      }

      condition = [
        {
          type     = "TYPE_PLATFORM"
          operator = "OPERATOR_IN"
          values   = ["PLATFORM_FILTER_PC"]
        }
      ]
    }
  ]
}

data "spa_session_policies" "paged" {
  offset     = 0
  limit      = 5
  depends_on = [spa_session_policy.paged_source]
}
`, policyName)
}

// =============================================================================
// Cleanup helpers
// =============================================================================

func testAccCleanupSessionPolicyByNameForDS(policyName string) {
	client, err := testAccCreateClient()
	if err != nil {
		return
	}
	ctx := context.Background()
	result, err := client.GetSessionPolicies(ctx, 0, -1, policyName, "name")
	if err != nil {
		return
	}
	for _, p := range result.Items {
		if strings.EqualFold(p.Name, policyName) {
			_ = client.DeleteSessionPolicy(ctx, p.ID)
		}
	}
}

// =============================================================================
// Tests — spa_session_policy data source
// =============================================================================

// TestAccSessionPolicyDataSource_byName creates a session policy and reads it
// back by name, asserting all attributes match the resource.
func TestAccSessionPolicyDataSource_byName(t *testing.T) {
	policyName := "tf-acc-ds-session-policy-byname"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByNameForDS(policyName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSessionPolicyDataSourceByNameConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.ds_source"),
					// id must be populated
					resource.TestCheckResourceAttrSet("data.spa_session_policy.by_name", "id"),
					// top-level scalar attributes
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "name", policyName),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "description", "DS by-name acceptance test"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "active", "false"),
					resource.TestCheckResourceAttrSet("data.spa_session_policy.by_name", "priority"),
					// rules list must be present
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "generic_rules.#", "1"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "generic_rules.0.name", "DS Rule"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "generic_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_name", "generic_rules.0.active", "true"),
					// data source must match the resource
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_name", "id",
						"spa_session_policy.ds_source", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_name", "name",
						"spa_session_policy.ds_source", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_name", "description",
						"spa_session_policy.ds_source", "description",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_name", "active",
						"spa_session_policy.ds_source", "active",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_name", "priority",
						"spa_session_policy.ds_source", "priority",
					),
				),
			},
		},
	})
}

// TestAccSessionPolicyDataSource_byID creates a session policy and reads it
// back by ID, asserting attribute parity with the resource.
func TestAccSessionPolicyDataSource_byID(t *testing.T) {
	policyName := "tf-acc-ds-session-policy-byid"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByNameForDS(policyName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSessionPolicyDataSourceByIDConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.ds_source"),
					resource.TestCheckResourceAttrSet("data.spa_session_policy.by_id", "id"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_id", "name", policyName),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_id", "description", "DS by-ID acceptance test"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_id", "active", "false"),
					resource.TestCheckResourceAttrSet("data.spa_session_policy.by_id", "priority"),
					resource.TestCheckResourceAttr("data.spa_session_policy.by_id", "generic_rules.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_id", "id",
						"spa_session_policy.ds_source", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_id", "name",
						"spa_session_policy.ds_source", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_id", "active",
						"spa_session_policy.ds_source", "active",
					),
					resource.TestCheckResourceAttrPair(
						"data.spa_session_policy.by_id", "priority",
						"spa_session_policy.ds_source", "priority",
					),
				),
			},
		},
	})
}

// TestAccSessionPolicyDataSource_notFound verifies that looking up a
// non-existent session policy produces an error.
func TestAccSessionPolicyDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSessionPolicyDataSourceNotFoundConfig(),
				ExpectError: regexp.MustCompile(`(?i)(not found|no session polic)`),
			},
		},
	})
}

// =============================================================================
// Tests — spa_session_policies data source
// =============================================================================

// TestAccSessionPoliciesDataSource_basic creates a session policy and confirms
// the list data source returns at least one item.
func TestAccSessionPoliciesDataSource_basic(t *testing.T) {
	policyName := "tf-acc-ds-session-policies-basic"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByNameForDS(policyName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSessionPoliciesDataSourceConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.list_source"),
					testAccCheckSessionPoliciesListNonEmpty("data.spa_session_policies.all"),
				),
			},
		},
	})
}

// TestAccSessionPoliciesDataSource_nameFilter creates a session policy and reads
// the list with a name filter; the result must contain exactly that policy.
func TestAccSessionPoliciesDataSource_nameFilter(t *testing.T) {
	policyName := "tf-acc-ds-session-policies-filter"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByNameForDS(policyName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSessionPoliciesDataSourceNameFilterConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.filter_source"),
					resource.TestCheckResourceAttr("data.spa_session_policies.filtered", "session_policies.#", "1"),
					resource.TestCheckResourceAttr("data.spa_session_policies.filtered", "session_policies.0.name", policyName),
					resource.TestCheckResourceAttr("data.spa_session_policies.filtered", "session_policies.0.description", "DS name-filter acceptance test"),
					resource.TestCheckResourceAttr("data.spa_session_policies.filtered", "session_policies.0.active", "false"),
					resource.TestCheckResourceAttrSet("data.spa_session_policies.filtered", "session_policies.0.id"),
					resource.TestCheckResourceAttrSet("data.spa_session_policies.filtered", "session_policies.0.priority"),
				),
			},
		},
	})
}

// TestAccSessionPoliciesDataSource_pagination creates a session policy and reads
// the list with explicit offset/limit pagination parameters.
func TestAccSessionPoliciesDataSource_pagination(t *testing.T) {
	policyName := "tf-acc-ds-session-policies-paged"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByNameForDS(policyName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSessionPoliciesDataSourcePaginationConfig(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.paged_source"),
					// With offset=0 and limit=5 we must get between 1 and 5 results
					testAccCheckSessionPoliciesCountBetween("data.spa_session_policies.paged", 1, 5),
				),
			},
		},
	})
}
// testAccCheckSessionPoliciesListNonEmpty verifies that the session_policies
// list returned by a spa_session_policies data source contains at least one item.
func testAccCheckSessionPoliciesListNonEmpty(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}
		var count int
		if _, err := fmt.Sscanf(rs.Primary.Attributes["session_policies.#"], "%d", &count); err != nil {
			return fmt.Errorf("could not parse session_policies.#: %w", err)
		}
		if count <= 0 {
			return fmt.Errorf("expected at least one session policy, got %d", count)
		}
		return nil
	}
}

// testAccCheckSessionPoliciesCountBetween verifies that the session_policies
// list count is within the inclusive range [min, max].
func testAccCheckSessionPoliciesCountBetween(dataSourceName string, min, max int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}
		var count int
		if _, err := fmt.Sscanf(rs.Primary.Attributes["session_policies.#"], "%d", &count); err != nil {
			return fmt.Errorf("could not parse session_policies.#: %w", err)
		}
		if count < min || count > max {
			return fmt.Errorf("expected session_policies count between %d and %d, got %d", min, max, count)
		}
		return nil
	}
}