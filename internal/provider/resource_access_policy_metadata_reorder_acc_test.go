package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccAccessPolicy_metadataReorderOmitted guards against positional metadata
// mis-association on reorder. It creates two TYPE_USERGROUP rules whose metadata
// value equals the rule's own values entry (rule A -> aaa, rule B -> bbb), then
// reorders the rules to B,A while OMITTING metadata (relying on the computed
// value). Metadata must be reconciled by content, not by list index, so after
// the reorder each rule still carries its OWN metadata.
func TestAccAccessPolicy_metadataReorderOmitted(t *testing.T) {
	name := "tf-acc-md-reorder-policy"
	fqdn := "tf-acc-md-reorder-app.example.com"
	rdResource := "test_domain_for_md_reorder"
	appResource := "test_app_for_md_reorder"

	sidA := "SID:/e2e/aaa-1111"
	sidB := "SID:/e2e/bbb-2222"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for metadata-reorder policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-app-for-md-reorder",
		appType:      "web",
		description:  "Test app for metadata-reorder policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	ruleWithMeta := func(rName, sid, mdKey string) string {
		return fmt.Sprintf(`
    {
      name        = %[1]q
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = [%[2]q]
          metadata   = { %[3]q = %[2]q }
        }
      ]
    }`, rName, sid, mdKey)
	}
	ruleNoMeta := func(rName, sid string) string {
		return fmt.Sprintf(`
    {
      name        = %[1]q
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = [%[2]q]
        }
      ]
    }`, rName, sid)
	}
	policy := func(rulesBody string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_md_reorder" {
  name        = %[1]q
  description = "Terraform acceptance test - metadata reorder"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [%[3]s
  ]
}
`, name, appResource, rulesBody)
	}

	// Step 1: rule A (values aaa, metadata alice=aaa), rule B (values bbb, metadata bob=bbb).
	configAB := rdConfig + appConfig + policy(
		ruleWithMeta("rule-a", sidA, "alice")+","+ruleWithMeta("rule-b", sidB, "bob"))
	// Step 2: reorder to B,A with metadata omitted (relies on computed/state value).
	configBA := rdConfig + appConfig + policy(
		ruleNoMeta("rule-b", sidB)+","+ruleNoMeta("rule-a", sidA))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: configAB,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_md_reorder"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_md_reorder", "access_rules.#", "2"),
					checkRuleMetadataMatchesOwnValues("citrixspa_access_policy.test_md_reorder", 2),
				),
			},
			{
				Config: configBA,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_md_reorder", "access_rules.#", "2"),
					checkRuleMetadataMatchesOwnValues("citrixspa_access_policy.test_md_reorder", 2),
				),
			},
			{
				Config:   configBA,
				PlanOnly: true,
			},
		},
	})
}

// checkRuleMetadataMatchesOwnValues fetches the policy from the API and asserts
// that, for every rule carrying values, the metadata value equals the rule's own
// values[0]. The test data is constructed so this invariant holds iff each rule
// kept its OWN metadata (no positional mis-association across the reorder). It
// FAILS (rather than skipping) when a rule that should carry metadata has none,
// so a regression that drops metadata during reconciliation is caught, and it
// verifies wantRules sub-rules were checked so a dropped rule cannot pass.
func checkRuleMetadataMatchesOwnValues(resourceName string, wantRules int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		client, err := testAccCreateClient()
		if err != nil {
			return err
		}
		pol, err := client.GetAccessPolicy(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}
		verified := 0
		for i, ar := range pol.AccessRules {
			for j, rl := range ar.Rules {
				if len(rl.Values) == 0 {
					continue
				}
				if len(rl.Metadata) == 0 {
					return fmt.Errorf("metadata unexpectedly empty: access_rules[%d].rules[%d] has values %v but no metadata (a reconciliation regression could drop it)", i, j, rl.Values)
				}
				own := rl.Values[0]
				for k, mv := range rl.Metadata {
					if got := fmt.Sprintf("%v", mv); got != own {
						return fmt.Errorf("metadata mis-association: access_rules[%d].rules[%d] has values[0]=%q but metadata %q=%q (belongs to a different rule)", i, j, own, k, got)
					}
				}
				verified++
			}
		}
		if verified != wantRules {
			return fmt.Errorf("expected %d metadata-carrying rules, verified %d (a rule may have been dropped)", wantRules, verified)
		}
		return nil
	}
}
