package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Deprecated conditions[].user_and_groups -> rules[] TYPE_USERGROUP -------
//
// User/group scope belongs in a rule with type TYPE_USERGROUP, which is what
// the SPA Console writes; the conditions[].userAndGroups field is retired and
// the provider never sends it.
//
// Rather than break existing configurations, the provider keeps accepting
// user_and_groups and translates it into an equivalent TYPE_USERGROUP rule on
// write — the same translation the Console performs. A plan-time warning tells
// the operator to move the scope into rules[].

const (
	userGroupRuleType     = "TYPE_USERGROUP"
	userGroupRuleOperator = "OPERATOR_IN"
)

// userGroupTokenPrefixes are the identity-token forms the SPA service resolves.
// The bare token "all_users" is accepted as well.
var userGroupTokenPrefixes = []string{"SID:/", "OID:/", "EMAIL:/"}

// userGroupScope is the merged, deterministic view of the deprecated
// user_and_groups maps declared on one access rule's conditions.
type userGroupScope struct {
	// Values holds the identity tokens, sorted so the emitted rule does not
	// change between runs (Go map iteration order is randomised).
	Values []string
	// Metadata maps a display name to the comma-joined, sorted list of the
	// tokens that carry it — the shape the Console expects and needs to render
	// the policy's edit page.
	Metadata map[string]string
}

// scopeFromMerged turns a merged token -> display name map into the
// deterministic scope both readers below hand to userGroupRuleFromScope. ok is
// false when there is nothing to translate.
func scopeFromMerged(merged map[string]string) (userGroupScope, bool) {
	if len(merged) == 0 {
		return userGroupScope{}, false
	}

	values := make([]string, 0, len(merged))
	byDisplayName := make(map[string][]string)
	for token, displayName := range merged {
		values = append(values, token)
		// An empty display name would produce an unusable metadata key, so the
		// token stands in for itself.
		if displayName == "" {
			displayName = token
		}
		byDisplayName[displayName] = append(byDisplayName[displayName], token)
	}
	sort.Strings(values)

	metadata := make(map[string]string, len(byDisplayName))
	for displayName, tokens := range byDisplayName {
		sort.Strings(tokens)
		metadata[displayName] = strings.Join(tokens, ",")
	}

	return userGroupScope{Values: values, Metadata: metadata}, true
}

// userAndGroupsHasUnknownElement reports whether a known user_and_groups map
// holds an unknown element. ElementsAs into map[string]string cannot represent
// that, so the readers below skip the condition rather than raise a conversion
// diagnostic during validation; the value is resolved by the time it matters.
func userAndGroupsHasUnknownElement(m types.Map) bool {
	for _, elem := range m.Elements() {
		if elem.IsUnknown() {
			return true
		}
	}
	return false
}

// userGroupScopeFromConditions merges every non-empty user_and_groups map on an
// access rule's conditions into a single scope. ok is false when there is
// nothing to translate, which is the common case (attribute absent, or the
// explicit empty map `{}`).
func userGroupScopeFromConditions(ctx context.Context, conditions []ConditionResourceModel) (userGroupScope, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// token -> display name, merged across all conditions.
	merged := make(map[string]string)
	for _, cond := range conditions {
		if cond.UserAndGroups.IsNull() || cond.UserAndGroups.IsUnknown() {
			continue
		}
		if userAndGroupsHasUnknownElement(cond.UserAndGroups) {
			continue
		}
		entries := make(map[string]string)
		diags.Append(cond.UserAndGroups.ElementsAs(ctx, &entries, false)...)
		if diags.HasError() {
			return userGroupScope{}, false, diags
		}
		for token, displayName := range entries {
			merged[token] = displayName
		}
	}

	scope, ok := scopeFromMerged(merged)
	return scope, ok, diags
}

// userGroupRuleFromScope builds the API rule that carries the translated scope.
func userGroupRuleFromScope(scope userGroupScope) Rule {
	metadata := make(map[string]interface{}, len(scope.Metadata))
	for k, v := range scope.Metadata {
		metadata[k] = v
	}
	return Rule{
		Type:      userGroupRuleType,
		Operator:  userGroupRuleOperator,
		TagSource: "",
		TagKey:    "",
		Values:    append([]string(nil), scope.Values...),
		Metadata:  metadata,
	}
}

