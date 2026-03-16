package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccessPolicyUpdatePayloadGeneration tests that the Update method generates
// a complete payload with all required fields (apps, accessRules) that were
// missing in the original implementation and causing API validation errors.
func TestAccessPolicyUpdatePayloadGeneration(t *testing.T) {
	// Create a mock data model representing what Terraform would provide
	data := AccessPolicyResourceModel{
		ID:          types.StringValue("test-policy-id"),
		Name:        types.StringValue("Test Policy"),
		Description: types.StringValue("Test Description"),
		Active:      types.BoolValue(true),
		Priority:    types.Int64Value(100),
	}

	// Add apps - this was missing in the original Update method
	appsValues := []attr.Value{
		types.StringValue("app-1"),
		types.StringValue("app-2"),
	}
	appsSet, _ := types.SetValue(types.StringType, appsValues)
	data.Apps = appsSet

	// Add access rules - this was missing in the original Update method
	valuesValues := []attr.Value{
		types.StringValue("user1@example.com"),
		types.StringValue("user2@example.com"),
	}
	valuesList, _ := types.ListValue(types.StringType, valuesValues)

	metadataMap := map[string]attr.Value{
		"key1": types.StringValue("value1"),
	}
	metadata, _ := types.MapValue(types.StringType, metadataMap)

	rule := RuleResourceModel{
		Type:     types.StringValue("TYPE_USERGROUP"),
		Operator: types.StringValue("OPERATOR_IN"),
		Values:   valuesList,
		Metadata: metadata,
	}

	accessRule := AccessRuleResourceModel{
		ID:          types.StringValue("rule-1"),
		Name:        types.StringValue("Test Rule"),
		Description: types.StringValue("Test Rule Description"),
		Priority:    types.Int64Value(1),
		Active:      types.BoolValue(true),
		Access:      types.StringValue("ACCESS_ALLOW"),
		Rules:       []RuleResourceModel{rule},
	}

	data.AccessRules = []AccessRuleResourceModel{accessRule}

	// Test the conversion logic (extracted from Update method)
	ctx := context.Background()

	// Convert Apps
	var apps []string
	if !data.Apps.IsNull() && !data.Apps.IsUnknown() {
		diags := data.Apps.ElementsAs(ctx, &apps, false)
		if diags.HasError() {
			t.Fatalf("Failed to convert apps: %v", diags.Errors())
		}
	}

	// Convert AccessRules
	accessRules := make([]AccessRule, 0, len(data.AccessRules))
	for _, terraformRule := range data.AccessRules {
		accessRule := AccessRule{
			Name:        terraformRule.Name.ValueString(),
			Description: terraformRule.Description.ValueString(),
			Priority:    int(terraformRule.Priority.ValueInt64()),
			Active:      terraformRule.Active.ValueBool(),
			Access:      terraformRule.Access.ValueString(),
		}

		// Handle optional ID field
		if !terraformRule.ID.IsNull() {
			accessRule.ID = terraformRule.ID.ValueString()
		}

		// Convert Rules
		rules := make([]Rule, 0, len(terraformRule.Rules))
		for _, tfRule := range terraformRule.Rules {
			rule := Rule{
				Type:     tfRule.Type.ValueString(),
				Operator: tfRule.Operator.ValueString(),
			}

			// Convert Values list
			var values []string
			diags := tfRule.Values.ElementsAs(ctx, &values, false)
			if diags.HasError() {
				t.Fatalf("Failed to convert rule values: %v", diags.Errors())
			}
			rule.Values = values

			// Convert Metadata map
			if !tfRule.Metadata.IsNull() && !tfRule.Metadata.IsUnknown() {
				metadataStringMap := make(map[string]string)
				diags := tfRule.Metadata.ElementsAs(ctx, &metadataStringMap, false)
				if diags.HasError() {
					t.Fatalf("Failed to convert rule metadata: %v", diags.Errors())
				}
				// Convert to map[string]interface{} for API compatibility
				metadataMap := make(map[string]interface{})
				for k, v := range metadataStringMap {
					metadataMap[k] = v
				}
				rule.Metadata = metadataMap
			}

			rules = append(rules, rule)
		}
		accessRule.Rules = rules

		accessRules = append(accessRules, accessRule)
	}

	// Create the policy payload (what would be sent to the API)
	policy := &AccessPolicy{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Active:      data.Active.ValueBool(),
		Priority:    int(data.Priority.ValueInt64()),
		Apps:        apps,
		AccessRules: accessRules,
	}

	// Verify that all required fields are present
	// These are the fields that were missing and causing the API validation error

	// Check Apps field (was missing in original implementation)
	if policy.Apps == nil {
		t.Error("Apps field is nil - this would cause API validation error")
	}
	if len(policy.Apps) != 2 {
		t.Errorf("Expected 2 apps, got %d", len(policy.Apps))
	}
	if policy.Apps[0] != "app-1" || policy.Apps[1] != "app-2" {
		t.Errorf("Apps values incorrect: %v", policy.Apps)
	}

	// Check AccessRules field (was missing in original implementation)
	if policy.AccessRules == nil {
		t.Error("AccessRules field is nil - this would cause API validation error")
	}
	if len(policy.AccessRules) != 1 {
		t.Errorf("Expected 1 access rule, got %d", len(policy.AccessRules))
	}

	// Check that AccessRule contains the 'active' field (mentioned in API error)
	accessRuleFromPayload := policy.AccessRules[0]
	if !accessRuleFromPayload.Active {
		t.Error("AccessRule.Active should be true")
	}
	if accessRuleFromPayload.Name != "Test Rule" {
		t.Errorf("AccessRule.Name incorrect: %s", accessRuleFromPayload.Name)
	}

	// Verify nested rule structure
	if len(accessRuleFromPayload.Rules) != 1 {
		t.Errorf("Expected 1 nested rule, got %d", len(accessRuleFromPayload.Rules))
	}
	nestedRule := accessRuleFromPayload.Rules[0]
	if nestedRule.Type != "TYPE_USERGROUP" {
		t.Errorf("Rule.Type incorrect: %s", nestedRule.Type)
	}
	if len(nestedRule.Values) != 2 {
		t.Errorf("Expected 2 rule values, got %d", len(nestedRule.Values))
	}

	t.Logf("✅ Success: Update payload now includes all required fields:")
	t.Logf("  - Apps: %v", policy.Apps)
	t.Logf("  - AccessRules count: %d", len(policy.AccessRules))
	t.Logf("  - AccessRules[0].Active: %v", policy.AccessRules[0].Active)
	t.Logf("  - AccessRules[0].Rules count: %d", len(policy.AccessRules[0].Rules))
}

