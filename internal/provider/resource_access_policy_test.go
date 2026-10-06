package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccessPolicyRedirectSBSHasDefault guards the write path for the Optional
// redirect_sbs restriction: it must be Computed with a static-false default so
// that omitting it (while still setting restrictions for
// enhanced_security_settings) plans as false instead of null, avoiding
// "inconsistent result after apply" when Read records the backend's false.
func TestAccessPolicyRedirectSBSHasDefault(t *testing.T) {
	ctx := context.Background()
	resp := &fwresource.SchemaResponse{}
	NewAccessPolicyResource().Schema(ctx, fwresource.SchemaRequest{}, resp)

	accessRules, ok := resp.Schema.Attributes["access_rules"].(rschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("access_rules is not a ListNestedAttribute: %T", resp.Schema.Attributes["access_rules"])
	}
	restrictions, ok := accessRules.NestedObject.Attributes["restrictions"].(rschema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("restrictions is not a SingleNestedAttribute: %T", accessRules.NestedObject.Attributes["restrictions"])
	}
	redirect, ok := restrictions.Attributes["redirect_sbs"].(rschema.BoolAttribute)
	if !ok {
		t.Fatalf("redirect_sbs is not a BoolAttribute: %T", restrictions.Attributes["redirect_sbs"])
	}

	if !redirect.Optional {
		t.Error("redirect_sbs must remain Optional")
	}
	if !redirect.Computed {
		t.Error("redirect_sbs must be Computed so an omitted value gets the default")
	}
	if redirect.Default == nil {
		t.Fatal("redirect_sbs must have a static-false default")
	}
	defResp := &defaults.BoolResponse{}
	redirect.Default.DefaultBool(ctx, defaults.BoolRequest{}, defResp)
	if defResp.PlanValue.ValueBool() {
		t.Error("redirect_sbs default = true, want false")
	}
}

// TestValidateTagRuleConfig verifies that TYPE_TAG rules require non-empty
// tag_source and tag_key, while other rule types are unaffected.
func TestValidateTagRuleConfig(t *testing.T) {
	tagRule := func(source, key string) RuleResourceModel {
		return RuleResourceModel{
			Type:      types.StringValue("TYPE_TAG"),
			Operator:  types.StringValue("OPERATOR_IN"),
			TagSource: types.StringValue(source),
			TagKey:    types.StringValue(key),
		}
	}
	model := func(rules ...RuleResourceModel) AccessPolicyResourceModel {
		return AccessPolicyResourceModel{
			AccessRules: []AccessRuleResourceModel{{Rules: rules}},
		}
	}

	tests := []struct {
		name      string
		data      AccessPolicyResourceModel
		wantCount int
	}{
		{
			name:      "valid tag rule",
			data:      model(tagRule("ITM", "location-geo-country-isocode")),
			wantCount: 0,
		},
		{
			name:      "empty tag_source and tag_key",
			data:      model(tagRule("", "")),
			wantCount: 2,
		},
		{
			name:      "empty tag_source only",
			data:      model(tagRule("", "some-key")),
			wantCount: 1,
		},
		{
			name: "null tag fields on tag rule",
			data: model(RuleResourceModel{
				Type:      types.StringValue("TYPE_TAG"),
				Operator:  types.StringValue("OPERATOR_IN"),
				TagSource: types.StringNull(),
				TagKey:    types.StringNull(),
			}),
			wantCount: 2,
		},
		{
			name: "non-tag rule with empty tag fields is fine",
			data: model(RuleResourceModel{
				Type:      types.StringValue("TYPE_USERGROUP"),
				Operator:  types.StringValue("OPERATOR_IN"),
				TagSource: types.StringValue(""),
				TagKey:    types.StringValue(""),
			}),
			wantCount: 0,
		},
		{
			name: "unknown tag fields defer to apply",
			data: model(RuleResourceModel{
				Type:      types.StringValue("TYPE_TAG"),
				Operator:  types.StringValue("OPERATOR_IN"),
				TagSource: types.StringUnknown(),
				TagKey:    types.StringUnknown(),
			}),
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateTagRuleConfig(tc.data)
			if got := diags.ErrorsCount(); got != tc.wantCount {
				t.Fatalf("expected %d errors, got %d: %v", tc.wantCount, got, diags.Errors())
			}
		})
	}
}

