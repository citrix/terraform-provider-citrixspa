package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Unit tests — no API access required
// =============================================================================

// TestSessionPolicyFromModel_basic verifies that sessionPolicyFromModel produces
// the correct API struct from a minimal Terraform model.
func TestSessionPolicyFromModel_basic(t *testing.T) {
	ctx := context.Background()

	valuesVals := []attr.Value{types.StringValue("Everyone")}
	valuesList, d := types.ListValue(types.StringType, valuesVals)
	if d.HasError() {
		t.Fatalf("failed to build values list: %v", d.Errors())
	}

	data := SessionPolicyResourceModel{
		ID:          types.StringValue(""),
		Name:        types.StringValue("test-policy"),
		Description: types.StringValue("a description"),
		Active:      types.BoolValue(false),
		Priority:    types.Int64Value(100),
		Rules: []SessionPolicyRuleModel{
			{
				ID:          types.StringNull(),
				Name:        types.StringValue("Rule 1"),
				Description: types.StringValue(""),
				Priority:    types.Int64Value(1),
				Active:      types.BoolValue(true),
				Actions:     nil,
				Conditions: []SessionPolicyConditionModel{
					{
						Type:      types.StringValue("TYPE_USERGROUP"),
						Operator:  types.StringValue("OPERATOR_IN"),
						TagSource: types.StringValue(""),
						TagKey:    types.StringValue(""),
						Values:    valuesList,
						Metadata:  types.MapNull(types.StringType),
					},
				},
			},
		},
	}

	var diags diag.Diagnostics
	policy := sessionPolicyFromModel(ctx, &data, &diags)
	if diags.HasError() {
		t.Fatalf("sessionPolicyFromModel returned errors: %v", diags.Errors())
	}

	if policy.Name != "test-policy" {
		t.Errorf("Name mismatch: got %q, want %q", policy.Name, "test-policy")
	}
	if policy.Description != "a description" {
		t.Errorf("Description mismatch: got %q, want %q", policy.Description, "a description")
	}
	if policy.Active != false {
		t.Errorf("Active mismatch: got %v, want false", policy.Active)
	}
	if policy.Priority == nil || *policy.Priority != 100 {
		t.Errorf("Priority mismatch: got %v, want 100", policy.Priority)
	}
	if len(policy.GenericRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(policy.GenericRules))
	}

	rule := policy.GenericRules[0]
	if rule.Name != "Rule 1" {
		t.Errorf("Rule.Name mismatch: got %q, want %q", rule.Name, "Rule 1")
	}
	if rule.Priority != 1 {
		t.Errorf("Rule.Priority mismatch: got %d, want 1", rule.Priority)
	}
	if rule.Active != true {
		t.Errorf("Rule.Active mismatch: got %v, want true", rule.Active)
	}
	if len(rule.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(rule.Conditions))
	}

	cond := rule.Conditions[0]
	if cond.Type != "TYPE_USERGROUP" {
		t.Errorf("Condition.Type mismatch: got %q, want %q", cond.Type, "TYPE_USERGROUP")
	}
	if cond.Operator != "OPERATOR_IN" {
		t.Errorf("Condition.Operator mismatch: got %q, want %q", cond.Operator, "OPERATOR_IN")
	}
	if len(cond.Values) != 1 || cond.Values[0] != "Everyone" {
		t.Errorf("Condition.Values mismatch: got %v", cond.Values)
	}
}

// TestSessionPolicyFromModel_withActions verifies that action fields are
// correctly mapped from the Terraform model to the API struct.
func TestSessionPolicyFromModel_withActions(t *testing.T) {
	ctx := context.Background()

	data := SessionPolicyResourceModel{
		Name:   types.StringValue("policy-with-actions"),
		Active: types.BoolValue(true),
		Rules: []SessionPolicyRuleModel{
			{
				Priority: types.Int64Value(1),
				Active:   types.BoolValue(true),
				Actions: &SessionPolicyActionsModel{
					Routing:               types.StringValue("default"),
					DisableSecurityGroups: types.StringNull(),
					LocalLanAccess:        types.StringNull(),
				},
				Conditions: []SessionPolicyConditionModel{},
			},
		},
	}

	var diags diag.Diagnostics
	policy := sessionPolicyFromModel(ctx, &data, &diags)
	if diags.HasError() {
		t.Fatalf("sessionPolicyFromModel returned errors: %v", diags.Errors())
	}

	if len(policy.GenericRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(policy.GenericRules))
	}
	rule := policy.GenericRules[0]
	if rule.Actions.Routing != "default" {
		t.Errorf("Actions.Routing mismatch: got %q, want %q", rule.Actions.Routing, "default")
	}
	if rule.Actions.DisableSecurityGroups != "" {
		t.Errorf("Actions.DisableSecurityGroups should be empty, got %q", rule.Actions.DisableSecurityGroups)
	}
}