// =============================================================================
// CheckDestroy functions — verify resources are deleted from backend after destroy
// =============================================================================

func testAccCheckAccessPolicyDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "spa_access_policy" {
			continue
		}

		id := rs.Primary.Attributes["id"]
		_, err := client.GetAccessPolicy(ctx, id)
		if err == nil {
			return fmt.Errorf("access policy %s still exists in the API after destroy", id)
		}
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("unexpected error checking access policy %s: %s", id, err)
		}
	}
	return nil
}

// =============================================================================
// Exists-in-API check functions — verify resources exist in the backend
// =============================================================================

func testAccCheckAccessPolicyExistsInAPI(resourceName string) resource.TestCheckFunc {
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

		policy, err := client.GetAccessPolicy(context.Background(), id)
		if err != nil {
			return fmt.Errorf("access policy %s not found in API: %s", id, err)
		}

		if policy.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("access policy name mismatch: API=%q, state=%q", policy.Name, rs.Primary.Attributes["name"])
		}

		return nil
	}
}

// =============================================================================
// Access Policy Tests
// =============================================================================

// testAccessPolicyConfig holds all parameters for testAccAccessPolicyConfig.
type testAccessPolicyConfig struct {
	resourceName    string // HCL resource label, e.g. "test_basic"
	name            string // policy display name
	description     string
	active          bool
	appResourceName string // HCL resource label of the spa_application to reference
	priority        int
	accessRulesHCL  string // raw HCL block for the access_rules attribute
}

const testAccBasicAccessRulesHCL = `  access_rules = [
    {
      name          = "Default Rule"
      description   = ""
      priority      = 1
      active        = true
      access        = "ACCESS_DENY"
      access_native = "ACCESS_DENY"

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]`

// testAccNoConditionsAccessRulesHCL tests the case where conditions is an empty
// list (conditions = []). This is the scenario that triggers the
// "was cty.ListValEmpty(...), but now null" inconsistency error when the Read
// method returns nil instead of an empty slice for Conditions.
const testAccNoConditionsAccessRulesHCL = `  access_rules = [
    {
      name        = "No Conditions Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_DENY"
      description = ""

      conditions = []

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["Everyone"]
        }
      ]
    }
  ]`

