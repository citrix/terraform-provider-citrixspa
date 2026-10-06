package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// listWithUnknownElem builds a KNOWN list whose single element is unknown, as
// produced when `values` is wired to an unresolved resource output. The list
// itself is known, so accessRulesHasUnknown does not defer ModifyPlan.
func listWithUnknownElem() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown()})
}

// TestMatchAccessRules_unknownValueElemNoIDSwap guards against reordering named
// rules whose `values` come from an unresolved output. Their content signature is
// computed from an unresolvable list and cannot match the concrete prior-state
// signature, so without identity-aware handling the positional fallback would swap
// the two rules' backend ids. Matching must instead recover them by unique name.
func TestMatchAccessRules_unknownValueElemNoIDSwap(t *testing.T) {
	ctx := context.Background()

	rule := func(name, id string, vals types.List) AccessRuleResourceModel {
		return AccessRuleResourceModel{
			ID:   types.StringValue(id),
			Name: types.StringValue(name),
			Rules: []RuleResourceModel{{
				Type:     types.StringValue("TYPE_USERGROUP"),
				Operator: types.StringValue("OPERATOR_IN"),
				Values:   vals,
			}},
		}
	}

	// Prior state: concrete values. Plan: reordered [b, a] with values unknown.
	state := []AccessRuleResourceModel{
		rule("a", "idA", mustList(t, []string{"v1"})),
		rule("b", "idB", mustList(t, []string{"v2"})),
	}
	plan := []AccessRuleResourceModel{
		rule("b", "", listWithUnknownElem()),
		rule("a", "", listWithUnknownElem()),
	}

	match := matchAccessRulesToPrior(ctx, plan, state, nil)

	got0, got1 := "<nil>", "<nil>"
	if match[0] != nil {
		got0 = match[0].ID.ValueString()
	}
	if match[1] != nil {
		got1 = match[1].ID.ValueString()
	}
	if got0 != "idB" || got1 != "idA" {
		t.Fatalf("id swap under unknown values: plan[0]=b->%s (want idB), plan[1]=a->%s (want idA)", got0, got1)
	}
}

// TestMatchAccessRules_unknownValuesUnnamedNotPositional asserts that an unnamed
// rule with unknown identity fields is NOT positionally reconciled (it stays
// unmatched so its id is planned unknown), rather than being paired with the wrong
// prior rule on a reorder.
func TestMatchAccessRules_unknownValuesUnnamedNotPositional(t *testing.T) {
	ctx := context.Background()

	named := func(name, id string, vals types.List) AccessRuleResourceModel {
		return AccessRuleResourceModel{
			ID:   types.StringValue(id),
			Name: types.StringValue(name),
			Rules: []RuleResourceModel{{
				Type:     types.StringValue("TYPE_USERGROUP"),
				Operator: types.StringValue("OPERATOR_IN"),
				Values:   vals,
			}},
		}
	}
	anon := func(vals types.List) AccessRuleResourceModel {
		return AccessRuleResourceModel{
			ID:   types.StringNull(),
			Name: types.StringNull(),
			Rules: []RuleResourceModel{{
				Type:     types.StringValue("TYPE_USERGROUP"),
				Operator: types.StringValue("OPERATOR_IN"),
				Values:   vals,
			}},
		}
	}

	state := []AccessRuleResourceModel{
		named("a", "idA", mustList(t, []string{"v1"})),
		named("b", "idB", mustList(t, []string{"v2"})),
	}
	// Reordered, unnamed, unknown values -> nothing to match on safely.
	plan := []AccessRuleResourceModel{
		anon(listWithUnknownElem()),
		anon(listWithUnknownElem()),
	}

	match := matchAccessRulesToPrior(ctx, plan, state, nil)
	if match[0] != nil || match[1] != nil {
		t.Fatalf("ambiguous unknown rules must stay unmatched, got match[0]=%v match[1]=%v", match[0], match[1])
	}
}