// TestAccessRulesHasUnknown verifies ValidateConfig defers (returns true) when
// access_rules or a nested rules collection is unknown, so it never emits a
// conversion error for a config wired to an unresolved expression.
func TestAccessRulesHasUnknown(t *testing.T) {
	ruleType := types.ObjectType{AttrTypes: map[string]attr.Type{"type": types.StringType}}
	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{"rules": types.ListType{ElemType: ruleType}}}

	elemWithRules := func(rules types.List) types.Object {
		o, _ := types.ObjectValue(elemType.AttrTypes, map[string]attr.Value{"rules": rules})
		return o
	}
	knownRule, _ := types.ObjectValue(ruleType.AttrTypes, map[string]attr.Value{"type": types.StringValue("TYPE_TAG")})
	knownRules, _ := types.ListValue(ruleType, []attr.Value{knownRule})
	unknownRuleInList, _ := types.ListValue(ruleType, []attr.Value{types.ObjectUnknown(ruleType.AttrTypes)})

	listOf := func(elems ...attr.Value) types.List {
		l, _ := types.ListValue(elemType, elems)
		return l
	}

	tests := []struct {
		name string
		list types.List
		want bool
	}{
		{"null list is known", types.ListNull(elemType), false},
		{"unknown list defers", types.ListUnknown(elemType), true},
		{"fully known does not defer", listOf(elemWithRules(knownRules)), false},
		{"unknown nested rules list defers", listOf(elemWithRules(types.ListUnknown(ruleType))), true},
		{"unknown nested rule element defers", listOf(elemWithRules(unknownRuleInList)), true},
		{"unknown access_rules element defers", listOf(types.ObjectUnknown(elemType.AttrTypes)), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := accessRulesHasUnknown(tc.list); got != tc.want {
				t.Fatalf("accessRulesHasUnknown = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestAccessRulesHasUnknownNestedBlocks verifies the guard defers for unknown
// values in every Go-native nested block of an access rule (conditions,
// advanced_settings, restrictions, domain_overrides), not just rules — each of
// those decodes into a Go slice/struct/pointer that Config.Get cannot fill with
// an unknown value.
func TestAccessRulesHasUnknownNestedBlocks(t *testing.T) {
	condType := types.ObjectType{AttrTypes: map[string]attr.Type{"platform_filter": types.StringType}}
	doType := types.ObjectType{AttrTypes: map[string]attr.Type{"fqdn": types.StringType}}
	advType := types.ObjectType{AttrTypes: map[string]attr.Type{"domain_overrides": types.ListType{ElemType: doType}}}
	restrType := types.ObjectType{AttrTypes: map[string]attr.Type{"redirect_sbs": types.BoolType}}
	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"conditions":        types.ListType{ElemType: condType},
		"advanced_settings": advType,
		"restrictions":      restrType,
	}}

	listOf := func(attrs map[string]attr.Value) types.List {
		o, _ := types.ObjectValue(elemType.AttrTypes, attrs)
		l, _ := types.ListValue(elemType, []attr.Value{o})
		return l
	}
	knownAdv, _ := types.ObjectValue(advType.AttrTypes, map[string]attr.Value{"domain_overrides": types.ListNull(doType)})
	knownRestr, _ := types.ObjectValue(restrType.AttrTypes, map[string]attr.Value{"redirect_sbs": types.BoolValue(false)})
	knownCond, _ := types.ListValue(condType, []attr.Value{})
	advUnknownDO, _ := types.ObjectValue(advType.AttrTypes, map[string]attr.Value{"domain_overrides": types.ListUnknown(doType)})

	base := map[string]attr.Value{"conditions": knownCond, "advanced_settings": knownAdv, "restrictions": knownRestr}
	with := func(k string, v attr.Value) map[string]attr.Value {
		m := map[string]attr.Value{}
		for kk, vv := range base {
			m[kk] = vv
		}
		m[k] = v
		return m
	}

	tests := []struct {
		name string
		list types.List
		want bool
	}{
		{"all known blocks", listOf(base), false},
		{"unknown conditions list", listOf(with("conditions", types.ListUnknown(condType))), true},
		{"unknown advanced_settings object", listOf(with("advanced_settings", types.ObjectUnknown(advType.AttrTypes))), true},
		{"unknown restrictions object", listOf(with("restrictions", types.ObjectUnknown(restrType.AttrTypes))), true},
		{"unknown domain_overrides nested list", listOf(with("advanced_settings", advUnknownDO)), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := accessRulesHasUnknown(tc.list); got != tc.want {
				t.Fatalf("accessRulesHasUnknown = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestAccessRulesHasUnknownMaps pins the deliberate exception: map attributes
// never force ValidateConfig to defer. A types.Map holds an unknown map, and an
// unknown element inside a known one, without a conversion error, so neither
// shape can break the whole-model Get that validation starts with. Deferring on
// them would silence every unrelated check — a TYPE_TAG rule missing tag_source,
// say — on any resource that wires one sub-rule's metadata to an apply-time
// expression.
func TestAccessRulesHasUnknownMaps(t *testing.T) {
	strMap := types.MapType{ElemType: types.StringType}
	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{"metadata": strMap}}

	listOf := func(m attr.Value) types.List {
		o, _ := types.ObjectValue(elemType.AttrTypes, map[string]attr.Value{"metadata": m})
		l, _ := types.ListValue(elemType, []attr.Value{o})
		return l
	}
	known, _ := types.MapValue(types.StringType, map[string]attr.Value{"k": types.StringValue("v")})
	unknownElem, _ := types.MapValue(types.StringType, map[string]attr.Value{"k": types.StringUnknown()})

	tests := []struct {
		name string
		list types.List
		want bool
	}{
		{"known map", listOf(known), false},
		{"null map", listOf(types.MapNull(types.StringType)), false},
		{"unknown map", listOf(types.MapUnknown(types.StringType)), false},
		{"map with unknown element", listOf(unknownElem), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := accessRulesHasUnknown(tc.list); got != tc.want {
				t.Fatalf("accessRulesHasUnknown = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestUserGroupScopeFromConditionsSkipsUnknownElements is the other half of the
// exception above. ElementsAs into map[string]string is the one read that cannot
// represent an unknown element, so the reader skips such a condition itself
// rather than raising a conversion diagnostic during validation — the value is
// resolved by apply, which is when the scope actually has to be right.
func TestUserGroupScopeFromConditionsSkipsUnknownElements(t *testing.T) {
	ctx := context.Background()
	unknownElem, _ := types.MapValue(types.StringType, map[string]attr.Value{"SID:/a": types.StringUnknown()})

	for _, tc := range []struct {
		name string
		m    types.Map
	}{
		{"wholly unknown map", types.MapUnknown(types.StringType)},
		{"known map with an unknown element", unknownElem},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scope, ok, diags := userGroupScopeFromConditions(ctx, []ConditionResourceModel{{UserAndGroups: tc.m}})
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if ok {
				t.Fatalf("expected the condition to be skipped, got scope %+v", scope)
			}
		})
	}

	// A resolved element in the same condition list is still read.
	known, _ := types.MapValue(types.StringType, map[string]attr.Value{"SID:/b": types.StringValue("Team")})
	scope, ok, diags := userGroupScopeFromConditions(ctx, []ConditionResourceModel{
		{UserAndGroups: unknownElem},
		{UserAndGroups: known},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !ok || !reflect.DeepEqual(scope.Values, []string{"SID:/b"}) {
		t.Fatalf("scope = %+v (ok=%v), want the known condition to still be read", scope, ok)
	}
}

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
		// The deprecated user_and_groups map must never reach the wire; it is
		// translated into a TYPE_USERGROUP rule instead.
		Conditions: []ConditionResourceModel{
			{
				PlatformFilter: types.StringValue("PLATFORM_FILTER_ANY"),
				UserAndGroups: userAndGroupsMap(t, map[string]string{
					"SID:/S-1-5-21-100": "Sales",
				}),
			},
		},
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

		// Convert Conditions — platform_filter only; the deprecated user_and_groups
		// map is folded into a synthesized TYPE_USERGROUP rule.
		conditions := make([]Condition, 0, len(terraformRule.Conditions))
		for _, condData := range terraformRule.Conditions {
			conditions = append(conditions, Condition{
				PlatformFilter: condData.PlatformFilter.ValueString(),
			})
		}
		accessRule.Conditions = conditions

		ugScope, hasUG, ugDiags := userGroupScopeFromConditions(ctx, terraformRule.Conditions)
		if ugDiags.HasError() {
			t.Fatalf("Failed to convert user_and_groups: %v", ugDiags.Errors())
		}
		if hasUG {
			rules = appendUserGroupRule(rules, ugScope)
		}
		accessRule.Rules = rules

		accessRules = append(accessRules, accessRule)
	}

	// Create the policy payload (what would be sent to the API)
	policyPriority := int(data.Priority.ValueInt64())
	policy := &AccessPolicy{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Active:      data.Active.ValueBool(),
		Priority:    &policyPriority,
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

	// Verify nested rule structure: the configured rule plus the rule synthesized
	// from the deprecated user_and_groups map.
	if len(accessRuleFromPayload.Rules) != 2 {
		t.Fatalf("Expected 2 nested rules, got %d", len(accessRuleFromPayload.Rules))
	}
	nestedRule := accessRuleFromPayload.Rules[0]
	if nestedRule.Type != "TYPE_USERGROUP" {
		t.Errorf("Rule.Type incorrect: %s", nestedRule.Type)
	}
	if len(nestedRule.Values) != 2 {
		t.Errorf("Expected 2 rule values, got %d", len(nestedRule.Values))
	}

	synthesized := accessRuleFromPayload.Rules[1]
	if synthesized.Type != userGroupRuleType || synthesized.Operator != userGroupRuleOperator {
		t.Errorf("Synthesized rule has wrong type/operator: %+v", synthesized)
	}
	if len(synthesized.Values) != 1 || synthesized.Values[0] != "SID:/S-1-5-21-100" {
		t.Errorf("Synthesized rule values incorrect: %v", synthesized.Values)
	}
	if got := synthesized.Metadata["Sales"]; got != "SID:/S-1-5-21-100" {
		t.Errorf("Synthesized rule metadata incorrect: %v", synthesized.Metadata)
	}

	// The deprecated field must never reach the wire.
	payloadJSON, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("Failed to marshal policy payload: %v", err)
	}
	for _, key := range []string{"userAndGroups", "userandgroups", "user_and_groups"} {
		if strings.Contains(string(payloadJSON), key) {
			t.Errorf("Payload must not contain %q: %s", key, string(payloadJSON))
		}
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

// TestAccessPolicyPriorityZeroSerialization verifies the pointer-based priority
// serialization: an omitted (nil) policy priority is absent from the request so
// the backend auto-assigns one, while an explicit value (including 0) is sent
// verbatim. Access-rule priority is required by the API, so it is always emitted.
func TestAccessPolicyPriorityZeroSerialization(t *testing.T) {
	// Omitted policy priority (nil) must NOT appear in the payload so the backend
	// assigns the next available priority.
	omitted := AccessPolicy{Name: "test-priority-omitted", Active: true}
	data, err := json.Marshal(omitted)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"priority"`) {
		t.Errorf("omitted priority should be absent from AccessPolicy JSON: %s", string(data))
	}

	// Explicit policy priority 0 must be preserved verbatim.
	zero := 0
	explicitZero := AccessPolicy{Name: "test-priority-zero", Active: true, Priority: &zero}
	data, err = json.Marshal(explicitZero)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"priority":0`) {
		t.Errorf("explicit priority 0 missing from AccessPolicy JSON: %s", string(data))
	}

	// Access-rule priority is required and always emitted.
	rule := AccessRule{
		Name:     "rule-priority-zero",
		Active:   true,
		Priority: 0,
	}
	data, err = json.Marshal(rule)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"priority":0`) {
		t.Errorf("Priority 0 missing from AccessRule JSON: %s", jsonStr)
	}

	// Non-zero explicit policy priority is sent as-is.
	five := 5
	explicitFive := AccessPolicy{Name: "test-priority-five", Active: true, Priority: &five}
	data, _ = json.Marshal(explicitFive)
	if !strings.Contains(string(data), `"priority":5`) {
		t.Errorf("Priority 5 missing from AccessPolicy JSON: %s", string(data))
	}
}

// TestAccessRuleNameOptional verifies that an access rule name is optional: when
// omitted it is dropped from the JSON payload (the SPAConfig serializer treats
// accessRules[].name as required=False), and when provided it is serialized.
func TestAccessRuleNameOptional(t *testing.T) {
	ruleWithoutName := AccessRule{
		Active:   true,
		Priority: 1,
		Access:   "ACCESS_ALLOW",
	}
	data, err := json.Marshal(ruleWithoutName)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"name"`) {
		t.Errorf("Expected name to be omitted from AccessRule JSON when empty: %s", string(data))
	}

	ruleWithName := AccessRule{
		Name:     "named-rule",
		Active:   true,
		Priority: 1,
		Access:   "ACCESS_ALLOW",
	}
	data, err = json.Marshal(ruleWithName)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"name":"named-rule"`) {
		t.Errorf("Expected name to be present in AccessRule JSON when set: %s", string(data))
	}
}

// TestAccessRuleNameSchemaOptional verifies at the schema level that the nested
// access_rules[].name attribute is optional (not required), matching the SPA
// service which does not require a name on an access rule.
func TestAccessRuleNameSchemaOptional(t *testing.T) {
	attrs := resourceSchemaAttributes(t, NewAccessPolicyResource())

	accessRules, ok := attrs["access_rules"].(rschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("expected access_rules to be a ListNestedAttribute, got %T", attrs["access_rules"])
	}
	nameAttr, ok := accessRules.NestedObject.Attributes["name"]
	if !ok {
		t.Fatal("expected access_rules[].name attribute to exist")
	}
	if nameAttr.IsRequired() {
		t.Error("expected access_rules[].name to be optional, but it is required")
	}
	if !nameAttr.IsOptional() {
		t.Error("expected access_rules[].name to be optional")
	}
}

// userAndGroupsMap is a test helper building the deprecated
// conditions[].user_and_groups value from token => display-name pairs.
func userAndGroupsMap(t *testing.T, entries map[string]string) types.Map {
	t.Helper()
	elems := make(map[string]attr.Value, len(entries))
	for k, v := range entries {
		elems[k] = types.StringValue(v)
	}
	m, diags := types.MapValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("failed to build user_and_groups map: %v", diags.Errors())
	}
	return m
}

// TestUserGroupScopeFromConditions covers the translation of the deprecated
// conditions[].user_and_groups map into the TYPE_USERGROUP rule the provider
// sends in its place. Ordering must be deterministic: Go randomises
// map iteration, so an unsorted rule would produce a different payload on every
// run and churn the policy on the backend.
func TestUserGroupScopeFromConditions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		conditions   []ConditionResourceModel
		wantOK       bool
		wantValues   []string
		wantMetadata map[string]string
	}{
		{
			name:       "no conditions",
			conditions: nil,
			wantOK:     false,
		},
		{
			name: "null map",
			conditions: []ConditionResourceModel{
				{PlatformFilter: types.StringValue(""), UserAndGroups: types.MapNull(types.StringType)},
			},
			wantOK: false,
		},
		{
			name: "explicit empty map",
			conditions: []ConditionResourceModel{
				{PlatformFilter: types.StringValue(""), UserAndGroups: userAndGroupsMap(t, map[string]string{})},
			},
			wantOK: false,
		},
		{
			name: "single entry",
			conditions: []ConditionResourceModel{
				{UserAndGroups: userAndGroupsMap(t, map[string]string{
					"SID:/cust/S-1-5-21-1": "Alice",
				})},
			},
			wantOK:       true,
			wantValues:   []string{"SID:/cust/S-1-5-21-1"},
			wantMetadata: map[string]string{"Alice": "SID:/cust/S-1-5-21-1"},
		},
		{
			name: "entries merged across conditions and sorted",
			conditions: []ConditionResourceModel{
				{UserAndGroups: userAndGroupsMap(t, map[string]string{
					"SID:/cust/S-1-5-21-2": "Bob",
				})},
				{UserAndGroups: userAndGroupsMap(t, map[string]string{
					"OID:/ad/None:abc": "Alice",
				})},
			},
			wantOK:     true,
			wantValues: []string{"OID:/ad/None:abc", "SID:/cust/S-1-5-21-2"},
			wantMetadata: map[string]string{
				"Alice": "OID:/ad/None:abc",
				"Bob":   "SID:/cust/S-1-5-21-2",
			},
		},
		{
			name: "tokens sharing a display name are comma-joined",
			conditions: []ConditionResourceModel{
				{UserAndGroups: userAndGroupsMap(t, map[string]string{
					"SID:/cust/S-1-5-21-9": "Alice",
					"OID:/ad/None:abc":     "Alice",
				})},
			},
			wantOK:     true,
			wantValues: []string{"OID:/ad/None:abc", "SID:/cust/S-1-5-21-9"},
			wantMetadata: map[string]string{
				"Alice": "OID:/ad/None:abc,SID:/cust/S-1-5-21-9",
			},
		},
		{
			name: "empty display name falls back to the token",
			conditions: []ConditionResourceModel{
				{UserAndGroups: userAndGroupsMap(t, map[string]string{"all_users": ""})},
			},
			wantOK:       true,
			wantValues:   []string{"all_users"},
			wantMetadata: map[string]string{"all_users": "all_users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, ok, diags := userGroupScopeFromConditions(ctx, tt.conditions)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if !reflect.DeepEqual(scope.Values, tt.wantValues) {
				t.Errorf("values = %v, want %v", scope.Values, tt.wantValues)
			}
			if !reflect.DeepEqual(scope.Metadata, tt.wantMetadata) {
				t.Errorf("metadata = %v, want %v", scope.Metadata, tt.wantMetadata)
			}

			rule := userGroupRuleFromScope(scope)
			if rule.Type != "TYPE_USERGROUP" {
				t.Errorf("rule type = %q, want TYPE_USERGROUP", rule.Type)
			}
			if rule.Operator != "OPERATOR_IN" {
				t.Errorf("rule operator = %q, want OPERATOR_IN", rule.Operator)
			}
			if rule.TagSource != "" || rule.TagKey != "" {
				t.Errorf("tag_source/tag_key should stay empty, got %q/%q", rule.TagSource, rule.TagKey)
			}
			if !reflect.DeepEqual(rule.Values, tt.wantValues) {
				t.Errorf("rule values = %v, want %v", rule.Values, tt.wantValues)
			}
		})
	}
}

// TestAppendUserGroupRule verifies the synthesized rule is added once and is
// skipped when the configuration already declares an equivalent rule.
func TestAppendUserGroupRule(t *testing.T) {
	ctx := context.Background()
	conditions := []ConditionResourceModel{
		{UserAndGroups: userAndGroupsMap(t, map[string]string{"SID:/cust/S-1": "Alice"})},
	}
	scope, ok, diags := userGroupScopeFromConditions(ctx, conditions)
	if diags.HasError() || !ok {
		t.Fatalf("failed to build scope: ok=%v diags=%v", ok, diags.Errors())
	}

	t.Run("appended when absent", func(t *testing.T) {
		rules := []Rule{{Type: "TYPE_PLATFORM", Operator: "OPERATOR_IN", Values: []string{"WINDOWS"}}}
		got := appendUserGroupRule(rules, scope)
		if len(got) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(got))
		}
		if got[1].Type != "TYPE_USERGROUP" {
			t.Fatalf("expected the appended rule to be TYPE_USERGROUP, got %q", got[1].Type)
		}
	})

	t.Run("not duplicated when already declared", func(t *testing.T) {
		rules := []Rule{userGroupRuleFromScope(scope)}
		got := appendUserGroupRule(rules, scope)
		if len(got) != 1 {
			t.Fatalf("expected the equivalent rule to be reused, got %d rules", len(got))
		}
	})
}

// TestReconcileSynthesizedUserGroupRule covers the Read-side collapse. rules[]
// is Required, so the rule the write path synthesizes must not reach state or
// Terraform reports an inconsistent result after apply — but a rule the
// operator declared themselves must survive, and a scope the service has lost
// must be reported as drift.
func TestReconcileSynthesizedUserGroupRule(t *testing.T) {
	ctx := context.Background()
	conditions := []ConditionResourceModel{
		{UserAndGroups: userAndGroupsMap(t, map[string]string{"SID:/cust/S-1": "Alice"})},
	}
	scope, ok, diags := userGroupScopeFromConditions(ctx, conditions)
	if diags.HasError() || !ok {
		t.Fatalf("failed to build scope: ok=%v diags=%v", ok, diags.Errors())
	}
	configRule := Rule{Type: "TYPE_PLATFORM", Operator: "OPERATOR_IN", Values: []string{"WINDOWS"}}

	t.Run("removes the synthesized rule", func(t *testing.T) {
		apiRules := []Rule{configRule, userGroupRuleFromScope(scope)}
		got := reconcileSynthesizedUserGroupRule(apiRules, 0, scope)
		if len(got) != 1 || got[0].Type != "TYPE_PLATFORM" {
			t.Fatalf("expected only the configured rule to remain, got %+v", got)
		}
	})

	t.Run("keeps a rule the configuration declares itself", func(t *testing.T) {
		// The configuration declares exactly the rule the map translates to, so
		// the write path appended nothing and there is nothing to strip.
		declared := userGroupRuleFromScope(scope)
		got := reconcileSynthesizedUserGroupRule([]Rule{declared}, 1, scope)
		if len(got) != 1 {
			t.Fatalf("expected the operator's own rule to survive, got %+v", got)
		}
	})

	t.Run("tolerates reordered values from the backend", func(t *testing.T) {
		multi := []ConditionResourceModel{
			{UserAndGroups: userAndGroupsMap(t, map[string]string{
				"SID:/cust/S-1":    "Alice",
				"OID:/ad/None:abc": "Bob",
			})},
		}
		multiScope, ok, diags := userGroupScopeFromConditions(ctx, multi)
		if diags.HasError() || !ok {
			t.Fatalf("failed to build scope: ok=%v diags=%v", ok, diags.Errors())
		}
		reordered := userGroupRuleFromScope(multiScope)
		reordered.Values = []string{"SID:/cust/S-1", "OID:/ad/None:abc"}
		got := reconcileSynthesizedUserGroupRule([]Rule{reordered}, 0, multiScope)
		if len(got) != 0 {
			t.Fatalf("expected the reordered rule to be recognised, got %+v", got)
		}
	})

	t.Run("leaves the list alone when the scope rule was deleted out of band", func(t *testing.T) {
		// Whether the scope is still enforced is decided by the caller with
		// countScopeMatches; there is nothing for this function to strip.
		apiRules := []Rule{configRule}
		got := reconcileSynthesizedUserGroupRule(apiRules, 0, scope)
		if len(got) != 1 {
			t.Fatalf("expected the list to be unchanged, got %+v", got)
		}
		if countScopeMatches(got, scope) != 0 {
			t.Fatal("expected the scope to no longer match any rule")
		}
	})
}

// TestCountDeclaredScopeMatches checks the state-side count that tells the read
// path whether the write path actually synthesized a rule.
func TestCountDeclaredScopeMatches(t *testing.T) {
	ctx := context.Background()
	conditions := []ConditionResourceModel{
		{UserAndGroups: userAndGroupsMap(t, map[string]string{"SID:/cust/S-1": "Alice"})},
	}
	scope, ok, diags := userGroupScopeFromConditions(ctx, conditions)
	if diags.HasError() || !ok {
		t.Fatalf("failed to build scope: ok=%v diags=%v", ok, diags.Errors())
	}

	declared := ruleModelFromRule(t, userGroupRuleFromScope(scope))
	other := ruleModelFromRule(t, Rule{
		Type: "TYPE_PLATFORM", Operator: "OPERATOR_IN", Values: []string{"WINDOWS"},
	})

	for _, tc := range []struct {
		name  string
		rules []RuleResourceModel
		want  int
	}{
		{"no rules", nil, 0},
		{"unrelated rule only", []RuleResourceModel{other}, 0},
		{"equivalent rule declared", []RuleResourceModel{other, declared}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := countDeclaredScopeMatches(ctx, tc.rules, scope)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if got != tc.want {
				t.Fatalf("expected %d matches, got %d", tc.want, got)
			}
		})
	}
}

// ruleModelFromRule converts an API rule into the state shape the read path
// works with.
func ruleModelFromRule(t *testing.T, rule Rule) RuleResourceModel {
	t.Helper()
	values := make([]attr.Value, 0, len(rule.Values))
	for _, v := range rule.Values {
		values = append(values, types.StringValue(v))
	}
	valueList, diags := types.ListValue(types.StringType, values)
	if diags.HasError() {
		t.Fatalf("failed to build values: %v", diags.Errors())
	}
	metadata := make(map[string]attr.Value, len(rule.Metadata))
	for k, v := range rule.Metadata {
		metadata[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	metadataMap, diags := types.MapValue(types.StringType, metadata)
	if diags.HasError() {
		t.Fatalf("failed to build metadata: %v", diags.Errors())
	}
	return RuleResourceModel{
		Type:      types.StringValue(rule.Type),
		Operator:  types.StringValue(rule.Operator),
		TagSource: types.StringValue(rule.TagSource),
		TagKey:    types.StringValue(rule.TagKey),
		Values:    valueList,
		Metadata:  metadataMap,
	}
}

// TestValidateDeprecatedUserAndGroups checks the plan-time warnings: one for the
// deprecation itself, one per token whose contents the service validates (those
// can fail the apply), and one per token the SPA Console cannot resolve (those
// apply cleanly but break its policy edit page).
func TestValidateDeprecatedUserAndGroups(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		userAndGroups  types.Map
		wantWarnings   int
		wantSnippet    string
		wantNoWarnings bool
	}{
		{
			name:           "absent map warns about nothing",
			userAndGroups:  types.MapNull(types.StringType),
			wantNoWarnings: true,
		},
		{
			name:           "empty map warns about nothing",
			userAndGroups:  userAndGroupsMap(t, map[string]string{}),
			wantNoWarnings: true,
		},
		{
			name:          "populated map warns once with the replacement rule",
			userAndGroups: userAndGroupsMap(t, map[string]string{"SID:/cust/S-1": "Alice"}),
			wantWarnings:  1,
			wantSnippet:   `type     = "TYPE_USERGROUP"`,
		},
		{
			name:          "unresolvable token adds a second warning",
			userAndGroups: userAndGroupsMap(t, map[string]string{"Alice": "Alice"}),
			wantWarnings:  2,
			wantSnippet:   "not a directory token",
		},
		{
			// EMAIL:/ is the form the service actually checks, so it must be
			// called out even though it is a well-formed directory token.
			name:          "service-validated token adds a second warning",
			userAndGroups: userAndGroupsMap(t, map[string]string{"EMAIL:/alice@example.com": "Alice"}),
			wantWarnings:  2,
			wantSnippet:   "checked by the service",
		},
		{
			// A plain SID is stored verbatim and resolves in the Console, so
			// the deprecation notice is the only thing worth saying.
			name:          "resolvable unvalidated token adds no extra warning",
			userAndGroups: userAndGroupsMap(t, map[string]string{"SID:/cust/S-1": "Alice"}),
			wantWarnings:  1,
			wantSnippet:   "TYPE_USERGROUP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := AccessPolicyResourceModel{
				AccessRules: []AccessRuleResourceModel{{
					Conditions: []ConditionResourceModel{{
						PlatformFilter: types.StringValue(""),
						UserAndGroups:  tt.userAndGroups,
					}},
				}},
			}
			diags := validateDeprecatedUserAndGroups(ctx, data)
			if diags.HasError() {
				t.Fatalf("expected warnings only, got errors: %v", diags.Errors())
			}
			warnings := diags.Warnings()
			if tt.wantNoWarnings {
				if len(warnings) != 0 {
					t.Fatalf("expected no warnings, got %d: %v", len(warnings), warnings)
				}
				return
			}
			if len(warnings) != tt.wantWarnings {
				t.Fatalf("expected %d warnings, got %d: %v", tt.wantWarnings, len(warnings), warnings)
			}
			joined := ""
			for _, w := range warnings {
				joined += w.Detail()
			}
			if !strings.Contains(joined, tt.wantSnippet) {
				t.Fatalf("warning detail missing %q, got: %s", tt.wantSnippet, joined)
			}
		})
	}
}

func testAccCheckAccessPolicyDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "citrixspa_access_policy" {
			continue
		}

		id := rs.Primary.Attributes["id"]
		_, err := client.GetAccessPolicy(ctx, id)
		if err == nil {
			return fmt.Errorf("access policy %s still exists in the API after destroy", id)
		}
		if !IsNotFound(err) {
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
	omitDescription bool // when true, the description attribute is omitted entirely (null in config)
	active          bool
	appResourceName string // HCL resource label of the citrixspa_application to reference
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
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccOmittedOptionalStringsAccessRulesHCL omits every optional string that
// this schema change makes null/"" equivalent: the access-rule description and
// the rule tag_source/tag_key. Combined with an omitted policy description, it
// exercises the null-config path that Optional+Computed must apply cleanly.
const testAccOmittedOptionalStringsAccessRulesHCL = `  access_rules = [
    {
      name          = "Omitted Optionals Rule"
      priority      = 1
      active        = true
      access        = "ACCESS_DENY"
      access_native = "ACCESS_DENY"

      rules = [
        {
          type     = "TYPE_USERGROUP"
          operator = "OPERATOR_IN"
          values   = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccSetOptionalStringsAccessRulesHCL sets every optional string to a
// non-empty value (access-rule description plus a TYPE_TAG rule with
// tag_source/tag_key). Transitioning from this to the omitted variant above
// proves that omission clears the previously-set values instead of retaining
// them (the "" default). A TYPE_USERGROUP rule is included because the API
// rejects an access rule that has no user/group rule.
const testAccSetOptionalStringsAccessRulesHCL = `  access_rules = [
    {
      name          = "Set Optionals Rule"
      description   = "Access rule description"
      priority      = 1
      active        = true
      access        = "ACCESS_DENY"
      access_native = "ACCESS_DENY"

      rules = [
        {
          type     = "TYPE_USERGROUP"
          operator = "OPERATOR_IN"
          values   = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        },
        {
          type       = "TYPE_TAG"
          operator   = "OPERATOR_IN"
          tag_source = "ITM"
          tag_key    = "location-geo-country-isocode"
          values     = ["AF"]
        }
      ]
    }
  ]`

// testAccClearedOptionalStringsAccessRulesHCL is the step-2 counterpart of
// testAccSetOptionalStringsAccessRulesHCL for the clear-on-omission test. It
// keeps the SAME two-rule shape but transitions the second rule away from
// TYPE_TAG to TYPE_USERGROUP with tag_source/tag_key omitted (and drops the
// access-rule description). This exercises clearing tag fields when a rule
// changes type in place — a regression that retained the old tag values would
// fail here — while the retained first user/group rule satisfies the API
// requirement that every access rule has a user/group rule.
const testAccClearedOptionalStringsAccessRulesHCL = `  access_rules = [
    {
      name          = "Set Optionals Rule"
      priority      = 1
      active        = true
      access        = "ACCESS_DENY"
      access_native = "ACCESS_DENY"

      rules = [
        {
          type     = "TYPE_USERGROUP"
          operator = "OPERATOR_IN"
          values   = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        },
        {
          type     = "TYPE_USERGROUP"
          operator = "OPERATOR_IN"
          values   = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
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
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccEmptyUserAndGroupsAccessRulesHCL tests the case where a condition sets
// user_and_groups to an explicit empty map ({}). This is the scenario that
// triggers the "was cty.MapValEmpty(cty.String), but now null" inconsistency
// error when the Read method converts an empty user_and_groups map to null
// instead of preserving the empty map written in the config.
const testAccEmptyUserAndGroupsAccessRulesHCL = `  access_rules = [
    {
      name        = "Empty UserAndGroups Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_PC"
          user_and_groups = {}
        }
      ]

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccTranslatedUserAndGroupsAccessRulesHCL covers the deprecated-but-populated
// user_and_groups map. The provider must not send userAndGroups; it
// folds the map into an extra TYPE_USERGROUP rule on the wire and collapses that
// rule again on Read, so the config's single declared rule is what ends up in state.
const testAccTranslatedUserAndGroupsAccessRulesHCL = `  access_rules = [
    {
      name        = "Translated UserAndGroups Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_PC"
          user_and_groups = {
            "SID:/citrix/S-1-5-21-777" = "Engineering"
          }
        }
      ]

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccCollidingUserAndGroupsAccessRulesHCL is the collision case: the rule
// the configuration declares is byte-for-byte the rule user_and_groups
// translates to. The write path therefore appends nothing, so there is no
// synthesized rule to collapse and the declared rule must survive into state.
const testAccCollidingUserAndGroupsAccessRulesHCL = `  access_rules = [
    {
      name        = "Colliding UserAndGroups Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_PC"
          user_and_groups = {
            "SID:/citrix/S-1-5-21-778" = "Engineering"
          }
        }
      ]

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-5-21-778"]
          metadata   = {
            "Engineering" = "SID:/citrix/S-1-5-21-778"
          }
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
          values   = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
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

// testAccEnhancedSecurityDefaultsAccessRulesHCL sets the enhanced-security keys
// whose backend default is "enabled" (downloadV1/uploadV1/clipboardV1/printingV1)
// to that default value. The API omits keys equal to their default from the
// create/read response, so without provider-side reconciliation apply fails with
// "Provider produced inconsistent result after apply ... element ... has vanished".
const testAccEnhancedSecurityDefaultsAccessRulesHCL = `  access_rules = [
    {
      name        = "Defaults Enabled Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      restrictions = {
        redirect_sbs = false
        enhanced_security_settings = {
          _browserV1      = "embeddedBrowser"
          downloadV1      = "enabled"
          uploadV1        = "enabled"
          clipboardV1     = "enabled"
          printingV1      = "enabled"
        }
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccEnhancedSecurityAllDefaultsAccessRulesHCL sets ONLY keys whose value
// equals the backend default and no non-default key. In that case the API omits
// the ENTIRE restrictions object from its response (not just the individual
// keys), so without provider-side reconciliation apply fails with "Provider
// produced inconsistent result after apply ... was cty.ObjectVal(...), but now
// null".
const testAccEnhancedSecurityAllDefaultsAccessRulesHCL = `  access_rules = [
    {
      name        = "All Defaults Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      restrictions = {
        redirect_sbs = false
        enhanced_security_settings = {
          downloadV1  = "enabled"
          uploadV1    = "enabled"
          clipboardV1 = "enabled"
          printingV1  = "enabled"
        }
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccEnhancedSecurityRedirectOnlyAccessRulesHCL sets a restrictions block
// with ONLY redirect_sbs=false (its default) and a NULL enhanced_security_
// settings map. Everything is at default, so the API omits the whole
// restrictions object; the reconstruction branch must fire even though the prior
// ESS map is null, preserving that null map instead of forcing an empty one.
const testAccEnhancedSecurityRedirectOnlyAccessRulesHCL = `  access_rules = [
    {
      name        = "Redirect Only Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      restrictions = {
        redirect_sbs = false
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccEnhancedSecurityRedirectTrueAccessRulesHCL sets a restrictions block
// with a NON-default redirect_sbs=true and a NULL enhanced_security_settings
// map. The API keeps the block (redirect_sbs is non-default) but returns no ESS
// keys, so state mapping must preserve the null ESS map rather than forcing an
// empty one.
const testAccEnhancedSecurityRedirectTrueAccessRulesHCL = `  access_rules = [
    {
      name        = "Redirect True Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      restrictions = {
        redirect_sbs = true
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccEnhancedSecurityNonDefaultNoRedirectAccessRulesHCL sets a restrictions
// block with NON-default enhanced_security_settings (so the API keeps the block)
// but OMITS redirect_sbs entirely (plan value is null). The API returns
// redirect_sbs=false (its backend default); state mapping must preserve the
// prior null rather than coercing it to false, which would otherwise raise
// "inconsistent result after apply (.redirect_sbs: was null, but now false)".
// The ESS keys are all non-default and round-trip exactly as returned by the API.
const testAccEnhancedSecurityNonDefaultNoRedirectAccessRulesHCL = `  access_rules = [
    {
      name        = "Non Default No Redirect Rule"
      priority    = 1
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

      conditions = [
        {
          platform_filter = "PLATFORM_FILTER_ANY"
        }
      ]

      restrictions = {
        enhanced_security_settings = {
          _browserV1      = "embeddedBrowser"
          keyLoggingV1    = "disabled"
          screenCaptureV1 = "disabled"
          uploadV1        = "disabled"
          proxyTrafficV1  = "secureBrowse"
        }
      }

      rules = [
        {
          type       = "TYPE_USERGROUP"
          operator   = "OPERATOR_IN"
          tag_source = ""
          tag_key    = ""
          values     = ["SID:/citrix/S-1-1-0"]
          metadata   = {
            "Everyone" = "SID:/citrix/S-1-1-0"
          }
        }
      ]
    }
  ]`

// testAccAccessPolicyConfig generates a Terraform HCL config for a citrixspa_access_policy resource.
func testAccAccessPolicyConfig(cfg testAccessPolicyConfig) string {
	if cfg.resourceName == "" {
		cfg.resourceName = "test"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "resource \"citrixspa_access_policy\" %q {\n", cfg.resourceName)
	fmt.Fprintf(&b, "  name        = %q\n", cfg.name)
	if !cfg.omitDescription {
		fmt.Fprintf(&b, "  description = %q\n", cfg.description)
	}
	fmt.Fprintf(&b, "  active      = %v\n", cfg.active)
	fmt.Fprintf(&b, "  apps        = [citrixspa_application.%s.id]\n", cfg.appResourceName)
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
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
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

func testAccAccessPolicyConfig_omittedOptionalStrings(name string) string {
	const rdResourceName = "test_domain_for_omitted_policy"
	const appResourceName = "test_app_for_omitted_policy"
	const fqdn = "tf-acc-test-omitted-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for omitted-optionals access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-omitted-policy",
		appType:      "web",
		description:  "Test app for omitted-optionals access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	// Policy description omitted; access_rules HCL omits rule tag_source/tag_key
	// and the access-rule description.
	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_omitted",
		name:            name,
		omitDescription: true,
		active:          false,
		appResourceName: appResourceName,
		priority:        997,
		accessRulesHCL:  testAccOmittedOptionalStringsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

// testAccAccessPolicyConfig_clearScenario builds a policy whose optional strings
// are either all set (setOptionals=true) or all omitted (false), on a shared
// routing domain/app so the two can be applied in sequence as an in-place update.
func testAccAccessPolicyConfig_clearScenario(name string, setOptionals bool) string {
	const rdResourceName = "test_domain_for_clear_policy"
	const appResourceName = "test_app_for_clear_policy"
	const fqdn = "tf-acc-test-clear-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for clear-optionals access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-clear-policy",
		appType:      "web",
		description:  "Test app for clear-optionals access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	cfg := testAccessPolicyConfig{
		resourceName:    "test_clear",
		name:            name,
		active:          false,
		appResourceName: appResourceName,
		priority:        996,
	}
	if setOptionals {
		cfg.description = "Policy description set"
		cfg.accessRulesHCL = testAccSetOptionalStringsAccessRulesHCL
	} else {
		cfg.omitDescription = true
		cfg.accessRulesHCL = testAccClearedOptionalStringsAccessRulesHCL
	}

	return rdConfig + appConfig + testAccAccessPolicyConfig(cfg)
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
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
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
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
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
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
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

func testAccAccessPolicyConfig_enhancedSecurityDefaults(name string) string {
	const rdResourceName = "test_domain_for_ess_policy"
	const appResourceName = "test_app_for_ess_policy"
	const fqdn = "tf-acc-test-ess-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for enhanced-security policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-ess-policy",
		appType:      "web",
		description:  "Test app for enhanced-security policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_ess_defaults",
		name:            name,
		description:     "Terraform acceptance test - enhanced_security_settings defaults",
		active:          false,
		appResourceName: appResourceName,
		priority:        996,
		accessRulesHCL:  testAccEnhancedSecurityDefaultsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_enhancedSecurityAllDefaults(name string) string {
	const rdResourceName = "test_domain_for_ess_alldef_policy"
	const appResourceName = "test_app_for_ess_alldef_policy"
	const fqdn = "tf-acc-test-ess-alldef-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for all-default enhanced-security policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-ess-alldef-policy",
		appType:      "web",
		description:  "Test app for all-default enhanced-security policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_ess_alldef",
		name:            name,
		description:     "Terraform acceptance test - enhanced_security_settings all defaults",
		active:          false,
		appResourceName: appResourceName,
		priority:        995,
		accessRulesHCL:  testAccEnhancedSecurityAllDefaultsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_enhancedSecurityRedirectOnly(name string) string {
	const rdResourceName = "test_domain_for_ess_redirect_policy"
	const appResourceName = "test_app_for_ess_redirect_policy"
	const fqdn = "tf-acc-test-ess-redirect-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for redirect-only restrictions policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-ess-redirect-policy",
		appType:      "web",
		description:  "Test app for redirect-only restrictions policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_ess_redirect",
		name:            name,
		description:     "Terraform acceptance test - restrictions redirect_sbs only",
		active:          false,
		appResourceName: appResourceName,
		priority:        993,
		accessRulesHCL:  testAccEnhancedSecurityRedirectOnlyAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_enhancedSecurityRedirectTrue(name string) string {
	const rdResourceName = "test_domain_for_ess_redirect_true_policy"
	const appResourceName = "test_app_for_ess_redirect_true_policy"
	const fqdn = "tf-acc-test-ess-redirect-true-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for redirect-true restrictions policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-ess-redirect-true-policy",
		appType:      "web",
		description:  "Test app for redirect-true restrictions policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_ess_redirect_true",
		name:            name,
		description:     "Terraform acceptance test - restrictions redirect_sbs true",
		active:          false,
		appResourceName: appResourceName,
		priority:        992,
		accessRulesHCL:  testAccEnhancedSecurityRedirectTrueAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_enhancedSecurityNonDefaultNoRedirect(name string) string {
	const rdResourceName = "test_domain_for_ess_nodefault_noredirect_policy"
	const appResourceName = "test_app_for_ess_nodefault_noredirect_policy"
	const fqdn = "tf-acc-test-ess-nodefault-noredirect-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test RD for non-default ESS no-redirect policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-ess-nodefault-noredirect-policy",
		appType:      "web",
		description:  "Test app for non-default-ESS no-redirect restrictions policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_ess_nodefault_noredirect",
		name:            name,
		description:     "Terraform acceptance test - non-default ESS, redirect_sbs omitted",
		active:          false,
		appResourceName: appResourceName,
		priority:        990,
		accessRulesHCL:  testAccEnhancedSecurityNonDefaultNoRedirectAccessRulesHCL,
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
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
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

func testAccAccessPolicyConfig_emptyUserAndGroups(name string) string {
	const rdResourceName = "test_domain_for_empty_uag_policy"
	const appResourceName = "test_app_for_empty_uag_policy"
	const fqdn = "tf-acc-test-empty-uag-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for empty user_and_groups access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-empty-uag-policy",
		appType:      "web",
		description:  "Test app for empty user_and_groups access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_empty_uag",
		name:            name,
		description:     "Terraform acceptance test - policy with user_and_groups = {}",
		active:          false,
		appResourceName: appResourceName,
		priority:        996,
		accessRulesHCL:  testAccEmptyUserAndGroupsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

func testAccAccessPolicyConfig_translatedUserAndGroups(name string) string {
	const rdResourceName = "test_domain_for_translated_uag_policy"
	const appResourceName = "test_app_for_translated_uag_policy"
	const fqdn = "tf-acc-test-translated-uag-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for translated user_and_groups access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-translated-uag-policy",
		appType:      "web",
		description:  "Test app for translated user_and_groups access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_translated_uag",
		name:            name,
		description:     "Terraform acceptance test - policy with a populated (deprecated) user_and_groups",
		active:          false,
		appResourceName: appResourceName,
		priority:        995,
		accessRulesHCL:  testAccTranslatedUserAndGroupsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

// TestAccAccessPolicy_userAndGroupsTranslated verifies the translation behaviour for a
// populated user_and_groups map: the provider translates it into an extra
// TYPE_USERGROUP rule on the wire (userAndGroups itself is never sent), collapses
// that synthesized rule back out on Read so state matches the config, and produces
// no drift on a follow-up plan.
func TestAccAccessPolicy_userAndGroupsTranslated(t *testing.T) {
	name := "tf-acc-test-translated-uag-policy"
	fqdn := "tf-acc-test-translated-uag-policy-app.example.com"
	const resourceAddr = "citrixspa_access_policy.test_translated_uag"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with a populated user_and_groups map.
			{
				Config: testAccAccessPolicyConfig_translatedUserAndGroups(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI(resourceAddr),
					resource.TestCheckResourceAttr(resourceAddr, "name", name),
					resource.TestCheckResourceAttr(resourceAddr, "priority", "995"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.conditions.0.platform_filter", "PLATFORM_FILTER_PC"),
					// The deprecated map is carried forward from config verbatim.
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.conditions.0.user_and_groups.%", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.conditions.0.user_and_groups.SID:/citrix/S-1-5-21-777", "Engineering"),
					// The synthesized rule must be collapsed out of state: only the
					// rule declared in the config remains.
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.rules.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.rules.0.values.0", "SID:/citrix/S-1-1-0"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
				),
			},
			// Step 2: Re-apply the same config — must be a no-op. This is the real
			// assertion: the synthesized rule does not show up as perpetual drift.
			{
				Config:   testAccAccessPolicyConfig_translatedUserAndGroups(name),
				PlanOnly: true,
			},
			// Step 3: Import. There is no prior state to collapse against, so import
			// legitimately yields the translated TYPE_USERGROUP rule and no
			// user_and_groups map — that is the shape users should migrate to.
			// ImportStateVerify cannot be used here because the imported state
			// deliberately differs from the config-derived state above.
			{
				ResourceName: resourceAddr,
				ImportState:  true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported state, got %d", len(states))
					}
					attrs := states[0].Attributes
					if got := attrs["access_rules.0.rules.#"]; got != "2" {
						return fmt.Errorf("expected 2 rules after import (declared + translated), got %q", got)
					}
					if got := attrs["access_rules.0.conditions.0.user_and_groups.%"]; got != "" && got != "0" {
						return fmt.Errorf("expected no user_and_groups after import, got %q entries", got)
					}
					return nil
				},
			},
		},
	})
}

func testAccAccessPolicyConfig_collidingUserAndGroups(name string) string {
	const rdResourceName = "test_domain_for_colliding_uag_policy"
	const appResourceName = "test_app_for_colliding_uag_policy"
	const fqdn = "tf-acc-test-colliding-uag-policy-app.example.com"

	rdConfig := testAccRoutingDomainConfig(
		rdResourceName, fqdn, "internal", "web",
		"Test routing domain for colliding user_and_groups access policy", "enabled", "false", "[]",
	)

	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResourceName,
		name:         "tf-acc-test-app-for-colliding-uag-policy",
		appType:      "web",
		description:  "Test app for colliding user_and_groups access policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResourceName},
	})

	policyConfig := testAccAccessPolicyConfig(testAccessPolicyConfig{
		resourceName:    "test_colliding_uag",
		name:            name,
		description:     "Terraform acceptance test - user_and_groups colliding with a declared rule",
		active:          false,
		appResourceName: appResourceName,
		priority:        994,
		accessRulesHCL:  testAccCollidingUserAndGroupsAccessRulesHCL,
	})

	return rdConfig + appConfig + policyConfig
}

// TestAccAccessPolicy_userAndGroupsCollidesWithDeclaredRule guards the case
// where the configuration already declares exactly the rule user_and_groups
// translates to. The write path skips the append (the scope is already covered),
// so the read path must not treat the single matching rule it gets back as the
// synthesized one — dropping it would leave state with zero rules while the
// config declares one, which rules[] being Required turns into "Provider
// produced inconsistent result after apply".
func TestAccAccessPolicy_userAndGroupsCollidesWithDeclaredRule(t *testing.T) {
	name := "tf-acc-test-colliding-uag-policy"
	fqdn := "tf-acc-test-colliding-uag-policy-app.example.com"
	const resourceAddr = "citrixspa_access_policy.test_colliding_uag"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_collidingUserAndGroups(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI(resourceAddr),
					// The declared rule survives; it is not mistaken for a
					// rule the provider synthesized.
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.rules.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.rules.0.values.0", "SID:/citrix/S-1-5-21-778"),
					resource.TestCheckResourceAttr(resourceAddr, "access_rules.0.conditions.0.user_and_groups.%", "1"),
				),
			},
			// Re-apply must be a no-op.
			{
				Config:   testAccAccessPolicyConfig_collidingUserAndGroups(name),
				PlanOnly: true,
			},
		},
	})
}

// TestAccAccessPolicy_userAndGroupsOutOfBandRemoval guards the drift case: when
// the translated TYPE_USERGROUP rule is deleted outside Terraform, the scope the
// deprecated map describes is no longer enforced. Read must stop carrying the
// map forward from prior state so the next plan proposes restoring the rule
// instead of reporting the policy as up to date.
func TestAccAccessPolicy_userAndGroupsOutOfBandRemoval(t *testing.T) {
	name := "tf-acc-test-oob-uag-policy"
	fqdn := "tf-acc-test-translated-uag-policy-app.example.com"
	const resourceAddr = "citrixspa_access_policy.test_translated_uag"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_translatedUserAndGroups(name),
				Check:  testAccCheckAccessPolicyExistsInAPI(resourceAddr),
			},
			{
				PreConfig: func() {
					testAccRemoveTranslatedUserGroupRule(t, name, "SID:/citrix/S-1-5-21-777")
				},
				Config:             testAccAccessPolicyConfig_translatedUserAndGroups(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccRemoveTranslatedUserGroupRule strips the TYPE_USERGROUP rule carrying
// token from the named policy, simulating an operator deleting it in the SPA
// Console.
func testAccRemoveTranslatedUserGroupRule(t *testing.T, policyName, token string) {
	t.Helper()
	ctx := context.Background()
	client, err := testAccCreateClient()
	if err != nil {
		t.Fatalf("failed to create API client: %s", err)
	}

	listed, err := client.GetAccessPolicies(ctx, 0, 100, policyName, "")
	if err != nil {
		t.Fatalf("failed to list access policies: %s", err)
	}
	var id string
	for _, p := range listed.Policies {
		if p.Name == policyName {
			id = p.ID
			break
		}
	}
	if id == "" {
		t.Fatalf("access policy %q not found", policyName)
	}

	policy, err := client.GetAccessPolicy(ctx, id)
	if err != nil {
		t.Fatalf("failed to read access policy %s: %s", id, err)
	}
	for i := range policy.AccessRules {
		kept := make([]Rule, 0, len(policy.AccessRules[i].Rules))
		for _, rule := range policy.AccessRules[i].Rules {
			if rule.Type == userGroupRuleType && slices.Contains(rule.Values, token) {
				continue
			}
			kept = append(kept, rule)
		}
		policy.AccessRules[i].Rules = kept
	}
	if err := client.UpdateAccessPolicy(ctx, id, policy); err != nil {
		t.Fatalf("failed to strip the translated rule from %s: %s", id, err)
	}
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
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_no_conditions"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_no_conditions", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_no_conditions", "active", "false"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_no_conditions", "priority", "997"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_no_conditions", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_no_conditions", "access_rules.0.name", "No Conditions Rule"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_no_conditions", "access_rules.0.conditions.#", "0"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_no_conditions", "id"),
				),
			},
			// Step 2: Import — verifies the empty conditions list round-trips.
			{
				ResourceName:      "citrixspa_access_policy.test_no_conditions",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAccessPolicy_priorityOmitted is a regression test for the
// Console-owned `priority` fields. Both the policy-level `priority` and the
// per-rule `access_rules[].priority` are assigned/normalized by the Console, so
// the provider declares them Optional + Computed. Omitting them from config must
// NOT fail with "Provider produced inconsistent result after apply"
// (.priority: was null, but now cty.NumberIntVal(0)), and a re-apply of the same
// config must produce an empty plan (no spurious drift).
func TestAccAccessPolicy_priorityOmitted(t *testing.T) {
	// checkPriorityPositive asserts a priority attribute parses to a positive
	// integer. When the priority is omitted the request must leave the key out so
	// the backend auto-assigns a value; a regression that sent literal 0 would
	// defeat that auto-assignment and fail this check.
	checkPriorityPositive := func(attrPath string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttrWith("citrixspa_access_policy.test_prio_omit", attrPath, func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("%s %q is not an integer: %w", attrPath, v, err)
			}
			if n <= 0 {
				return fmt.Errorf("%s must be auto-assigned to a positive value when omitted, got %d", attrPath, n)
			}
			return nil
		})
	}
	name := "tf-acc-test-priority-omitted-policy"
	fqdn := "tf-acc-test-priority-omitted-app.example.com"
	rdResource := "test_domain_for_prio_omit"
	appResource := "test_app_for_prio_omit"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for priority-omitted policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-test-app-for-prio-omit",
		appType:      "web",
		description:  "Test app for priority-omitted policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})
	// Both the policy-level priority and the access_rules[].priority are omitted
	// on purpose.
	policyConfig := fmt.Sprintf(`
resource "citrixspa_access_policy" "test_prio_omit" {
  name        = %[1]q
  description = "Terraform acceptance test - priority omitted"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [
    {
      name        = "Priority Omitted Rule"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

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
  ]
}
`, name, appResource)

	// Same policy with a sibling field changed (active) while the priority
	// fields stay omitted — exercises an update that touches the resource
	// without ever setting priority.
	policyConfigUpdated := fmt.Sprintf(`
resource "citrixspa_access_policy" "test_prio_omit" {
  name        = %[1]q
  description = "Terraform acceptance test - priority omitted (updated)"
  active      = true
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [
    {
      name        = "Priority Omitted Rule"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""

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
  ]
}
`, name, appResource)

	config := rdConfig + appConfig + policyConfig
	configUpdated := rdConfig + appConfig + policyConfigUpdated

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with priority fields omitted — the Console-assigned
			// values must be adopted into state as computed values.
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_prio_omit"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_omit", "name", name),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_prio_omit", "priority"),
					checkPriorityPositive("priority"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_omit", "access_rules.#", "1"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_prio_omit", "access_rules.0.priority"),
					checkPriorityPositive("access_rules.0.priority"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_prio_omit", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the same config yields no drift.
			{
				Config:   config,
				PlanOnly: true,
			},
			// Step 3: Update a sibling field (active) while priority stays
			// omitted — the apply must succeed without an inconsistent-result
			// error and the computed priorities must remain set.
			{
				Config: configUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_omit", "active", "true"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_prio_omit", "priority"),
					checkPriorityPositive("priority"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_prio_omit", "access_rules.0.priority"),
					checkPriorityPositive("access_rules.0.priority"),
				),
			},
			// Step 4: Idempotency after the update.
			{
				Config:   configUpdated,
				PlanOnly: true,
			},
		},
	})
}

// TestAccAccessPolicy_priorityReorderOmitted covers a policy with multiple rules
// whose priorities are omitted, then reorders the rules. Because the nested
// priority no longer copies prior state by list index, a reorder must apply
// cleanly and remain idempotent.
func TestAccAccessPolicy_priorityReorderOmitted(t *testing.T) {
	name := "tf-acc-test-priority-reorder-policy"
	fqdn := "tf-acc-test-priority-reorder-app.example.com"
	rdResource := "test_domain_for_prio_reorder"
	appResource := "test_app_for_prio_reorder"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for priority-reorder policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-test-app-for-prio-reorder",
		appType:      "web",
		description:  "Test app for priority-reorder policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	rule := func(rName string) string {
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
          values     = ["Everyone"]
        }
      ]
    }`, rName)
	}
	policy := func(rulesBody string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_prio_reorder" {
  name        = %[1]q
  description = "Terraform acceptance test - priority reorder"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [%[3]s
  ]
}
`, name, appResource, rulesBody)
	}

	configAB := rdConfig + appConfig + policy(rule("rule-a")+","+rule("rule-b"))
	configBA := rdConfig + appConfig + policy(rule("rule-b")+","+rule("rule-a"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create two rules with priority omitted.
			{
				Config: configAB,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_prio_reorder"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_reorder", "access_rules.#", "2"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_reorder", "access_rules.0.name", "rule-a"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_reorder", "access_rules.1.name", "rule-b"),
				),
			},
			// Step 2: Reorder the rules — must apply without error.
			{
				Config: configBA,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_reorder", "access_rules.0.name", "rule-b"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_reorder", "access_rules.1.name", "rule-a"),
				),
			},
			// Step 3: Idempotency after reorder.
			{
				Config:   configBA,
				PlanOnly: true,
			},
		},
	})
}

// TestAccAccessPolicy_insertRuleAtFront covers inserting a new rule at the head
// of an existing access_rules list. Because rules map positionally, inserting a
// rule shifts every following rule to a new index. A genuinely new rule (no
// prior-state counterpart) must have its computed id/metadata planned as unknown
// so the backend-assigned values are accepted; otherwise the apply fails with
// "Provider produced inconsistent result after apply". Shifted existing rules
// keep their id.
func TestAccAccessPolicy_insertRuleAtFront(t *testing.T) {
	name := "tf-acc-test-insert-front-policy"
	fqdn := "tf-acc-test-insert-front-app.example.com"
	rdResource := "test_domain_for_insert_front"
	appResource := "test_app_for_insert_front"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for insert-front policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-test-app-for-insert-front",
		appType:      "web",
		description:  "Test app for insert-front policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	rule := func(rName string) string {
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
          values     = ["Everyone"]
        }
      ]
    }`, rName)
	}
	policy := func(rulesBody string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_insert_front" {
  name        = %[1]q
  description = "Terraform acceptance test - insert rule at front"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [%[3]s
  ]
}
`, name, appResource, rulesBody)
	}

	configTwo := rdConfig + appConfig + policy(rule("rule-a")+","+rule("rule-b"))
	configInserted := rdConfig + appConfig + policy(rule("rule-c")+","+rule("rule-a")+","+rule("rule-b"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create two rules.
			{
				Config: configTwo,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_insert_front"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.#", "2"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.0.name", "rule-a"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.1.name", "rule-b"),
				),
			},
			// Step 2: Insert a new rule at the front — must apply without an
			// inconsistent-result error.
			{
				Config: configInserted,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.#", "3"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.0.name", "rule-c"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.1.name", "rule-a"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_insert_front", "access_rules.2.name", "rule-b"),
				),
			},
			// Step 3: Idempotency after the insert.
			{
				Config:   configInserted,
				PlanOnly: true,
			},
		},
	})
}