// matchesUserGroupScope reports whether an API rule is the TYPE_USERGROUP rule
// that scope would produce. Values are compared as sets because the service is
// free to reorder them.
func matchesUserGroupScope(rule Rule, scope userGroupScope) bool {
	if rule.Type != userGroupRuleType || rule.Operator != userGroupRuleOperator {
		return false
	}
	if len(rule.Values) != len(scope.Values) {
		return false
	}
	actual := append([]string(nil), rule.Values...)
	sort.Strings(actual)
	for i, v := range scope.Values {
		if actual[i] != v {
			return false
		}
	}
	if len(rule.Metadata) != len(scope.Metadata) {
		return false
	}
	for k, want := range scope.Metadata {
		got, ok := rule.Metadata[k]
		if !ok || fmt.Sprintf("%v", got) != want {
			return false
		}
	}
	return true
}

// appendUserGroupRule appends the rule translated from the deprecated
// user_and_groups maps, unless the configuration already declares an equivalent
// TYPE_USERGROUP rule (in which case the scope is already covered and a second
// rule would only widen the policy).
func appendUserGroupRule(rules []Rule, scope userGroupScope) []Rule {
	for _, existing := range rules {
		if matchesUserGroupScope(existing, scope) {
			return rules
		}
	}
	return append(rules, userGroupRuleFromScope(scope))
}

// countScopeMatches reports how many API rules carry exactly scope.
func countScopeMatches(rules []Rule, scope userGroupScope) int {
	n := 0
	for _, rule := range rules {
		if matchesUserGroupScope(rule, scope) {
			n++
		}
	}
	return n
}

// countDeclaredScopeMatches reports how many rules the configuration itself
// declares that carry exactly scope. State mirrors the configuration for
// rules[] (it is Required, and the synthesized rule is stripped before state is
// written), so the prior state's rules answer the question.
func countDeclaredScopeMatches(ctx context.Context, rules []RuleResourceModel, scope userGroupScope) (int, diag.Diagnostics) {
	var diags diag.Diagnostics
	n := 0
	for _, rule := range rules {
		converted, convertDiags := ruleFromModel(ctx, rule)
		diags.Append(convertDiags...)
		if convertDiags.HasError() {
			continue
		}
		if matchesUserGroupScope(converted, scope) {
			n++
		}
	}
	return n, diags
}

// ruleFromModel converts a state rule into the API shape so it can be compared
// against a scope with matchesUserGroupScope.
func ruleFromModel(ctx context.Context, rule RuleResourceModel) (Rule, diag.Diagnostics) {
	var diags diag.Diagnostics

	var values []string
	if !rule.Values.IsNull() && !rule.Values.IsUnknown() {
		diags.Append(rule.Values.ElementsAs(ctx, &values, false)...)
	}

	metadata := make(map[string]interface{})
	if !rule.Metadata.IsNull() && !rule.Metadata.IsUnknown() {
		entries := make(map[string]string)
		diags.Append(rule.Metadata.ElementsAs(ctx, &entries, false)...)
		for k, v := range entries {
			metadata[k] = v
		}
	}

	return Rule{
		Type:      rule.Type.ValueString(),
		Operator:  rule.Operator.ValueString(),
		TagSource: rule.TagSource.ValueString(),
		TagKey:    rule.TagKey.ValueString(),
		Values:    values,
		Metadata:  metadata,
	}, diags
}

// reconcileSynthesizedUserGroupRule strips the TYPE_USERGROUP rule the write
// path synthesized from user_and_groups, so state keeps mirroring the
// configuration (rules[] is Required, not Computed).
//
// The write path only appends when the configuration does not already declare
// an equivalent rule, so the read path must mirror that decision rather than
// blindly dropping the first match: when the operator wrote the very same rule
// by hand, nothing was synthesized and dropping a match would delete their rule
// from state ("element 0 has vanished" on the next apply). Comparing the API's
// match count against the count the configuration declares settles it.
//
// Whether the service still carries the scope at all is a separate question,
// answered by countScopeMatches at the call site: the conditions are built
// before the rules are, and that answer is needed for both.
func reconcileSynthesizedUserGroupRule(apiRules []Rule, declared int, scope userGroupScope) []Rule {
	apiMatches := countScopeMatches(apiRules, scope)
	if apiMatches <= declared {
		// Nothing matched, or every match is one the configuration asked for and
		// the write path appended nothing.
		return apiRules
	}

	// Drop the surplus match — the synthesized rule.
	for i, rule := range apiRules {
		if !matchesUserGroupScope(rule, scope) {
			continue
		}
		trimmed := make([]Rule, 0, len(apiRules)-1)
		trimmed = append(trimmed, apiRules[:i]...)
		trimmed = append(trimmed, apiRules[i+1:]...)
		return trimmed
	}
	return apiRules
}

