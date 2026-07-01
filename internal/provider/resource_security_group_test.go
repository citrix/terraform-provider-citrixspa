package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Security Group Resource Tests
// =============================================================================

// testSecurityGroupConfig holds parameters for testAccSecurityGroupConfig.
type testSecurityGroupConfig struct {
	resourceName   string   // HCL resource label
	name           string   // security group display name
	appRefs        []string // HCL references for app_ids, e.g. "spa_application.app1.id"
	systemIn       string   // "enabled" or "disabled"
	systemOut      string
	unpublishedIn  string
	unpublishedOut string
}

// testAccSecurityGroupConfig generates HCL for a spa_security_group resource.
func testAccSecurityGroupConfig(cfg testSecurityGroupConfig) string {
	if cfg.resourceName == "" {
		cfg.resourceName = "test"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "resource \"spa_security_group\" %q {\n", cfg.resourceName)
	fmt.Fprintf(&b, "  name = %q\n", cfg.name)

	// app_ids
	fmt.Fprintf(&b, "  app_ids = [%s]\n", strings.Join(cfg.appRefs, ", "))

	// system attribute
	fmt.Fprintf(&b, "  system = {\n")
	fmt.Fprintf(&b, "    data_in  = %q\n", cfg.systemIn)
	fmt.Fprintf(&b, "    data_out = %q\n", cfg.systemOut)
	fmt.Fprintf(&b, "  }\n")

	// unpublished_app attribute
	fmt.Fprintf(&b, "  unpublished_app = {\n")
	fmt.Fprintf(&b, "    data_in  = %q\n", cfg.unpublishedIn)
	fmt.Fprintf(&b, "    data_out = %q\n", cfg.unpublishedOut)
	fmt.Fprintf(&b, "  }\n")

	fmt.Fprintf(&b, "}\n")
	return b.String()
}

// testAccSecurityGroupPrereqAppConfig generates HCL for a prerequisite routing
// domain and complete web application that can be assigned to a security group.
// Two routing domains are created: one for the app URL and one for the related
// URL, both required by the API for a complete web application.
func testAccSecurityGroupPrereqAppConfig(resourceName, appName string) string {
	fqdn := appName + ".example.com"
	relatedFQDN := "api." + fqdn
	rdResourceName := resourceName + "_rd"
	rdRelatedResourceName := resourceName + "_rd_api"
	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Security group acceptance test", "enabled", "false", "[]",
	)
	rdRelatedConfig := testAccRoutingDomainConfig(
		rdRelatedResourceName, relatedFQDN, "internal", "web",
		"Security group acceptance test", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: resourceName,
		name:         appName,
		appType:      "web",
		description:  "Prerequisite app for security group acceptance test",
		url:          "https://" + fqdn,
		relatedURLs:  []string{relatedFQDN},
		state:        "complete",
		dependsOn:    []string{"spa_routing_domain." + rdResourceName, "spa_routing_domain." + rdRelatedResourceName},
	})
	return rdConfig + rdRelatedConfig + appConfig
}

// testAccCheckSecurityGroupDestroy verifies that all spa_security_group
// resources have been removed from the API after a test run.
func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "spa_security_group" {
			continue
		}
		id := rs.Primary.Attributes["id"]
		_, err := client.GetSecurityGroup(ctx, id)
		if err == nil {
			return fmt.Errorf("security group %s still exists in the API after destroy", id)
		}
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("unexpected error checking security group %s: %s", id, err)
		}
	}
	return nil
}

// testAccCheckSecurityGroupExistsInAPI verifies the security group exists in
// the API and its name matches the Terraform state.
func testAccCheckSecurityGroupExistsInAPI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}

		id := rs.Primary.Attributes["id"]
		if id == "" {
			return fmt.Errorf("no ID set for resource %s", resourceName)
		}

		client, err := testAccCreateClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		sg, err := client.GetSecurityGroup(context.Background(), id)
		if err != nil {
			return fmt.Errorf("security group %s not found in API: %s", id, err)
		}

		if sg.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("security group name mismatch: API=%q, state=%q", sg.Name, rs.Primary.Attributes["name"])
		}

		return nil
	}
}