// TestSessionPolicyToModel_basic verifies that sessionPolicyToModel correctly
// maps an API response back into a Terraform model.
func TestSessionPolicyToModel_basic(t *testing.T) {
	ctx := context.Background()

	p50 := 50
	policy := &SessionPolicy{
		ID:          "policy-uuid-1",
		Name:        "my-session-policy",
		Description: "session policy description",
		Active:      true,
		Priority:    &p50,
		GenericRules: []SessionPolicyRule{
			{
				ID:       "rule-uuid-1",
				Name:     "Rule One",
				Priority: 1,
				Active:   true,
				Actions: SessionPolicyAction{
					Routing: "default",
				},
				Conditions: []SessionPolicyCondition{
					{
						Type:     "TYPE_USERGROUP",
						Operator: "OPERATOR_IN",
						Values:   []string{"Everyone"},
					},
				},
			},
		},
	}

	data := &SessionPolicyResourceModel{}
	var diags diag.Diagnostics
	sessionPolicyToModel(ctx, policy, data, &diags)
	if diags.HasError() {
		t.Fatalf("sessionPolicyToModel returned errors: %v", diags.Errors())
	}

	if data.ID.ValueString() != "policy-uuid-1" {
		t.Errorf("ID mismatch: got %q, want %q", data.ID.ValueString(), "policy-uuid-1")
	}
	if data.Name.ValueString() != "my-session-policy" {
		t.Errorf("Name mismatch: got %q, want %q", data.Name.ValueString(), "my-session-policy")
	}
	if data.Description.ValueString() != "session policy description" {
		t.Errorf("Description mismatch: got %q, want %q", data.Description.ValueString(), "session policy description")
	}
	if !data.Active.ValueBool() {
		t.Errorf("Active should be true")
	}
	if data.Priority.ValueInt64() != 50 {
		t.Errorf("Priority mismatch: got %d, want 50", data.Priority.ValueInt64())
	}

	if len(data.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(data.Rules))
	}
	rule := data.Rules[0]
	if rule.ID.ValueString() != "rule-uuid-1" {
		t.Errorf("Rule.ID mismatch: got %q, want %q", rule.ID.ValueString(), "rule-uuid-1")
	}
	if rule.Name.ValueString() != "Rule One" {
		t.Errorf("Rule.Name mismatch: got %q, want %q", rule.Name.ValueString(), "Rule One")
	}
	if rule.Priority.ValueInt64() != 1 {
		t.Errorf("Rule.Priority mismatch: got %d, want 1", rule.Priority.ValueInt64())
	}
	if !rule.Active.ValueBool() {
		t.Errorf("Rule.Active should be true")
	}

	// Actions should be populated since Routing is non-empty
	if rule.Actions == nil {
		t.Fatal("Rule.Actions should not be nil when routing is set")
	}
	if rule.Actions.Routing.ValueString() != "default" {
		t.Errorf("Rule.Actions.Routing mismatch: got %q, want %q", rule.Actions.Routing.ValueString(), "default")
	}

	if len(rule.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(rule.Conditions))
	}
	cond := rule.Conditions[0]
	if cond.Type.ValueString() != "TYPE_USERGROUP" {
		t.Errorf("Condition.Type mismatch: got %q", cond.Type.ValueString())
	}
}

// TestSessionPolicyToModel_emptyDescription verifies that an empty description
// from the API is always stored as an empty string in state, regardless of
// prior state. description is Computed: true so the schema (UseStateForUnknown)
// handles plan-time drift — the Read function must faithfully reflect the API
// value so that import sets state = "" when the API returns "".
func TestSessionPolicyToModel_emptyDescription(t *testing.T) {
	ctx := context.Background()

	p10 := 10
	policy := &SessionPolicy{
		ID:          "policy-uuid-2",
		Name:        "no-description-policy",
		Description: "",
		Active:      false,
		Priority:    &p10,
	}

	// Prior state has description as null (e.g. fresh import)
	data := &SessionPolicyResourceModel{
		Description: types.StringNull(),
	}
	var diags diag.Diagnostics
	sessionPolicyToModel(ctx, policy, data, &diags)
	if diags.HasError() {
		t.Fatalf("sessionPolicyToModel returned errors: %v", diags.Errors())
	}

	if data.Description.ValueString() != "" {
		t.Errorf("Description should be empty string when API returns empty, got %q", data.Description.ValueString())
	}

	// Prior state has description as empty string — same expected result
	data2 := &SessionPolicyResourceModel{
		Description: types.StringValue(""),
	}
	var diags2 diag.Diagnostics
	sessionPolicyToModel(ctx, policy, data2, &diags2)
	if diags2.HasError() {
		t.Fatalf("sessionPolicyToModel returned errors: %v", diags2.Errors())
	}

	if data2.Description.ValueString() != "" {
		t.Errorf("Description should be empty string when API returns empty, got %q", data2.Description.ValueString())
	}
}

