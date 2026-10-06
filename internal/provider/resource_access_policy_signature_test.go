package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestRuleContentSignature_noCollisionOnSemanticFields asserts that two rules that
// differ in exactly one otherwise-excluded semantic field produce DISTINCT content
// signatures, so the prior-state match cannot swap their preserved name/priority on
// reorder.
func TestRuleContentSignature_noCollisionOnSemanticFields(t *testing.T) {
	ctx := context.Background()

	base := func() AccessRuleResourceModel {
		return AccessRuleResourceModel{
			Name:     types.StringValue("ignored"),
			Priority: types.Int64Value(1),
			Active:   types.BoolValue(true),
			Access:   types.StringValue("ACCESS_ALLOW"),
		}
	}

	cases := []struct {
		name   string
		mutate func(a, b *AccessRuleResourceModel)
	}{
		{
			name: "advanced_settings.domain_overrides",
			mutate: func(a, b *AccessRuleResourceModel) {
				a.AdvancedSettings = &AdvancedSettingsResourceModel{DomainOverrides: []DomainOverrideResourceModel{{
					FQDN: types.StringValue("a.example.com"), Type: types.StringValue("internal"), LocationIDs: mustList(t, []string{"loc-1"}),
				}}}
				b.AdvancedSettings = &AdvancedSettingsResourceModel{DomainOverrides: []DomainOverrideResourceModel{{
					FQDN: types.StringValue("b.example.com"), Type: types.StringValue("internal"), LocationIDs: mustList(t, []string{"loc-1"}),
				}}}
			},
		},
		{
			name: "restrictions.redirect_sbs",
			mutate: func(a, b *AccessRuleResourceModel) {
				a.Restrictions = &RestrictionsResourceModel{RedirectSBS: types.BoolValue(true)}
				b.Restrictions = &RestrictionsResourceModel{RedirectSBS: types.BoolValue(false)}
			},
		},
		{
			name: "conditions.user_and_groups",
			mutate: func(a, b *AccessRuleResourceModel) {
				a.Conditions = []ConditionResourceModel{{
					PlatformFilter: types.StringValue("PLATFORM_FILTER_ANY"),
					UserAndGroups:  mustMap(t, map[string]string{"grp": "SID:/a"}),
				}}
				b.Conditions = []ConditionResourceModel{{
					PlatformFilter: types.StringValue("PLATFORM_FILTER_ANY"),
					UserAndGroups:  mustMap(t, map[string]string{"grp": "SID:/b"}),
				}}
			},
		},
		{
			name: "rules.metadata",
			mutate: func(a, b *AccessRuleResourceModel) {
				a.Rules = []RuleResourceModel{{
					Type:     types.StringValue("TYPE_USERGROUP"),
					Operator: types.StringValue("OPERATOR_IN"),
					Values:   mustList(t, []string{"SID:/x"}),
					Metadata: mustMap(t, map[string]string{"name": "SID:/x"}),
				}}
				b.Rules = []RuleResourceModel{{
					Type:     types.StringValue("TYPE_USERGROUP"),
					Operator: types.StringValue("OPERATOR_IN"),
					Values:   mustList(t, []string{"SID:/x"}),
					Metadata: mustMap(t, map[string]string{"name": "SID:/y"}),
				}}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := base(), base()
			tc.mutate(&a, &b)
			sa := ruleContentSignature(ctx, a, true)
			sb := ruleContentSignature(ctx, b, true)
			if sa == sb {
				t.Fatalf("collision: rules differing only in %s share signature %q", tc.name, sa)
			}
		})
	}
}

func mustMap(t *testing.T, m map[string]string) types.Map {
	t.Helper()
	v, diags := types.MapValueFrom(context.Background(), types.StringType, m)
	if diags.HasError() {
		t.Fatalf("map build: %v", diags)
	}
	return v
}

func mustList(t *testing.T, l []string) types.List {
	t.Helper()
	v, diags := types.ListValueFrom(context.Background(), types.StringType, l)
	if diags.HasError() {
		t.Fatalf("list build: %v", diags)
	}
	return v
}
