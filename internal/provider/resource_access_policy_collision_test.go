package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestRuleContentSignature_valuesCollisionSafe asserts that a rule whose values
// list is ["a,b"] and one whose list is ["a","b"] produce DISTINCT content
// signatures. A naive comma join would encode both as "a,b" and let the
// prior-state match swap their identities/name/priority on reorder.
func TestRuleContentSignature_valuesCollisionSafe(t *testing.T) {
	ctx := context.Background()
	one := arNamed("", "", types.BoolValue(true), mustList(t, []string{"a,b"}))
	two := arNamed("", "", types.BoolValue(true), mustList(t, []string{"a", "b"}))
	if ruleContentSignature(ctx, one, true) == ruleContentSignature(ctx, two, true) {
		t.Fatal("values [\"a,b\"] and [\"a\",\"b\"] must not share a content signature")
	}
}

// TestRuleContentSignature_metadataCollisionSafe asserts that a user_and_groups
// map {"a":"b,c=d"} and {"a":"b","c":"d"} produce DISTINCT content signatures. An
// unescaped "k=v" comma join would collide both to "a=b,c=d".
func TestRuleContentSignature_metadataCollisionSafe(t *testing.T) {
	ctx := context.Background()
	mk := func(m map[string]string) AccessRuleResourceModel {
		r := arNamed("", "", types.BoolValue(true), mustList(t, []string{"v"}))
		r.Conditions = []ConditionResourceModel{{
			PlatformFilter: types.StringValue(""),
			UserAndGroups:  mustMap(t, m),
		}}
		return r
	}
	one := mk(map[string]string{"a": "b,c=d"})
	two := mk(map[string]string{"a": "b", "c": "d"})
	if ruleContentSignature(ctx, one, true) == ruleContentSignature(ctx, two, true) {
		t.Fatal("maps {\"a\":\"b,c=d\"} and {\"a\":\"b\",\"c\":\"d\"} must not share a content signature")
	}
}

// TestRuleMetadataMatchKey_valuesCollisionSafe asserts the metadata match key is
// also collision-free across the same values ambiguity.
func TestRuleMetadataMatchKey_valuesCollisionSafe(t *testing.T) {
	ctx := context.Background()
	one := RuleResourceModel{
		Type:     types.StringValue("TYPE_USERGROUP"),
		Operator: types.StringValue("OPERATOR_IN"),
		Values:   mustList(t, []string{"a,b"}),
	}
	two := RuleResourceModel{
		Type:     types.StringValue("TYPE_USERGROUP"),
		Operator: types.StringValue("OPERATOR_IN"),
		Values:   mustList(t, []string{"a", "b"}),
	}
	if ruleMetadataMatchKey(ctx, one) == ruleMetadataMatchKey(ctx, two) {
		t.Fatal("metadata match key must distinguish values [\"a,b\"] from [\"a\",\"b\"]")
	}
}

// TestMatchAccessRules_unknownLeafNotContentMatched asserts that a reordered plan
// rule with an unknown signature leaf (user_and_groups map unresolved from an
// output) is NOT content-matched to a prior rule that merely had a null map at
// that field. Both normalize to the same sentinel, so a content match would carry
// the wrong id/metadata; the rule must stay unmatched instead.
func TestMatchAccessRules_unknownLeafNotContentMatched(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	// Prior rule with a concrete (null) user_and_groups condition.
	prior := arNamed("", "idA", tru, mustList(t, []string{"v"}))
	prior.Conditions = []ConditionResourceModel{{
		PlatformFilter: types.StringValue(""),
		UserAndGroups:  types.MapNull(types.StringType),
	}}
	state := []AccessRuleResourceModel{prior}

	// Plan rule (reordered, unnamed) whose user_and_groups is unknown.
	planRule := arNamed("", "", tru, mustList(t, []string{"v"}))
	planRule.Conditions = []ConditionResourceModel{{
		PlatformFilter: types.StringValue(""),
		UserAndGroups:  types.MapUnknown(types.StringType),
	}}
	plan := []AccessRuleResourceModel{planRule}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if m[0] != nil {
		t.Fatalf("unknown-leaf rule must not content-match a null-leaf prior, got id=%s", idOf(m[0]))
	}
}

// TestAccessRuleModel_fieldsClassifiedForSignature guards against a future access
// rule field silently falling out of ruleContentSignature. Every exported field
// of AccessRuleResourceModel must be classified as either excluded from the
// content identity (Console-owned or deliberately excluded) or included in it.
// Adding a field forces a conscious choice here so a reorder-stability regression
// cannot ship unnoticed.
func TestAccessRuleModel_fieldsClassifiedForSignature(t *testing.T) {
	excluded := map[string]bool{
		"ID":           true, // Console-owned identity
		"Name":         true, // Console-owned, preserved separately
		"Priority":     true, // Console-owned, preserved separately
		"AccessNative": true, // deliberately excluded (see ruleContentSignature doc)
	}
	included := map[string]bool{
		"Access":           true,
		"Active":           true,
		"Description":      true,
		"AdvancedSettings": true,
		"Conditions":       true,
		"Restrictions":     true,
		"Rules":            true,
	}
	tp := reflect.TypeOf(AccessRuleResourceModel{})
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		if !f.IsExported() {
			continue
		}
		if !excluded[f.Name] && !included[f.Name] {
			t.Fatalf("field %q of AccessRuleResourceModel is unclassified for ruleContentSignature; "+
				"add it to the excluded or included set here, and (if semantic) to ruleContentSignature", f.Name)
		}
	}
}