// TestSessionPolicyStructSerialization verifies that the SessionPolicy struct
// serialises correctly to JSON. Priority is a pointer with omitempty so it is
// omitted when nil (letting the server auto-assign on POST) and present when
// explicitly set, including when the value is zero.
func TestSessionPolicyStructSerialization(t *testing.T) {
	// Nil priority must be omitted so the server can auto-assign on POST.
	policy := SessionPolicy{
		Name:   "serialization-test",
		Active: true,
	}
	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"name":"serialization-test"`) {
		t.Errorf("Name missing from JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"active":true`) {
		t.Errorf("Active missing from JSON: %s", jsonStr)
	}
	// Nil pointer: priority must be absent so the server can auto-assign.
	if strings.Contains(jsonStr, `"priority"`) {
		t.Errorf("Priority should be absent when nil, but found in: %s", jsonStr)
	}

	// Explicit zero priority must appear in JSON.
	p0 := 0
	policy.Priority = &p0
	data, _ = json.Marshal(policy)
	if !strings.Contains(string(data), `"priority":0`) {
		t.Errorf("Priority 0 missing from JSON (must be serialised when explicitly set): %s", string(data))
	}

	// A non-zero priority must also be serialised.
	p42 := 42
	policy.Priority = &p42
	data, _ = json.Marshal(policy)
	if !strings.Contains(string(data), `"priority":42`) {
		t.Errorf("Priority 42 missing from JSON: %s", string(data))
	}
}

