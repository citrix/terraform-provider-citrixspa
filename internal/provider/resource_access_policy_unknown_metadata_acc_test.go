package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccAccessPolicy_configuredUnknownMetadataNotAdoptedFromPrior verifies that a
// rule whose metadata map is EXPLICITLY configured but wholly unknown at plan time
// (wired to a computed value that resolves at apply) keeps the requested new value
// instead of silently adopting prior-state metadata.
//
// A wholly-unknown metadata map does not force ValidateConfig to defer, so it
// reaches the known reconciliation path in ModifyPlan. If that path treated the
// unknown map as an omission and copied prior state into the plan, apply would
// then produce a different value and Terraform would fail with "Provider produced
// inconsistent result after apply" (or, worse, silently keep the stale metadata).
// The metadata source is terraform_data.output, which is known-after-apply on the
// step that introduces it, yielding a wholly-unknown map at plan time.
func TestAccAccessPolicy_configuredUnknownMetadataNotAdoptedFromPrior(t *testing.T) {
	name := "tf-acc-unknown-meta-policy"
	fqdn := "tf-acc-unknown-meta-app.example.com"
	rdResource := "test_domain_for_unknown_meta"
	appResource := "test_app_for_unknown_meta"
	sid := "SID:/e2e/meta-2222"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for unknown-metadata policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-app-for-unknown-meta",
		appType:      "web",
		description:  "Test app for unknown-metadata policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	policy := func(rulesBody string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_unknown_meta" {
  name        = %[1]q
  description = "Terraform acceptance test - unknown metadata"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [%[3]s
  ]
}
`, name, appResource, rulesBody)
	}

	// Step 1: metadata is an explicit literal (known at plan).
	ruleLiteral := fmt.Sprintf(`
    {
      name        = "rule-x"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = [%q], metadata = { "display-name" = "v1" } }]
    }`, sid)

	// Step 2: metadata is sourced from a computed value that is unknown at plan
	// but resolves to a DIFFERENT value ("v2") at apply. The rule keeps the same
	// name/values so it matches its prior state (whose metadata is still "v1").
	metaSource := `
resource "terraform_data" "meta" {
  input = { "display-name" = "v2" }
}
`
	ruleComputed := fmt.Sprintf(`
    {
      name        = "rule-x"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = [%q], metadata = terraform_data.meta.output }]
    }`, sid)

	configV1 := rdConfig + appConfig + policy(ruleLiteral)
	configV2 := rdConfig + appConfig + metaSource + policy(ruleComputed)

	resourceName := "citrixspa_access_policy.test_unknown_meta"

	metaOf := func(s *terraform.State) (map[string]interface{}, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return nil, fmt.Errorf("resource not found: %s", resourceName)
		}
		client, err := testAccCreateClient()
		if err != nil {
			return nil, err
		}
		pol, err := client.GetAccessPolicy(context.Background(), rs.Primary.ID)
		if err != nil {
			return nil, err
		}
		for _, ar := range pol.AccessRules {
			if ar.Name == "rule-x" && len(ar.Rules) > 0 {
				return ar.Rules[0].Metadata, nil
			}
		}
		return nil, fmt.Errorf("rule-x not found in policy")
	}

	assertMeta := func(want string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			m, err := metaOf(s)
			if err != nil {
				return err
			}
			if got := fmt.Sprintf("%v", m["display-name"]); got != want {
				return fmt.Errorf("rule-x metadata display-name = %q, want %q (full: %v)", got, want, m)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: configV1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI(resourceName),
					assertMeta("v1"),
				),
			},
			{
				// With the bug, ModifyPlan copies prior ("v1") into the plan for the
				// unknown map, and apply then materializes "v2" -> inconsistent result.
				// With the fix, the plan keeps the map unknown and apply stores "v2".
				Config: configV2,
				Check:  assertMeta("v2"),
			},
		},
	})
}
