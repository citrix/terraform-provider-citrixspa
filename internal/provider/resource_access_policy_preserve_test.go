package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Regression: verify a rule's name and priority are preserved when set then
// omitted on a later update, and that content-matching survives a reorder.
func TestAccAccessPolicy_ruleNamePriorityPreservedOnUpdate(t *testing.T) {
	name := "tf-acc-tmp-namepri-policy"
	fqdn := "tf-acc-tmp-namepri-app.example.com"
	rdResource := "tmp_rd_namepri"
	appResource := "tmp_app_namepri"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"tmp namepri rd", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-tmp-namepri-app",
		appType:      "web",
		description:  "tmp namepri app",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	ruleA := func(name, prio string) string {
		n := ""
		if name != "" {
			n = "name        = " + fmt.Sprintf("%q", name) + "\n      "
		}
		p := ""
		if prio != "" {
			p = "priority    = " + prio + "\n      "
		}
		return fmt.Sprintf(`    {
      %s%saccess      = "ACCESS_ALLOW"
      active      = true
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = ["GroupA"] }]
    }`, n, p)
	}
	ruleB := func(name, prio string) string {
		n := ""
		if name != "" {
			n = "name        = " + fmt.Sprintf("%q", name) + "\n      "
		}
		p := ""
		if prio != "" {
			p = "priority    = " + prio + "\n      "
		}
		return fmt.Sprintf(`    {
      %s%saccess      = "ACCESS_DENY"
      active      = true
      description = ""
      rules = [{ type = "TYPE_USERGROUP", operator = "OPERATOR_IN", tag_source = "", tag_key = "", values = ["GroupB"] }]
    }`, n, p)
	}
	policy := func(desc, active, rulesBlock string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "tmp_namepri" {
  name        = %[1]q
  description = %[2]q
  active      = %[3]s
  apps        = [citrixspa_application.%[4]s.id]
  access_rules = [
%[5]s
  ]
}
`, name, desc, active, appResource, rulesBlock)
	}

	// Step 1: A(name=RuleA-Name,prio=5), B(name=RuleB-Name,prio=9) explicit.
	explicit := policy("explicit", "false",
		ruleA("RuleA-Name", "5")+",\n"+ruleB("RuleB-Name", "9"))
	// Step 2: same order, name+priority OMITTED, sibling (policy active) toggled.
	omitted := policy("omitted", "true",
		ruleA("", "")+",\n"+ruleB("", ""))
	// Step 3: REORDER (B first, A second), still omitted.
	reordered := policy("reordered", "true",
		ruleB("", "")+",\n"+ruleA("", ""))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: rdConfig + appConfig + explicit,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.0.name", "RuleA-Name"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.0.priority", "5"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.1.name", "RuleB-Name"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.1.priority", "9"),
				),
			},
			{
				// omitted name+priority must be PRESERVED from prior state.
				Config: rdConfig + appConfig + omitted,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.0.name", "RuleA-Name"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.0.priority", "5"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.1.name", "RuleB-Name"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.1.priority", "9"),
				),
			},
			{
				// Reorder: B now index 0, A index 1 — content-match must carry the
				// right rule's name/priority to the new position.
				Config: rdConfig + appConfig + reordered,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.0.name", "RuleB-Name"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.0.priority", "9"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.1.name", "RuleA-Name"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.tmp_namepri", "access_rules.1.priority", "5"),
				),
			},
		},
	})
}