// isResolvableUserGroupToken reports whether a user_and_groups key looks like a
// directory token. The API itself accepts arbitrary strings here, so this is
// not an API-validity check: it flags values the SPA Console cannot resolve,
// which make its policy edit page fail to render.
func isResolvableUserGroupToken(token string) bool {
	if token == "all_users" {
		return true
	}
	for _, prefix := range userGroupTokenPrefixes {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}
	return false
}

// isServiceValidatedUserGroupToken reports whether the SPA service validates a
// token's contents rather than storing it verbatim. These are the forms that
// can fail the apply: an EMAIL:/ token is checked against the directory's email
// claims, and all_users is gated on a per-tenant feature flag.
func isServiceValidatedUserGroupToken(token string) bool {
	return token == "all_users" || strings.HasPrefix(token, "EMAIL:/")
}

// userGroupRuleHCL renders just the rule object the provider will synthesize.
// It deliberately stops short of a whole `rules = [...]` assignment: most
// policies already declare rules, and an operator who pasted a complete
// assignment would silently replace them.
func userGroupRuleHCL(scope userGroupScope) string {
	quoted := make([]string, 0, len(scope.Values))
	for _, v := range scope.Values {
		quoted = append(quoted, fmt.Sprintf("%q", v))
	}

	displayNames := make([]string, 0, len(scope.Metadata))
	for k := range scope.Metadata {
		displayNames = append(displayNames, k)
	}
	sort.Strings(displayNames)
	metaPairs := make([]string, 0, len(displayNames))
	for _, name := range displayNames {
		metaPairs = append(metaPairs, fmt.Sprintf("%q = %q", name, scope.Metadata[name]))
	}

	var b strings.Builder
	b.WriteString("{\n")
	b.WriteString(fmt.Sprintf("  type     = %q\n", userGroupRuleType))
	b.WriteString(fmt.Sprintf("  operator = %q\n", userGroupRuleOperator))
	b.WriteString(fmt.Sprintf("  values   = [%s]\n", strings.Join(quoted, ", ")))
	b.WriteString(fmt.Sprintf("  metadata = { %s }\n", strings.Join(metaPairs, ", ")))
	b.WriteString("}")
	return b.String()
}

// validateDeprecatedUserAndGroups warns, once per access rule, that
// conditions[].user_and_groups is deprecated, and shows the equivalent
// TYPE_USERGROUP rule the provider will send in its place. It also flags tokens
// whose contents the service checks, and tokens the SPA Console cannot resolve.
//
// Each access rule is evaluated independently: an error on one rule must not
// suppress the guidance for the rest, so the per-rule diagnostics decide
// whether to skip, not the accumulated set.
func validateDeprecatedUserAndGroups(ctx context.Context, data AccessPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	for i, ar := range data.AccessRules {
		scope, ok, scopeDiags := userGroupScopeFromConditions(ctx, ar.Conditions)
		diags.Append(scopeDiags...)
		if scopeDiags.HasError() || !ok {
			continue
		}

		condPath := path.Root("access_rules").AtListIndex(i).AtName("conditions")
		diags.AddAttributeWarning(
			condPath,
			"Deprecated: conditions[].user_and_groups",
			"conditions[].user_and_groups is deprecated and is no longer sent to the SPA "+
				"service, which is retiring the field. The provider translates it into an "+
				"equivalent TYPE_USERGROUP rule so this configuration keeps working, but "+
				"the attribute will be removed in a future release.\n\n"+
				"Add this rule to the access rule's existing rules[] list — keep the rules "+
				"already there — and then delete user_and_groups:\n\n"+userGroupRuleHCL(scope),
		)

		for _, token := range scope.Values {
			if isServiceValidatedUserGroupToken(token) {
				diags.AddAttributeWarning(
					condPath,
					"user_and_groups token is validated by the SPA service",
					fmt.Sprintf("%q is checked by the service when the translated "+
						"TYPE_USERGROUP rule is written: an EMAIL:/ token must match a "+
						"directory email claim, and \"all_users\" requires the "+
						"corresponding tenant feature to be enabled. If it does not, the "+
						"apply fails with an API error. Other token forms are stored "+
						"verbatim and are not validated.", token),
				)
				continue
			}
			if isResolvableUserGroupToken(token) {
				continue
			}
			diags.AddAttributeWarning(
				condPath,
				"Unresolvable user_and_groups identity token",
				fmt.Sprintf("%q is not a directory token (expected SID:/, OID:/ or "+
					"EMAIL:/, or \"all_users\"). The API stores it as written, so the "+
					"apply succeeds, but the SPA Console cannot resolve it and its policy "+
					"edit page fails to render for this policy.", token),
			)
		}
	}
	return diags
}