// TestAccAccessPolicy_ruleIdRetainedOnInPlaceEdit verifies that editing a rule
// in place (here, toggling active) does not churn its backend-assigned id. Only
// genuinely new rules get their id planned as unknown; an existing rule keeps
// its identity across content edits.
func TestAccAccessPolicy_ruleIdRetainedOnInPlaceEdit(t *testing.T) {
	name := "tf-acc-test-id-retain-policy"
	fqdn := "tf-acc-test-id-retain-app.example.com"
	rdResource := "test_domain_for_id_retain"
	appResource := "test_app_for_id_retain"
	resourceName := "citrixspa_access_policy.test_id_retain"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for id-retain policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-test-app-for-id-retain",
		appType:      "web",
		description:  "Test app for id-retain policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})

	policy := func(active string) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_id_retain" {
  name        = %[1]q
  description = "Terraform acceptance test - rule id retained on in-place edit"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [
    {
      name        = "rule-a"
      active      = %[3]s
      access      = "ACCESS_ALLOW"
      description = ""

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
  ]
}
`, name, appResource, active)
	}

	var ruleID string
	captureID := func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		ruleID = rs.Primary.Attributes["access_rules.0.id"]
		if ruleID == "" {
			return fmt.Errorf("access_rules.0.id was empty after create")
		}
		return nil
	}
	assertSameID := func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		got := rs.Primary.Attributes["access_rules.0.id"]
		if got != ruleID {
			return fmt.Errorf("rule id changed on in-place edit: was %q, now %q", ruleID, got)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create one active rule and capture its backend id.
			{
				Config: rdConfig + appConfig + policy("true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_rules.0.active", "true"),
					captureID,
				),
			},
			// Step 2: Toggle active in place — the rule id must be unchanged.
			{
				Config: rdConfig + appConfig + policy("false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_rules.0.active", "false"),
					assertSameID,
				),
			},
		},
	})
}

// TestAccAccessPolicy_priorityExplicit verifies that an explicitly set nested
// priority is stored verbatim and that changing it is reflected in state.
func TestAccAccessPolicy_priorityExplicit(t *testing.T) {
	name := "tf-acc-test-priority-explicit-policy"
	fqdn := "tf-acc-test-priority-explicit-app.example.com"
	rdResource := "test_domain_for_prio_explicit"
	appResource := "test_app_for_prio_explicit"

	rdConfig := testAccRoutingDomainConfig(
		rdResource, fqdn, "internal", "web",
		"Test routing domain for priority-explicit policy", "enabled", "false", "[]",
	)
	appConfig := testAccApplicationConfig(testAppConfig{
		resourceName: appResource,
		name:         "tf-acc-test-app-for-prio-explicit",
		appType:      "web",
		description:  "Test app for priority-explicit policy",
		url:          "https://" + fqdn,
		relatedURLs:  []string{"*." + fqdn},
		dependsOn:    []string{"citrixspa_routing_domain." + rdResource},
	})
	policy := func(prio int) string {
		return fmt.Sprintf(`
resource "citrixspa_access_policy" "test_prio_explicit" {
  name        = %[1]q
  description = "Terraform acceptance test - priority explicit"
  active      = false
  apps        = [citrixspa_application.%[2]s.id]

  access_rules = [
    {
      name        = "Explicit Priority Rule"
      active      = true
      access      = "ACCESS_ALLOW"
      description = ""
      priority    = %[3]d

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
  ]
}
`, name, appResource, prio)
	}

	config10 := rdConfig + appConfig + policy(10)
	config20 := rdConfig + appConfig + policy(20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with an explicit nested priority — stored verbatim.
			{
				Config: config10,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_prio_explicit"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_explicit", "access_rules.0.priority", "10"),
				),
			},
			// Step 2: Idempotency.
			{
				Config:   config10,
				PlanOnly: true,
			},
			// Step 3: Change the explicit priority — the new value is reflected.
			{
				Config: config20,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_prio_explicit", "access_rules.0.priority", "20"),
				),
			},
		},
	})
}

// TestAccAccessPolicy_emptyUserAndGroups reproduces the scenario where writing a
// condition with user_and_groups = {} (an explicit empty map) caused:
//
//	Provider produced inconsistent result after apply:
//	.access_rules[0].conditions[0].user_and_groups: was cty.MapValEmpty(cty.String), but now null.
//
// The root cause was that the Read method converted an empty user_and_groups map
// to null instead of preserving the empty map the config declared. The fix
// mirrors the prior state so an explicit {} round-trips as {} (and null stays
// null). A clean second plan verifies there is no spurious diff.
func TestAccAccessPolicy_emptyUserAndGroups(t *testing.T) {
	name := "tf-acc-test-empty-uag-policy"
	fqdn := "tf-acc-test-empty-uag-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with user_and_groups = {} — previously failed with
			// "was cty.MapValEmpty(cty.String), but now null".
			{
				Config: testAccAccessPolicyConfig_emptyUserAndGroups(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_empty_uag"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "active", "false"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "priority", "996"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "access_rules.0.name", "Empty UserAndGroups Rule"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "access_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "access_rules.0.conditions.0.platform_filter", "PLATFORM_FILTER_PC"),
					// The empty map must be preserved (present, with 0 elements), not null.
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_empty_uag", "access_rules.0.conditions.0.user_and_groups.%", "0"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_empty_uag", "id"),
				),
			},
			// Step 2: Re-apply the same config — must be a no-op (no spurious diff
			// from {} vs null).
			{
				Config:   testAccAccessPolicyConfig_emptyUserAndGroups(name),
				PlanOnly: true,
			},
			// Step 3: Import — verifies the empty map round-trips.
			{
				ResourceName:      "citrixspa_access_policy.test_empty_uag",
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
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_basic"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "description", "Terraform acceptance test - basic access policy"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "active", "false"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "priority", "999"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.name", "Default Rule"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.access", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.access_native", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.tag_source", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.tag_key", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.values.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.values.0", "SID:/citrix/S-1-1-0"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_basic", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan
			{
				Config:   testAccAccessPolicyConfig_basic(name),
				PlanOnly: true,
			},
			// Step 3: Update - change description, enable active, lower priority
			{
				Config: testAccAccessPolicyConfig_basicUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_basic"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "description", "Terraform acceptance test - basic access policy UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "active", "true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "priority", "998"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.name", "Default Rule"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.access", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.access_native", "ACCESS_DENY"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.values.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_basic", "access_rules.0.rules.0.values.0", "SID:/citrix/S-1-1-0"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_basic", "id"),
				),
			},
			// Step 4: Import by ID and verify state matches
			{
				ResourceName:      "citrixspa_access_policy.test_basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAccessPolicy_omittedOptionalStrings verifies that omitting the optional
// string attributes entirely (policy description, access-rule description, and
// rule tag_source/tag_key) applies cleanly and yields an empty follow-up plan.
// The backend returns "" for these fields; Optional+Computed with a static
// stringdefault.StaticString("") default plans "" for a null config, keeping a
// null config and a "" state equivalent, so there is no post-apply
// inconsistency or permadiff. Removing Computed or the "" default would
// reintroduce that failure and fail Step 2.
func TestAccAccessPolicy_omittedOptionalStrings(t *testing.T) {
	name := "tf-acc-test-omitted-policy"
	fqdn := fmt.Sprintf("%s-app.example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with all optional strings omitted
			{
				Config: testAccAccessPolicyConfig_omittedOptionalStrings(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_omitted"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_omitted", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_omitted", "description", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_omitted", "access_rules.0.description", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_omitted", "access_rules.0.rules.0.tag_source", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_omitted", "access_rules.0.rules.0.tag_key", ""),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_omitted", "id"),
				),
			},
			// Step 2: Re-apply the same omitted-fields config — the plan must be
			// empty. This is the regression guard for the null/"" equivalence.
			{
				Config:   testAccAccessPolicyConfig_omittedOptionalStrings(name),
				PlanOnly: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAccessPolicy_clearOptionalStringsByOmission verifies that removing a
// previously-set optional string from configuration clears it (sends "") rather
// than retaining the old value. This guards against the Optional+Computed
// "retain prior value" trap: the "" default must win when the config is omitted,
// including when a rule changes type in place from TYPE_TAG to TYPE_USERGROUP and
// its tag_source/tag_key must be cleared.
func TestAccAccessPolicy_clearOptionalStringsByOmission(t *testing.T) {
	name := "tf-acc-test-clear-policy"
	fqdn := fmt.Sprintf("%s-app.example.com", name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: create with every optional string set to a non-empty value.
			// rules[1] is a TYPE_TAG rule with tag_source/tag_key set.
			{
				Config: testAccAccessPolicyConfig_clearScenario(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_clear"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "description", "Policy description set"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.description", "Access rule description"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.#", "2"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.1.type", "TYPE_TAG"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.1.tag_source", "ITM"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.1.tag_key", "location-geo-country-isocode"),
				),
			},
			// Step 2: omit the optional strings and transition rules[1] in place
			// from TYPE_TAG to TYPE_USERGROUP. The policy/access-rule descriptions
			// and the transitioned rule's tag_source/tag_key must clear to "",
			// not retain their prior values.
			{
				Config: testAccAccessPolicyConfig_clearScenario(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "description", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.description", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.#", "2"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.1.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.1.tag_source", ""),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_clear", "access_rules.0.rules.1.tag_key", ""),
				),
			},
			// Step 3: re-apply the omitted config — plan must be empty
			{
				Config:   testAccAccessPolicyConfig_clearScenario(name, false),
				PlanOnly: true,
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
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_with_rules"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "description", "Terraform acceptance test - policy with access rules"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "active", "false"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "priority", "998"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.name", "Allow Rule"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.access", "ACCESS_ALLOW"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.conditions.0.platform_filter", "PLATFORM_FILTER_ANY"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.redirect_sbs", "false"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings._browserV1", "embeddedBrowser"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.watermarkV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.downloadV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.uploadV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.clipboardV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.printingV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.keyLoggingV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.screenCaptureV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.proxyTrafficV1", "direct"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.rules.#", "3"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_with_rules", "id"),
				),
			},
			// Step 2: Update - change description, enable active, lower priority
			{
				Config: testAccAccessPolicyConfig_complexUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_with_rules"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "name", name),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "description", "Terraform acceptance test - policy with access rules UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "active", "true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "priority", "997"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.name", "Allow Rule"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.priority", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.active", "true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.access", "ACCESS_ALLOW"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.conditions.0.platform_filter", "PLATFORM_FILTER_ANY"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.redirect_sbs", "false"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings._browserV1", "embeddedBrowser"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.watermarkV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.downloadV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.uploadV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.clipboardV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.printingV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.keyLoggingV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.screenCaptureV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.restrictions.enhanced_security_settings.proxyTrafficV1", "direct"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.rules.#", "3"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.rules.0.type", "TYPE_USERGROUP"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_with_rules", "access_rules.0.rules.0.operator", "OPERATOR_IN"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_with_rules", "id"),
				),
			},
			// Step 3: Import by ID and verify state matches
			{
				ResourceName:      "citrixspa_access_policy.test_with_rules",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAccessPolicy_enhancedSecurityDefaultsPersist is a regression test for
// the case where enhanced_security_settings keys are set to their backend
// default value ("enabled" for download/upload/clipboard/printing). The API
// omits default-valued keys from the create/read response; without provider
// reconciliation, apply fails with "Provider produced inconsistent result after
// apply ... element ... has vanished". The provider must retain the configured
// keys, so both apply and a follow-up empty plan (idempotency) succeed.
func TestAccAccessPolicy_enhancedSecurityDefaultsPersist(t *testing.T) {
	name := "tf-acc-test-ess-defaults-policy"
	fqdn := "tf-acc-test-ess-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — apply must not fail with "has vanished".
			{
				Config: testAccAccessPolicyConfig_enhancedSecurityDefaults(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_ess_defaults"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_defaults", "access_rules.0.restrictions.enhanced_security_settings._browserV1", "embeddedBrowser"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_defaults", "access_rules.0.restrictions.enhanced_security_settings.downloadV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_defaults", "access_rules.0.restrictions.enhanced_security_settings.uploadV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_defaults", "access_rules.0.restrictions.enhanced_security_settings.clipboardV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_defaults", "access_rules.0.restrictions.enhanced_security_settings.printingV1", "enabled"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_ess_defaults", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan.
			{
				Config:   testAccAccessPolicyConfig_enhancedSecurityDefaults(name),
				PlanOnly: true,
			},
			// Step 3: Import by ID and verify state matches.
			//
			// Only the default-valued keys the API omits on a fresh import
			// (download/upload/clipboard/printing = "enabled") plus the map count
			// (%) are ignored: a fresh import has no prior state to reconcile
			// against, so those keys are absent from the imported state and the
			// one-time diff self-heals on the next apply. Every other key (e.g.
			// _browserV1) is still verified, so a regression there fails the test.
			{
				ResourceName:      "citrixspa_access_policy.test_ess_defaults",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"access_rules.0.restrictions.enhanced_security_settings.%",
					"access_rules.0.restrictions.enhanced_security_settings.downloadV1",
					"access_rules.0.restrictions.enhanced_security_settings.uploadV1",
					"access_rules.0.restrictions.enhanced_security_settings.clipboardV1",
					"access_rules.0.restrictions.enhanced_security_settings.printingV1",
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAccessPolicy_enhancedSecurityAllDefaultsOmitted is a regression test
// for the case where EVERY enhanced_security_settings key is set to its backend
// default and no non-default key is present. The API then omits the entire
// restrictions object from its response, so a positional/in-block reconciliation
// is skipped and apply fails with "Provider produced inconsistent result after
// apply ... was cty.ObjectVal(...), but now null". The provider must rebuild the
// restrictions block from prior config so both apply and a follow-up empty plan
// (idempotency) succeed.
func TestAccAccessPolicy_enhancedSecurityAllDefaultsOmitted(t *testing.T) {
	name := "tf-acc-test-ess-alldef-policy"
	fqdn := "tf-acc-test-ess-alldef-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — apply must not fail with the whole restrictions
			// block turning null.
			{
				Config: testAccAccessPolicyConfig_enhancedSecurityAllDefaults(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_ess_alldef"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_alldef", "access_rules.0.restrictions.enhanced_security_settings.downloadV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_alldef", "access_rules.0.restrictions.enhanced_security_settings.uploadV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_alldef", "access_rules.0.restrictions.enhanced_security_settings.clipboardV1", "enabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_alldef", "access_rules.0.restrictions.enhanced_security_settings.printingV1", "enabled"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_ess_alldef", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan.
			{
				Config:   testAccAccessPolicyConfig_enhancedSecurityAllDefaults(name),
				PlanOnly: true,
			},
		},
	})
}

// TestAccAccessPolicy_enhancedSecurityRedirectOnlyOmitted is a regression test
// for a restrictions block that sets ONLY redirect_sbs=false and leaves
// enhanced_security_settings null. Everything is at its default, so the API
// omits the whole restrictions object; the reconstruction branch must fire even
// with a null prior ESS map and preserve that null map (not force an empty one),
// otherwise apply fails with "Provider produced inconsistent result after apply
// ... was cty.ObjectVal(...), but now null".
func TestAccAccessPolicy_enhancedSecurityRedirectOnlyOmitted(t *testing.T) {
	name := "tf-acc-test-ess-redirect-policy"
	fqdn := "tf-acc-test-ess-redirect-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — apply must not fail with the whole restrictions
			// block turning null.
			{
				Config: testAccAccessPolicyConfig_enhancedSecurityRedirectOnly(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_ess_redirect"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_redirect", "access_rules.0.restrictions.redirect_sbs", "false"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_ess_redirect", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan.
			{
				Config:   testAccAccessPolicyConfig_enhancedSecurityRedirectOnly(name),
				PlanOnly: true,
			},
		},
	})
}

// TestAccAccessPolicy_enhancedSecurityRedirectTrueOmitted is a regression test
// for a restrictions block with a NON-default redirect_sbs=true and a null
// enhanced_security_settings map. The API keeps the block (redirect_sbs is
// non-default) but returns no ESS keys, so state mapping must preserve the null
// ESS map rather than forcing an empty {}; otherwise create fails with
// "Provider produced inconsistent result after apply ... enhanced_security_
// settings: was null, but now cty.MapValEmpty(cty.String)".
func TestAccAccessPolicy_enhancedSecurityRedirectTrueOmitted(t *testing.T) {
	name := "tf-acc-test-ess-redirect-true-policy"
	fqdn := "tf-acc-test-ess-redirect-true-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — apply must not fail with the ESS map turning
			// from null into an empty map.
			{
				Config: testAccAccessPolicyConfig_enhancedSecurityRedirectTrue(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_ess_redirect_true"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_redirect_true", "access_rules.0.restrictions.redirect_sbs", "true"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_ess_redirect_true", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan.
			{
				Config:   testAccAccessPolicyConfig_enhancedSecurityRedirectTrue(name),
				PlanOnly: true,
			},
			// Step 3: Import — a redirect_sbs-only block (null ESS) must import
			// with enhanced_security_settings as null, not an empty {}. On a fresh
			// import there is no prior map to preserve, so the API's omitted ESS
			// field must map to null; recording {} would surface as a spurious
			// post-import diff. No ignore is needed here.
			{
				ResourceName:      "citrixspa_access_policy.test_ess_redirect_true",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAccessPolicy_enhancedSecurityNonDefaultNoRedirectSBS is a regression
// test for a restrictions block that sets NON-default enhanced_security_settings
// (so the API keeps the block) while OMITTING redirect_sbs. redirect_sbs is
// Optional+Computed with a static false default, so an omitted value resolves to
// false at plan time and the API echoes false back — create and re-plan must be
// clean, and a fresh import must not show a spurious redirect_sbs diff. The ESS
// keys are all non-default and round-trip exactly as returned, so redirect_sbs
// is the only variable under test.
func TestAccAccessPolicy_enhancedSecurityNonDefaultNoRedirectSBS(t *testing.T) {
	name := "tf-acc-test-ess-nodefault-noredirect-policy"
	fqdn := "tf-acc-test-ess-nodefault-noredirect-policy-app.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — apply must not fail on redirect_sbs.
			{
				Config: testAccAccessPolicyConfig_enhancedSecurityNonDefaultNoRedirect(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessPolicyExistsInAPI("citrixspa_access_policy.test_ess_nodefault_noredirect"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_nodefault_noredirect", "access_rules.0.restrictions.enhanced_security_settings.keyLoggingV1", "disabled"),
					resource.TestCheckResourceAttr("citrixspa_access_policy.test_ess_nodefault_noredirect", "access_rules.0.restrictions.redirect_sbs", "false"),
					resource.TestCheckResourceAttrSet("citrixspa_access_policy.test_ess_nodefault_noredirect", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan.
			{
				Config:   testAccAccessPolicyConfig_enhancedSecurityNonDefaultNoRedirect(name),
				PlanOnly: true,
			},
			// Step 3: Fresh import — the imported state must match, i.e. an
			// omitted redirect_sbs on a rule with non-default ESS does not
			// produce a phantom diff. All ESS keys are non-default, so none are
			// omitted by the API and the map verifies exactly.
			{
				ResourceName:      "citrixspa_access_policy.test_ess_nodefault_noredirect",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAccessPolicy_restrictionsDriftNotMasked verifies that out-of-band
// removal of a NON-default restrictions block (redirect_sbs=true) is detected
// and corrected rather than masked by the reconstruction branch. The API omits
// the whole restrictions object once every setting is default AND once it is
// removed entirely, so reconstruction is gated on the prior block being
// semantically all-default. A non-default prior block must NOT be rebuilt from
// state; instead the missing block surfaces as drift and is re-applied.
func TestAccAccessPolicy_restrictionsDriftNotMasked(t *testing.T) {
	name := "tf-acc-test-restrictions-drift"
	fqdn := "tf-acc-test-ess-redirect-true-policy-app.example.com"
	const resourceName = "citrixspa_access_policy.test_ess_redirect_true"

	var policyID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccCleanupRoutingDomain(fqdn) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create redirect_sbs=true and capture the policy ID.
			{
				Config: testAccAccessPolicyConfig_enhancedSecurityRedirectTrue(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_rules.0.restrictions.redirect_sbs", "true"),
					testAccCaptureResourceID(resourceName, &policyID),
				),
			},
			// Step 2: Remove the restrictions block out of band, then re-apply the
			// same config. Because the prior block is non-default, it is not
			// reconstructed from state — the drift is detected and re-applied, so
			// the API ends up with restrictions.redirect_sbs=true again.
			{
				PreConfig: func() {
					if err := testAccRemoveAccessPolicyRestrictionsByID(policyID); err != nil {
						t.Fatalf("failed to remove restrictions out of band: %v", err)
					}
				},
				Config: testAccAccessPolicyConfig_enhancedSecurityRedirectTrue(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "access_rules.0.restrictions.redirect_sbs", "true"),
					testAccCheckAccessPolicyRestrictionsRedirectInAPI(resourceName, true),
				),
			},
			// Step 3: Idempotency — no drift remains.
			{
				Config:   testAccAccessPolicyConfig_enhancedSecurityRedirectTrue(name),
				PlanOnly: true,
			},
		},
	})
}

// testAccCaptureResourceID stores the Terraform-state ID of a resource into out.
func testAccCaptureResourceID(resourceName string, out *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		*out = rs.Primary.Attributes["id"]
		return nil
	}
}

// testAccRemoveAccessPolicyRestrictionsByID removes the restrictions block from
// every access rule of the given policy directly via the API (out-of-band drift).
func testAccRemoveAccessPolicyRestrictionsByID(id string) error {
	client, err := testAccCreateClient()
	if err != nil {
		return err
	}
	p, err := client.GetAccessPolicy(context.Background(), id)
	if err != nil {
		return err
	}
	for i := range p.AccessRules {
		p.AccessRules[i].Restrictions = nil
	}
	return client.UpdateAccessPolicy(context.Background(), id, p)
}

// testAccCheckAccessPolicyRestrictionsRedirectInAPI asserts, by reading the
// policy back from the API, that at least one access rule carries a restrictions
// block with the expected redirect_sbs value. Reading from the API (not state)
// proves the block was actually re-applied to the backend and drift was not
// merely masked in state.
func testAccCheckAccessPolicyRestrictionsRedirectInAPI(resourceName string, wantRedirect bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		id := rs.Primary.Attributes["id"]
		client, err := testAccCreateClient()
		if err != nil {
			return err
		}
		p, err := client.GetAccessPolicy(context.Background(), id)
		if err != nil {
			return err
		}
		for _, ar := range p.AccessRules {
			if ar.Restrictions != nil && ar.Restrictions.RedirectSBS == wantRedirect {
				return nil
			}
		}
		return fmt.Errorf("expected an access rule with restrictions.redirect_sbs=%t in the API, but none was present (drift was masked, not re-applied)", wantRedirect)
	}
}

// TestAccessPolicyEmptyDescriptionIsSent ensures an emptied description is
// still transmitted. When description is set to "", the update payload must
// carry "description":"" so the API clears any drifted value. Regression test
// for the `omitempty` bug that dropped an empty description from the request.
func TestAccessPolicyEmptyDescriptionIsSent(t *testing.T) {
	policy := &AccessPolicy{
		Name:        "DEV",
		Description: "", // user wants to clear the description
		Active:      true,
	}

	body, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	if !strings.Contains(string(body), `"description":""`) {
		t.Fatalf("empty description must be sent as \"description\":\"\" so the API clears it.\npayload: %s", body)
	}
}