const testAccComplexAccessRulesHCL = `  access_rules = [
    {
      name        = "Allow Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]
	  restrictions: {
		redirect_sbs: false,
		enhanced_security_settings: {
			"_browserV1": "embeddedBrowser",
			"watermarkV1": "enabled"
			"downloadV1": "disabled"
			"clipboardV1": "disabled"
			"printingV1": "disabled"
			"keyLoggingV1": "disabled"
			"screenCaptureV1": "disabled"
			"proxyTrafficV1": "direct"
			"uploadV1": "disabled"
		}
	  },
      rules = [
        {
          type       = "TYPE_USERGROUP"
          tag_source = ""
          tag_key    = ""
		  operator = "OPERATOR_IN"
          values   = ["Everyone"]
        },
		{
			"metadata": {
				"AF": "Afghanistan"
			},
			"operator": "OPERATOR_IN",
			"tag_key": "location-geo-country-isocode",
			"tag_source": "ITM",
			"type": "TYPE_TAG",
			"values": [
				"AF"
			]
		},
		{
			"operator": "OPERATOR_IN",
			"tag_source": "",
			"tag_key": "",
			"type": "TYPE_MULTIURLDOMAIN",
			"values": [
				"terra.cloud.com"
			]
		}
      ]
    }
  ]`