// TestSessionPolicyRuleActionOmitempty ensures that action fields with omitempty
// are omitted from the JSON when they are empty strings.
func TestSessionPolicyRuleActionOmitempty(t *testing.T) {
	rule := SessionPolicyRule{
		Name:     "empty-actions-rule",
		Priority: 1,
		Active:   true,
		Actions:  SessionPolicyAction{}, // all empty
	}
	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := string(data)
	if strings.Contains(jsonStr, `"routing"`) {
		t.Errorf("routing should be omitted when empty, but found in: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"disableSecurityGroups"`) {
		t.Errorf("disableSecurityGroups should be omitted when empty, but found in: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"localLanAccess"`) {
		t.Errorf("localLanAccess should be omitted when empty, but found in: %s", jsonStr)
	}
}

// TestSessionPolicyConditionRoundtrip exercises the condition struct
// serialisation. metadata is omitted when nil (omitempty); tagSource and
// tagKey are always serialised because the API round-trips them even as
// empty strings.
func TestSessionPolicyConditionRoundtrip(t *testing.T) {
	cond := SessionPolicyCondition{
		Type:     "TYPE_USERGROUP",
		Operator: "OPERATOR_IN",
		Values:   []string{"Everyone"},
	}
	data, err := json.Marshal(cond)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := string(data)
	// tagSource and tagKey are always serialised (no omitempty) because the API
	// may return empty strings and we need to round-trip them faithfully.
	if strings.Contains(jsonStr, `"metadata"`) {
		t.Errorf("metadata should be omitted when nil: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"TYPE_USERGROUP"`) {
		t.Errorf("type missing from JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"Everyone"`) {
		t.Errorf("values missing from JSON: %s", jsonStr)
	}
}

// =============================================================================
// Acceptance test config helpers
// =============================================================================

// testAccSessionPolicyConfig_basic returns an HCL configuration for a minimal
// session policy resource with a single rule and one condition.
// Actions include routing and local_lan_access as required by the test tenant configuration.
func testAccSessionPolicyConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "test" {
  name        = %q
  description = "Terraform acceptance test - basic session policy"
  active      = false

  generic_rules = [
    {
      name        = "Default Rule"
      description = "Allow all users"
      priority    = 1
      active      = true

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
`, name)
}

// testAccSessionPolicyConfig_updated returns the same policy with different
// description, active state, and rule name so that the Update path is exercised.
func testAccSessionPolicyConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "test" {
  name        = %q
  description = "Terraform acceptance test - session policy UPDATED"
  active      = true

  generic_rules = [
    {
      name        = "Updated Rule"
      description = "Updated description"
      priority    = 1
      active      = false

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
`, name)
}

// testAccSessionPolicyConfig_multipleRules returns a session policy with two
// rules to verify that list-of-rules is handled correctly.
func testAccSessionPolicyConfig_multipleRules(name string) string {
	return fmt.Sprintf(`
resource "spa_session_policy" "multi" {
  name        = %q
  description = "Terraform acceptance test - multiple rules"
  active      = false

  generic_rules = [
    {
      name     = "Rule One"
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
    },
    {
      name     = "Rule Two"
      priority = 2
      active   = false

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
`, name)
}

// =============================================================================
// Acceptance test helper functions
// =============================================================================

// testAccCleanupSessionPolicyByName deletes any session policy matching the
// given name left over from a previous failed test run. Errors are ignored.
func testAccCleanupSessionPolicyByName(policyName string) {
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

// testAccCheckSessionPolicyDestroy verifies that all session policy resources
// managed in the test have been removed from the API after destroy.
func testAccCheckSessionPolicyDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "spa_session_policy" {
			continue
		}

		id := rs.Primary.Attributes["id"]
		_, err := client.GetSessionPolicy(ctx, id)
		if err == nil {
			return fmt.Errorf("session policy %s still exists in the API after destroy", id)
		}
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("unexpected error checking session policy %s: %s", id, err)
		}
	}
	return nil
}

// testAccCheckSessionPolicyExistsInAPI queries the API directly to verify that
// the resource exists and its name matches the state.
func testAccCheckSessionPolicyExistsInAPI(resourceName string) resource.TestCheckFunc {
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

		policy, err := client.GetSessionPolicy(context.Background(), id)
		if err != nil {
			return fmt.Errorf("session policy %s not found in API: %s", id, err)
		}

		if policy.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("session policy name mismatch: API=%q, state=%q", policy.Name, rs.Primary.Attributes["name"])
		}

		return nil
	}
}

// =============================================================================
// Acceptance tests
// =============================================================================

// TestAccSessionPolicy_basic covers the full resource lifecycle:
// Create → Read → Update → Import → Delete.
func TestAccSessionPolicy_basic(t *testing.T) {
	name := "tf-acc-session-policy-basic"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccSessionPolicyConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.test"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "name", name),
					resource.TestCheckResourceAttr("spa_session_policy.test", "description", "Terraform acceptance test - basic session policy"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "active", "false"),
					resource.TestCheckResourceAttrSet("spa_session_policy.test", "id"),
					// priority is computed; it must be set after create
					resource.TestCheckResourceAttrSet("spa_session_policy.test", "priority"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.name", "Default Rule"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.description", "Allow all users"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.active", "true"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.condition.#", "1"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.condition.0.type", "TYPE_PLATFORM"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.condition.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.condition.0.values.#", "1"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.condition.0.values.0", "PLATFORM_FILTER_PC"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.actions.routing", "default"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.actions.local_lan_access", "enabled"),
				),
			},
			// Step 2: Update — description, active flag, rule name
			{
				Config: testAccSessionPolicyConfig_updated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.test"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "name", name),
					resource.TestCheckResourceAttr("spa_session_policy.test", "description", "Terraform acceptance test - session policy UPDATED"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "active", "true"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.#", "1"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.name", "Updated Rule"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.description", "Updated description"),
					resource.TestCheckResourceAttr("spa_session_policy.test", "generic_rules.0.active", "false"),
				),
			},
			// Step 3: Import — verifies the state round-trips correctly.
			{
				ResourceName:      "spa_session_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccSessionPolicy_multipleRules verifies that a session policy with more
// than one rule is created, stored, and read back correctly.
func TestAccSessionPolicy_multipleRules(t *testing.T) {
	name := "tf-acc-session-policy-multi"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupSessionPolicyByName(name)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSessionPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSessionPolicyConfig_multipleRules(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSessionPolicyExistsInAPI("spa_session_policy.multi"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "name", name),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.#", "2"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.0.name", "Rule One"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.0.active", "true"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.1.name", "Rule Two"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.1.priority", "2"),
					resource.TestCheckResourceAttr("spa_session_policy.multi", "generic_rules.1.active", "false"),
					resource.TestCheckResourceAttrSet("spa_session_policy.multi", "id"),
				),
			},
			// Import must round-trip all rules.
			{
				ResourceName:      "spa_session_policy.multi",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
