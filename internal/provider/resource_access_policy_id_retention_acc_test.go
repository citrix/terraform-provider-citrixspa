package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccAccessPolicy_ruleIDRetainedOnInsert guards against backend rule-identity
// churn when a rule is inserted at the front of the list. The SPA backend honors
// whatever access-rule id it is sent, so filling an omitted id from the rule at the
// same list index (positional) reassigns existing identities on an insert. This test
// creates two rules A and B, records their backend ids, then inserts C at the front
// with ids omitted, and asserts A and B keep their ORIGINAL ids while C gets a new one.
func TestAccAccessPolicy_ruleIDRetainedOnInsert(t *testing.T) {
	name := "tf-acc-id-retention-policy"
	fqdn := "tf-acc-id-retention-app.example.com"
	rdResource := "test_domain_for_id_retention"
	appResource := "test_app_for_id_retention"

	sidA := "SID:/e2e/aaa-1111"
	sidB := "SID:/e2e/bbb-2222"
	sidC := "SID:/e2e/ccc-3333"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for id-retention policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-app-for-id-retention",
		appType:      "web",
		description:  "Test app for id-retention policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	rule := func(rName, sid string) string {
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
resource "citrixspa_access_policy" "test_id_retention" {
  name        = %[1]q
  description = "Terraform acceptance test - id retention"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [%[3]s
  ]
}
`, name, appResource, rulesBody)
	}

	configAB := rdConfig + appConfig + policy(rule("rule-a", sidA)+","+rule("rule-b", sidB))
	configCAB := rdConfig + appConfig + policy(
		rule("rule-c", sidC)+","+rule("rule-a", sidA)+","+rule("rule-b", sidB))

	resourceName := "citrixspa_access_policy.test_id_retention"

	var idA, idB string
	recordIDs := func(s *terraform.State) error {
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
		for _, ar := range pol.AccessRules {
			if len(ar.Rules) == 0 || len(ar.Rules[0].Values) == 0 {
				continue
			}
			switch ar.Rules[0].Values[0] {
			case sidA:
				idA = ar.ID
			case sidB:
				idB = ar.ID
			}
		}
		if idA == "" || idB == "" {
			return fmt.Errorf("failed to record baseline ids: idA=%q idB=%q", idA, idB)
		}
		return nil
	}
	assertRetained := func(s *terraform.State) error {
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
		got := map[string]string{}
		for _, ar := range pol.AccessRules {
			if len(ar.Rules) > 0 && len(ar.Rules[0].Values) > 0 {
				got[ar.Rules[0].Values[0]] = ar.ID
			}
		}
		if got[sidA] != idA {
			return fmt.Errorf("rule A id churned on insert: was %q now %q", idA, got[sidA])
		}
		if got[sidB] != idB {
			return fmt.Errorf("rule B id churned on insert: was %q now %q", idB, got[sidB])
		}
		if got[sidC] == "" || got[sidC] == idA || got[sidC] == idB {
			return fmt.Errorf("inserted rule C did not get a fresh id: %q (idA=%q idB=%q)", got[sidC], idA, idB)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: configAB,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_rules.#", "2"),
					recordIDs,
				),
			},
			{
				Config: configCAB,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_rules.#", "3"),
					assertRetained,
				),
			},
			{
				Config:   configCAB,
				PlanOnly: true,
			},
		},
	})
}