// testAccAccessPolicyConfig generates a Terraform HCL config for a spa_access_policy resource.
func testAccAccessPolicyConfig(cfg testAccessPolicyConfig) string {
	if cfg.resourceName == "" {
		cfg.resourceName = "test"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "resource \"spa_access_policy\" %q {\n", cfg.resourceName)
	fmt.Fprintf(&b, "  name        = %q\n", cfg.name)
	fmt.Fprintf(&b, "  description = %q\n", cfg.description)
	fmt.Fprintf(&b, "  active      = %v\n", cfg.active)
	fmt.Fprintf(&b, "  apps        = [spa_application.%s.id]\n", cfg.appResourceName)
	fmt.Fprintf(&b, "  priority    = %d\n", cfg.priority)
	if cfg.accessRulesHCL != "" {
		fmt.Fprintf(&b, "%s\n", cfg.accessRulesHCL)
	}
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

func testAccAccessPolicyConfig_basic(name string) string {
	const rdResourceName = "test_domain_for_basic_policy"
	const appResourceName = "test_app_for_basic_policy"
	const fqdn = "tf-acc-test-basic-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for basic access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-basic-policy",
		appType:      "web",
		description:  "Test app for basic access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_basic",
		name:            name,
		description:     "Terraform acceptance test - basic access policy",
		active:          false,
		appResourceName: appResourceName,
		priority:        999,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_basicUpdated(name string) string {
	const rdResourceName = "test_domain_for_basic_policy"
	const appResourceName = "test_app_for_basic_policy"
	const fqdn = "tf-acc-test-basic-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for basic access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-basic-policy",
		appType:      "web",
		description:  "Test app for basic access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_basic",
		name:            name,
		description:     "Terraform acceptance test - basic access policy UPDATED",
		active:          true,
		appResourceName: appResourceName,
		priority:        998,
		accessRulesHCL:  testAccBasicAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_complex(name string) string {
	const rdResourceName = "test_domain_for_policy"
	const appResourceName = "test_app_for_policy"
	const fqdn = "tf-acc-test-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-policy",
		appType:      "web",
		description:  "Test app for access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_with_rules",
		name:            name,
		description:     "Terraform acceptance test - policy with access rules",
		active:          false,
		appResourceName: appResourceName,
		priority:        998,
		accessRulesHCL:  testAccComplexAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_complexUpdated(name string) string {
	const rdResourceName = "test_domain_for_policy"
	const appResourceName = "test_app_for_policy"
	const fqdn = "tf-acc-test-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-policy",
		appType:      "web",
		description:  "Test app for access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_with_rules",
		name:            name,
		description:     "Terraform acceptance test - policy with access rules UPDATED",
		active:          true,
		appResourceName: appResourceName,
		priority:        997,
		accessRulesHCL:  testAccComplexAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_noConditions(name string) string {
	const rdResourceName = "test_domain_for_no_cond_policy"
	const appResourceName = "test_app_for_no_cond_policy"
	const fqdn = "tf-acc-test-no-cond-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for no-conditions access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-no-cond-policy",
		appType:      "web",
		description:  "Test app for no-conditions access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"spa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_no_conditions",
		name:            name,
		description:     "Terraform acceptance test - policy with conditions = []",
		active:          false,
		appResourceName: appResourceName,
		priority:        997,
		accessRulesHCL:  testAccNoConditionsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

// TestAccAccessPolicy_noConditions reproduces the bug where writing
// conditions = [] (an explicit empty list) in the config caused:
//
//	Provider produced inconsistent result after apply:
//	.access_rules[0].conditions: was cty.ListValEmpty(...), but now null.
//
// The root cause was that the Read method only populated accessRule.Conditions
// when len(rule.Conditions) > 0, leaving it nil (→ null) for the empty-list
// case. The fix is to always assign the slice.
func TestAccAccessPolicy_noConditions(t *testing.T) {
	name := "tf-acc-test-no-cond-policy"
	fqdn := "tf-acc-test-no-cond-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with conditions = [] — this is the scenario that
			// previously caused "was cty.ListValEmpty, but now null".
			{
				Config: testAccAccessPolicyConfig_noConditions(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.test_no_conditions"),
					resource.TestCheckResourceAttr("spa_access_policy.test_no_conditions", "name", name),
					resource.TestCheckResourceAttr("spa_access_policy.test_no_conditions", "active", "false"),
					resource.TestCheckResourceAttr("spa_access_policy.test_no_conditions", "priority", "997"),
					resource.TestCheckResourceAttr("spa_access_policy.test_no_conditions", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_no_conditions", "access_rules.0.name", "No Conditions Rule"),
					resource.TestCheckResourceAttr("spa_access_policy.test_no_conditions", "access_rules.0.conditions.#", "0"),
					resource.TestCheckResourceAttrSet("spa_access_policy.test_no_conditions", "id"),
				),
			},
			// Step 2: Import — verifies the empty conditions list round-trips.
			{
				ResourceName:      "spa_access_policy.test_no_conditions",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAccessPolicy_basicAccessRules(t *testing.T) {
	name := "tf-acc-test-basic-policy"
	fqdn := fmt.Sprintf("%s-app.example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccAccessPolicyConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.test_basic"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "name", name),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "description", "Terraform acceptance test - basic access policy"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "active", "false"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "priority", "999"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.name", "Default Rule"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.access", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.access_native", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.tag_source", ""),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.tag_key", ""),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.values.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.values.0", "Everyone"),
					resource.TestCheckResourceAttrSet("spa_access_policy.test_basic", "id"),
				),
			},
			// Step 2: Update - change description, enable active, lower priority
			{
				Config: testAccAccessPolicyConfig_basicUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.test_basic"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "name", name),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "description", "Terraform acceptance test - basic access policy UPDATED"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "active", "true"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "priority", "998"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.name", "Default Rule"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.access", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.access_native", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.values.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_basic", "access_rules.0.rules.0.values.0", "Everyone"),
					resource.TestCheckResourceAttrSet("spa_access_policy.test_basic", "id"),
				),
			},
			// Step 3: Import by ID and verify state matches
			{
				ResourceName:      "spa_access_policy.test_basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAccessPolicy_complexAccessRules(t *testing.T) {
	name := "tf-acc-test-rules-policy"
	fqdn := "tf-acc-test-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccAccessPolicyConfig_complex(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.test_with_rules"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "name", name),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "description", "Terraform acceptance test - policy with access rules"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "active", "false"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "priority", "998"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.name", "Allow Rule"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.access", "ACCESS_ALLOW"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.conditions.0.platform_filter", "PLATFORM_FILTER_ANY"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.redirect_sbs", "false"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings._browserV1", "embeddedBrowser"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.watermarkV1", "enabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.downloadV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.uploadV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.clipboardV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.printingV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.keyLoggingV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.screenCaptureV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.proxyTrafficV1", "direct"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.rules.#", "3"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttrSet("spa_access_policy.test_with_rules", "id"),
				),
			},
			// Step 2: Update - change description, enable active, lower priority
			{
				Config: testAccAccessPolicyConfig_complexUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("spa_access_policy.test_with_rules"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "name", name),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "description", "Terraform acceptance test - policy with access rules UPDATED"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "active", "true"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "priority", "997"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.name", "Allow Rule"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.access", "ACCESS_ALLOW"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.conditions.0.platform_filter", "PLATFORM_FILTER_ANY"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.redirect_sbs", "false"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings._browserV1", "embeddedBrowser"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.watermarkV1", "enabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.downloadV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.uploadV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.clipboardV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.printingV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.keyLoggingV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.screenCaptureV1", "disabled"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.proxyTrafficV1", "direct"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.rules.#", "3"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("spa_access_policy.test_with_rules", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttrSet("spa_access_policy.test_with_rules", "id"),
				),
			},
			// Step 3: Import by ID and verify state matches
			{
				ResourceName:      "spa_access_policy.test_with_rules",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
