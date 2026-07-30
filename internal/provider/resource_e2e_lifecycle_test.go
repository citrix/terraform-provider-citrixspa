package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// End-to-end multi-resource lifecycle test
// =============================================================================
//
// This test exercises a full dependency graph in a single config:
//
//	citrixspa_routing_domain (app FQDN)     ─┐
//	citrixspa_routing_domain (related FQDN) ─┼─► citrixspa_application (web, complete)
//	                                         ├─► citrixspa_access_policy   (references the app)
//	                                         └─► citrixspa_security_group  (references the app)
//
// A unique name prefix is generated per run so the test can run safely against
// a shared tenant without colliding with other resources.

// testAccE2ELifecycleConfig builds the full dependency-graph HCL for the given
// unique prefix, reusing the per-resource config builders.
func testAccE2ELifecycleConfig(prefix string) string {
	appFQDN := prefix + ".example.com"
	relatedFQDN := "api." + prefix + ".example.com"

	// A "complete" web application requires routing domains for both the app
	// URL and the related URL.
	rd := testAccRoutingDomainConfig(
		"e2e_rd", appFQDN, "internal", "web",
		"E2E lifecycle test - routing domain", "enabled", "false", "[]",
	)
	rdRelated := testAccRoutingDomainConfig(
		"e2e_rd_api", relatedFQDN, "internal", "web",
		"E2E lifecycle test - related routing domain", "enabled", "false", "[]",
	)

	app := testAccApplicationConfig(testAppConfig{
		resourceName: "e2e_app",
		name:         prefix + "-app",
		appType:      "web",
		description:  "E2E lifecycle test - web application",
		url:          "https://" + appFQDN,
		relatedURLs:  []string{relatedFQDN},
		state:        "complete",
		dependsOn:    []string{"citrixspa_routing_domain.e2e_rd", "citrixspa_routing_domain.e2e_rd_api"},
	})

	policy := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "e2e_policy",
		name:            prefix + "-policy",
		description:     "E2E lifecycle test - access policy",
		active:          false,
		appResourceName: "e2e_app",
		priority:        999,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	sg := testAccSecurityGroupConfig(testSecurityGroupConfig{
		resourceName:   "e2e_sg",
		name:           prefix + "-sg",
		appRefs:        []string{"citrixspa_application.e2e_app.id"},
		systemIn:       "enabled",
		systemOut:      "disabled",
		unpublishedIn:  "disabled",
		unpublishedOut: "disabled",
	})

	return rd + rdRelated + app + policy + sg
}

// testAccCheckE2ELifecycleDestroy verifies that every resource created by the
// lifecycle test has been removed from the backend after destroy.
func testAccCheckE2ELifecycleDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		id := rs.Primary.Attributes["id"]
		switch rs.Type {
		case "citrixspa_application":
			if _, err := client.GetApplication(ctx, id); err == nil {
				return fmt.Errorf("application %s still exists in the API after destroy", id)
			} else if !strings.Contains(err.Error(), "404") {
				return fmt.Errorf("unexpected error checking application %s: %s", id, err)
			}
		case "citrixspa_access_policy":
			if _, err := client.GetAccessPolicy(ctx, id); err == nil {
				return fmt.Errorf("access policy %s still exists in the API after destroy", id)
			} else if !strings.Contains(err.Error(), "404") {
				return fmt.Errorf("unexpected error checking access policy %s: %s", id, err)
			}
		case "citrixspa_security_group":
			if _, err := client.GetSecurityGroup(ctx, id); err == nil {
				return fmt.Errorf("security group %s still exists in the API after destroy", id)
			} else if !strings.Contains(err.Error(), "404") {
				return fmt.Errorf("unexpected error checking security group %s: %s", id, err)
			}
		case "citrixspa_routing_domain":
			fqdn := rs.Primary.Attributes["fqdn"]
			if _, err := client.GetRoutingDomain(ctx, fqdn); err == nil {
				return fmt.Errorf("routing domain %s still exists in the API after destroy", fqdn)
			} else if !strings.Contains(err.Error(), "404") {
				return fmt.Errorf("unexpected error checking routing domain %s: %s", fqdn, err)
			}
		}
	}
	return nil
}

// TestAccE2ELifecycle exercises create → idempotency → import across a
// multi-resource dependency graph, then relies on the framework to destroy and
// verify cleanup via testAccCheckE2ELifecycleDestroy.
func TestAccE2ELifecycle(t *testing.T) {
	prefix := fmt.Sprintf("tf-acc-e2e-%s", strings.ToLower(acctest.RandString(6)))
	appFQDN := prefix + ".example.com"
	relatedFQDN := "api." + prefix + ".example.com"
	config := testAccE2ELifecycleConfig(prefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			// Best-effort cleanup of leftovers from a previous failed run.
			testAccCleanupSecurityGroupByName(prefix + "-sg")
			testAccCleanupApplicationByName(prefix + "-app")
			testAccCleanupRoutingDomain(appFQDN)
			testAccCleanupRoutingDomain(relatedFQDN)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckE2ELifecycleDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create/apply the full dependency graph.
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.e2e_rd"),
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.e2e_rd_api"),
					testAccCheckApplicationExistsInAPI("citrixspa_application.e2e_app"),
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.e2e_policy"),
					testAccCheckSecurityGroupExistsInAPI("citrixspa_security_group.e2e_sg"),
					resource.TestCheckResourceAttr("citrixspa_application.e2e_app", "name", prefix+"-app"),
					resource.TestCheckResourceAttr("citrixspa_application.e2e_app", "type", "web"),
					resource.TestCheckResourceAttrSet("citrixspa_application.e2e_app", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.e2e_policy", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_security_group.e2e_sg", "id"),
					resource.TestCheckResourceAttr("citrixspa_security_group.e2e_sg", "app_ids.#", "1"),
				),
			},
			// Step 2: Idempotency — re-applying the same config yields an empty plan.
			{
				Config:   config,
				PlanOnly: true,
			},
			// Step 3: Import-verify at least one resource (the application).
			{
				ResourceName:      "citrixspa_application.e2e_app",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase.
		},
	})
}
