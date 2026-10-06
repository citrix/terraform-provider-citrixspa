package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccAccessPolicy_deferredReorderIDMetadataRetained verifies the "residual #1"
// case: when the WHOLE access_rules list is unknown at plan time (a rule value is
// wired to an unresolved output) AND the rules are reordered with metadata omitted,
// existing rules must keep their backend id and their OWN metadata. It forces the
// unknown by referencing a random_string that is replaced (keeper change) on the
// second step, so ModifyPlan defers and the reconciliation must be recovered in
// Update. Rules are keyed by their stable name in the API assertions because one
// rule's values legitimately change between steps.
func TestAccAccessPolicy_deferredReorderIDMetadataRetained(t *testing.T) {
	name := "tf-acc-deferred-reorder-policy"
	fqdn := "tf-acc-deferred-reorder-app.example.com"
	rdResource := "test_domain_for_deferred_reorder"
	appResource := "test_app_for_deferred_reorder"

	sidA := "SID:/e2e/aaa-1111"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for deferred-reorder policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-app-for-deferred-reorder",
		appType:      "web",
		description:  "Test app for deferred-reorder policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	// churn is referenced by rule-b's values, so when its keeper changes it is
	// replaced and its result is unknown at plan time -> access_rules is unknown.
	churn := func(keeper string) string {
		return fmt.Sprintf(`
resource "random_string" "churn" {
  length  = 12
  special = false
  keepers = { v = %[1]q }
}
`, keeper)
	}
	ruleAExplicit := fmt.Sprintf(`
    {
      name        = "rule-a"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = [%q], metadata = { alice = "aaa" } }]
    }`, sidA)
	ruleAOmitted := fmt.Sprintf(`
    {
      name        = "rule-a"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = [%q] }]
    }`, sidA)
	// rule-b's first value is the (plan-time-unknown) random result.
	ruleBExplicit := `
    {
      name        = "rule-b"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = ["SID:/e2e/${random_string.churn.result}"], metadata = { bob = "bbb" } }]
    }`
	ruleBOmitted := `
    {
      name        = "rule-b"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = ["SID:/e2e/${random_string.churn.result}"] }]
    }`
	policy := func(rulesBody string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_deferred_reorder" {
  name        = %[1]q
  description = "Terraform acceptance test - deferred reorder"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [%[3]s
  ]
}
`, name, appResource, rulesBody)
	}

	// Step 1: [rule-a, rule-b] explicit metadata, churn keeper "1".
	configAB := rdConfig + appConfig + churn("1") + policy(ruleAExplicit+","+ruleBExplicit)
	// Step 2: reorder to [rule-b, rule-a], metadata omitted, churn keeper "2"
	// (churn replaced -> rule-b value unknown at plan -> ModifyPlan defers).
	configBA := rdConfig + appConfig + churn("2") + policy(ruleBOmitted+","+ruleAOmitted)

	resourceName := "citrixspa_access_policy.test_deferred_reorder"

	fetchByName := func(s *terraform.State) (map[string]string, map[string]map[string]interface{}, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return nil, nil, fmt.Errorf("resource not found: %s", resourceName)
		}
		client, err := testAccCreateClient()
		if err != nil {
			return nil, nil, err
		}
		pol, err := client.GetAccessPolicy(context.Background(), rs.Primary.ID)
		if err != nil {
			return nil, nil, err
		}
		ids := map[string]string{}
		meta := map[string]map[string]interface{}{}
		for _, ar := range pol.AccessRules {
			ids[ar.Name] = ar.ID
			if len(ar.Rules) > 0 {
				meta[ar.Name] = ar.Rules[0].Metadata
			}
		}
		return ids, meta, nil
	}

	var idA, idB string
	recordIDs := func(s *terraform.State) error {
		ids, meta, err := fetchByName(s)
		if err != nil {
			return err
		}
		idA, idB = ids["rule-a"], ids["rule-b"]
		if idA == "" || idB == "" {
			return fmt.Errorf("failed to record baseline ids: rule-a=%q rule-b=%q", idA, idB)
		}
		if fmt.Sprintf("%v", meta["rule-a"]["alice"]) != "aaa" {
			return fmt.Errorf("rule-a baseline metadata wrong: %v", meta["rule-a"])
		}
		if fmt.Sprintf("%v", meta["rule-b"]["bob"]) != "bbb" {
			return fmt.Errorf("rule-b baseline metadata wrong: %v", meta["rule-b"])
		}
		return nil
	}
	assertRetained := func(s *terraform.State) error {
		ids, meta, err := fetchByName(s)
		if err != nil {
			return err
		}
		if ids["rule-a"] != idA {
			return fmt.Errorf("rule-a id churned on deferred reorder: was %q now %q", idA, ids["rule-a"])
		}
		if ids["rule-b"] != idB {
			return fmt.Errorf("rule-b id churned on deferred reorder: was %q now %q", idB, ids["rule-b"])
		}
		if fmt.Sprintf("%v", meta["rule-a"]["alice"]) != "aaa" {
			return fmt.Errorf("rule-a metadata lost/mis-associated on deferred reorder: %v", meta["rule-a"])
		}
		if fmt.Sprintf("%v", meta["rule-b"]["bob"]) != "bbb" {
			return fmt.Errorf("rule-b metadata lost/mis-associated on deferred reorder: %v", meta["rule-b"])
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {Source: "registry.terraform.io/hashicorp/random"},
		},
		CheckDestroy: testAccCheckAccessPolicyDestroy,
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
				Config: configBA,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_rules.#", "2"),
					assertRetained,
				),
			},
		},
	})
}