// testAccCleanupSecurityGroupByName removes a security group by name to avoid
// collisions from previously failed test runs.
func testAccCleanupSecurityGroupByName(name string) {
	client, err := testAccCreateClient()
	if err != nil {
		return
	}
	ctx := context.Background()

	result, err := client.GetSecurityGroups(ctx, 0, -1)
	if err != nil {
		return
	}
	for _, sg := range result.SecurityGroups {
		if sg.Name == name {
			_ = client.DeleteSecurityGroup(ctx, sg.ID)
			return
		}
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccSecurityGroup_basic exercises the full CRUD lifecycle: create, update,
// import, and automatic destroy of a security group.
func TestAccSecurityGroup_basic(t *testing.T) {
	sgName := "tf-acc-test-sg-basic"
	appName := "tf-acc-test-sg-basic-app"

	appHCL := testAccSecurityGroupPrereqAppConfig("sg_app", appName)
	resAddr := "spa_security_group.test_sg"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSecurityGroupByName(sgName)
			testAccCleanupSecurityGroupByName(sgName + "-updated")
			testAccCleanupApplicationByName(appName)
			testAccCleanupRoutingDomain(appName + ".example.com")
			testAccCleanupRoutingDomain("api." + appName + ".example.com")
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckSecurityGroupDestroy,
			testAccCheckApplicationDestroy,
			testAccCheckRoutingDomainDestroy,
		),
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: appHCL + testAccSecurityGroupConfig(testSecurityGroupConfig{
					resourceName:   "test_sg",
					name:           sgName,
					appRefs:        []string{"spa_application.sg_app.id"},
					systemIn:       "enabled",
					systemOut:      "disabled",
					unpublishedIn:  "disabled",
					unpublishedOut: "disabled",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI(resAddr),
					resource.TestCheckResourceAttr(resAddr, "name", sgName),
					resource.TestCheckResourceAttr(resAddr, "app_ids.#", "1"),
					resource.TestCheckResourceAttr(resAddr, "system.data_in", "enabled"),
					resource.TestCheckResourceAttr(resAddr, "system.data_out", "disabled"),
					resource.TestCheckResourceAttr(resAddr, "unpublished_app.data_in", "disabled"),
					resource.TestCheckResourceAttr(resAddr, "unpublished_app.data_out", "disabled"),
					resource.TestCheckResourceAttrSet(resAddr, "id"),
					resource.TestCheckResourceAttrSet(resAddr, "modified"),
				),
			},
			// Step 2: Update — change name and flip system.data_out
			{
				Config: appHCL + testAccSecurityGroupConfig(testSecurityGroupConfig{
					resourceName:   "test_sg",
					name:           sgName + "-updated",
					appRefs:        []string{"spa_application.sg_app.id"},
					systemIn:       "enabled",
					systemOut:      "enabled",
					unpublishedIn:  "disabled",
					unpublishedOut: "disabled",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI(resAddr),
					resource.TestCheckResourceAttr(resAddr, "name", sgName+"-updated"),
					resource.TestCheckResourceAttr(resAddr, "system.data_in", "enabled"),
					resource.TestCheckResourceAttr(resAddr, "system.data_out", "enabled"),
					resource.TestCheckResourceAttr(resAddr, "unpublished_app.data_in", "disabled"),
					resource.TestCheckResourceAttr(resAddr, "unpublished_app.data_out", "disabled"),
					resource.TestCheckResourceAttrSet(resAddr, "id"),
				),
			},
			// Step 3: ImportState
			{
				ResourceName:      resAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccSecurityGroup_updateApps creates a security group with one app, then
// updates it to include a second app, verifying the app_ids set grows.
func TestAccSecurityGroup_updateApps(t *testing.T) {
	sgName := "tf-acc-test-sg-apps"
	app1Name := "tf-acc-test-sg-apps-app1"
	app2Name := "tf-acc-test-sg-apps-app2"

	app1HCL := testAccSecurityGroupPrereqAppConfig("sg_app1", app1Name)
	app2HCL := testAccSecurityGroupPrereqAppConfig("sg_app2", app2Name)
	resAddr := "spa_security_group.test_sg"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSecurityGroupByName(sgName)
			testAccCleanupApplicationByName(app1Name)
			testAccCleanupApplicationByName(app2Name)
			testAccCleanupRoutingDomain(app1Name + ".example.com")
			testAccCleanupRoutingDomain("api." + app1Name + ".example.com")
			testAccCleanupRoutingDomain(app2Name + ".example.com")
			testAccCleanupRoutingDomain("api." + app2Name + ".example.com")
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckSecurityGroupDestroy,
			testAccCheckApplicationDestroy,
			testAccCheckRoutingDomainDestroy,
		),
		Steps: []resource.TestStep{
			// Step 1: Create with one app
			{
				Config: app1HCL + app2HCL + testAccSecurityGroupConfig(testSecurityGroupConfig{
					resourceName:   "test_sg",
					name:           sgName,
					appRefs:        []string{"spa_application.sg_app1.id"},
					systemIn:       "enabled",
					systemOut:      "enabled",
					unpublishedIn:  "enabled",
					unpublishedOut: "enabled",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI(resAddr),
					resource.TestCheckResourceAttr(resAddr, "name", sgName),
					resource.TestCheckResourceAttr(resAddr, "app_ids.#", "1"),
				),
			},
			// Step 2: Update to include both apps
			{
				Config: app1HCL + app2HCL + testAccSecurityGroupConfig(testSecurityGroupConfig{
					resourceName:   "test_sg",
					name:           sgName,
					appRefs:        []string{"spa_application.sg_app1.id", "spa_application.sg_app2.id"},
					systemIn:       "enabled",
					systemOut:      "enabled",
					unpublishedIn:  "enabled",
					unpublishedOut: "enabled",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExistsInAPI(resAddr),
					resource.TestCheckResourceAttr(resAddr, "name", sgName),
					resource.TestCheckResourceAttr(resAddr, "app_ids.#", "2"),
				),
			},
		},
	})
}
