package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These are pure, deterministic unit tests for the restrictions/enhanced-security
// reconciliation helpers. They run in CI without live credentials (unlike the
// TestAcc* acceptance tests, which are gated behind RUN_ACCEPTANCE_TESTS), and
// lock in the reconciliation behavior and the enhancedSecurityDefaults map.

// essMap is a small helper to build a non-null enhanced_security_settings map.
func essMap(t *testing.T, kv map[string]string) types.Map {
	t.Helper()
	elems := make(map[string]attr.Value, len(kv))
	for k, v := range kv {
		elems[k] = types.StringValue(v)
	}
	m, diags := types.MapValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("failed to build ESS map: %v", diags)
	}
	return m
}

func restrictionsRule(redirectSBS types.Bool, ess types.Map) *AccessRuleResourceModel {
	return &AccessRuleResourceModel{
		Restrictions: &RestrictionsResourceModel{
			RedirectSBS:              redirectSBS,
			EnhancedSecuritySettings: ess,
		},
	}
}

func TestEnhancedSecurityDefaults_Pinned(t *testing.T) {
	// Pin the map so an accidental edit (wrong key, wrong default) is caught.
	// clipboard/download/printing/upload/keyLogging/screenCapture default to
	// "enabled"; watermark and insecure_content_allowed_for_urls_v1 default to
	// "disabled". _browserV1 (always emitted) and proxyTrafficV1 (no enabled/
	// disabled default; only omitted when cleared to null) are intentionally
	// absent — neither has a reconcilable default.
	want := map[string]string{
		"clipboardV1":                          "enabled",
		"downloadV1":                           "enabled",
		"printingV1":                           "enabled",
		"uploadV1":                             "enabled",
		"keyLoggingV1":                         "enabled",
		"screenCaptureV1":                      "enabled",
		"watermarkV1":                          "disabled",
		"insecure_content_allowed_for_urls_v1": "disabled",
	}
	if len(enhancedSecurityDefaults) != len(want) {
		t.Fatalf("enhancedSecurityDefaults has %d keys, want %d: %v",
			len(enhancedSecurityDefaults), len(want), enhancedSecurityDefaults)
	}
	for k, v := range want {
		got, ok := enhancedSecurityDefaults[k]
		if !ok {
			t.Errorf("enhancedSecurityDefaults missing key %q", k)
			continue
		}
		if got != v {
			t.Errorf("enhancedSecurityDefaults[%q] = %q, want %q", k, got, v)
		}
	}
	for _, absent := range []string{"_browserV1", "proxyTrafficV1"} {
		if _, ok := enhancedSecurityDefaults[absent]; ok {
			t.Errorf("enhancedSecurityDefaults must not contain %q (no reconcilable default)", absent)
		}
	}
}

func TestPriorRestrictionsAllDefault(t *testing.T) {
	tests := []struct {
		name string
		rule *AccessRuleResourceModel
		want bool
	}{
		{
			name: "nil prior rule",
			rule: nil,
			want: false,
		},
		{
			name: "nil restrictions",
			rule: &AccessRuleResourceModel{Restrictions: nil},
			want: false,
		},
		{
			name: "redirect_sbs true is non-default",
			rule: restrictionsRule(types.BoolValue(true), types.MapNull(types.StringType)),
			want: false,
		},
		{
			name: "redirect_sbs false and null ESS map is all-default",
			rule: restrictionsRule(types.BoolValue(false), types.MapNull(types.StringType)),
			want: true,
		},
		{
			name: "redirect_sbs null and null ESS map is all-default",
			rule: restrictionsRule(types.BoolNull(), types.MapNull(types.StringType)),
			want: true,
		},
		{
			name: "all known-default ESS keys",
			rule: restrictionsRule(types.BoolValue(false), essMap(t, map[string]string{
				"clipboardV1": "enabled",
				"downloadV1":  "enabled",
				"watermarkV1": "disabled",
			})),
			want: true,
		},
		{
			name: "one non-default key",
			rule: restrictionsRule(types.BoolValue(false), essMap(t, map[string]string{
				"clipboardV1": "enabled",
				"downloadV1":  "disabled", // non-default
			})),
			want: false,
		},
		{
			name: "unknown key treated as non-default",
			rule: restrictionsRule(types.BoolValue(false), essMap(t, map[string]string{
				"bogusKeyV1": "enabled",
			})),
			want: false,
		},
		{
			name: "watermark non-default (enabled) is non-default",
			rule: restrictionsRule(types.BoolValue(false), essMap(t, map[string]string{
				"watermarkV1": "enabled",
			})),
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := priorRestrictionsAllDefault(tc.rule); got != tc.want {
				t.Errorf("priorRestrictionsAllDefault() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatchPriorAccessRule(t *testing.T) {
	prior := []AccessRuleResourceModel{
		{ID: types.StringValue("id-0"), Name: types.StringValue("rule-a")},
		{ID: types.StringValue("id-1"), Name: types.StringValue("rule-b")},
		{ID: types.StringNull(), Name: types.StringValue("rule-c")},
	}

	tests := []struct {
		name    string
		apiRule AccessRule
		idx     int
		wantIdx int // index into prior, or -1 for nil
	}{
		{
			name:    "match by ID takes precedence",
			apiRule: AccessRule{ID: "id-1", Name: "rule-a"}, // name would point elsewhere
			idx:     0,
			wantIdx: 1,
		},
		{
			name:    "match by name when ID absent",
			apiRule: AccessRule{Name: "rule-c"},
			idx:     0,
			wantIdx: 2,
		},
		{
			// Legitimate name fallback: the API rule carries an ID but the matching
			// prior rule has a null ID (Create/import re-read), so the ID-conflict
			// guard stays inert and name matching applies.
			name:    "match by name when prior ID is null",
			apiRule: AccessRule{ID: "id-missing", Name: "rule-c"},
			idx:     0,
			wantIdx: 2,
		},
		{
			// Out-of-band replacement: the API rule and the name-matching prior rule
			// both carry known IDs that differ, so they are treated as different
			// rules and are not paired (idx is out of range so positional cannot
			// rescue). Backend rule IDs are stable across updates, so this only
			// occurs when a rule was replaced out of band.
			name:    "name match blocked when both IDs known and differ",
			apiRule: AccessRule{ID: "id-missing", Name: "rule-b"},
			idx:     99,
			wantIdx: -1,
		},
		{
			name:    "positional fallback when no ID/name",
			apiRule: AccessRule{},
			idx:     2,
			wantIdx: 2,
		},
		{
			name:    "no match and out-of-range index returns nil",
			apiRule: AccessRule{ID: "id-x", Name: "rule-x"},
			idx:     99,
			wantIdx: -1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchPriorAccessRule(prior, tc.apiRule, tc.idx, map[int]bool{})
			if tc.wantIdx == -1 {
				if got != nil {
					t.Errorf("matchPriorAccessRule() = %+v, want nil", got)
				}
				return
			}
			if got != &prior[tc.wantIdx] {
				t.Errorf("matchPriorAccessRule() = %+v, want prior[%d] = %+v", got, tc.wantIdx, prior[tc.wantIdx])
			}
		})
	}
}

// TestMatchPriorAccessRule_DuplicateNames locks in that two rules sharing a
// non-empty name (schema permits non-unique names) are paired one-to-one in
// order rather than both resolving to the first prior rule. This mirrors the
// Create path, where planned rules carry unknown IDs so the name/positional
// passes do the disambiguation.
func TestMatchPriorAccessRule_DuplicateNames(t *testing.T) {
	prior := []AccessRuleResourceModel{
		{ID: types.StringNull(), Name: types.StringValue("dup")},
		{ID: types.StringNull(), Name: types.StringValue("dup")},
	}
	used := map[int]bool{}

	// First API rule with name "dup" claims prior[0].
	got0 := matchPriorAccessRule(prior, AccessRule{Name: "dup"}, 0, used)
	if got0 != &prior[0] {
		t.Fatalf("first match = %+v, want prior[0]", got0)
	}
	// Second API rule with the same name must claim prior[1], not prior[0] again.
	got1 := matchPriorAccessRule(prior, AccessRule{Name: "dup"}, 1, used)
	if got1 != &prior[1] {
		t.Fatalf("second match = %+v, want prior[1] (duplicate name must not reuse prior[0])", got1)
	}
}

func TestRestoreOmittedEnhancedSecurity(t *testing.T) {
	ctx := context.Background()

	t.Run("nil prior rule is a no-op", func(t *testing.T) {
		var diags diag.Diagnostics
		m := map[string]attr.Value{}
		restoreOmittedEnhancedSecurity(ctx, nil, m, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(m) != 0 {
			t.Errorf("expected no keys added, got %v", m)
		}
	})

	t.Run("null prior ESS map is a no-op", func(t *testing.T) {
		var diags diag.Diagnostics
		m := map[string]attr.Value{}
		prior := restrictionsRule(types.BoolNull(), types.MapNull(types.StringType))
		restoreOmittedEnhancedSecurity(ctx, prior, m, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(m) != 0 {
			t.Errorf("expected no keys added, got %v", m)
		}
	})

	t.Run("known-default key restored to literal default, not prior value", func(t *testing.T) {
		var diags diag.Diagnostics
		m := map[string]attr.Value{}
		// Prior stored downloadV1 as its default "enabled"; the API dropped it.
		prior := restrictionsRule(types.BoolNull(), essMap(t, map[string]string{
			"downloadV1": "enabled",
		}))
		restoreOmittedEnhancedSecurity(ctx, prior, m, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		got, ok := m["downloadV1"].(types.String)
		if !ok {
			t.Fatalf("downloadV1 not restored: %v", m)
		}
		if got.ValueString() != "enabled" {
			t.Errorf("downloadV1 = %q, want default \"enabled\"", got.ValueString())
		}
	})

	t.Run("known-default key restored to default even if prior held non-default", func(t *testing.T) {
		// This is the drift-detection guarantee: Read restores the literal
		// default so out-of-band drift is surfaced, not the stale prior value.
		var diags diag.Diagnostics
		m := map[string]attr.Value{}
		prior := restrictionsRule(types.BoolNull(), essMap(t, map[string]string{
			"downloadV1": "disabled", // prior non-default
		}))
		restoreOmittedEnhancedSecurity(ctx, prior, m, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		got := m["downloadV1"].(types.String)
		if got.ValueString() != "enabled" {
			t.Errorf("downloadV1 = %q, want literal default \"enabled\" (drift must surface)", got.ValueString())
		}
	})

	t.Run("unknown-default key is left absent so out-of-band drift surfaces", func(t *testing.T) {
		// proxyTrafficV1 has no known default; the API only omits it when the
		// value was cleared out of band. Restoring the prior value would mask
		// that drift, so the key must be left absent instead.
		var diags diag.Diagnostics
		m := map[string]attr.Value{}
		prior := restrictionsRule(types.BoolNull(), essMap(t, map[string]string{
			"proxyTrafficV1": "secureBrowse", // no known default
		}))
		restoreOmittedEnhancedSecurity(ctx, prior, m, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if _, ok := m["proxyTrafficV1"]; ok {
			t.Errorf("proxyTrafficV1 = %v, want it left absent so drift is not masked", m["proxyTrafficV1"])
		}
	})

	t.Run("does not overwrite keys already present", func(t *testing.T) {
		var diags diag.Diagnostics
		m := map[string]attr.Value{
			"downloadV1": types.StringValue("disabled"), // API returned this non-default
		}
		prior := restrictionsRule(types.BoolNull(), essMap(t, map[string]string{
			"downloadV1": "enabled",
		}))
		restoreOmittedEnhancedSecurity(ctx, prior, m, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		got := m["downloadV1"].(types.String)
		if got.ValueString() != "disabled" {
			t.Errorf("downloadV1 = %q, want API value \"disabled\" preserved", got.ValueString())
		}
	})
}

// TestRestoreOmittedEnhancedSecurity_InsecureContent locks in that the backend
// key insecure_content_allowed_for_urls_v1 is restored to its "disabled" default
// when the API drops it (the value is elided at that value), preventing the
// "inconsistent result after apply" vanish error. At "enabled" the API echoes
// it, so no restoration is needed.
func TestRestoreOmittedEnhancedSecurity_InsecureContent(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics
	m := map[string]attr.Value{}
	prior := restrictionsRule(types.BoolNull(), essMap(t, map[string]string{
		"insecure_content_allowed_for_urls_v1": "disabled",
	}))
	restoreOmittedEnhancedSecurity(ctx, prior, m, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	got, ok := m["insecure_content_allowed_for_urls_v1"].(types.String)
	if !ok {
		t.Fatalf("insecure_content_allowed_for_urls_v1 not restored: %v", m)
	}
	if got.ValueString() != "disabled" {
		t.Errorf("insecure_content_allowed_for_urls_v1 = %q, want default \"disabled\"", got.ValueString())
	}
}

// TestMatchPriorAccessRule_PositionalIDConflict locks in the positional-fallback
// guard: when the API rule carries a non-empty ID that has no prior match and
// the positional prior rule carries a different known ID, the two are treated as
// distinct rules (e.g. one replaced out of band) and not paired — so stale
// conditions/restrictions are not cross-applied. When the prior ID is unknown
// (Create/import) or the API rule has no ID, positional pairing still applies.
func TestMatchPriorAccessRule_PositionalIDConflict(t *testing.T) {
	t.Run("conflicting known IDs are not positionally paired", func(t *testing.T) {
		prior := []AccessRuleResourceModel{
			{ID: types.StringValue("prior-0")},
		}
		got := matchPriorAccessRule(prior, AccessRule{ID: "api-different"}, 0, map[int]bool{})
		if got != nil {
			t.Errorf("matchPriorAccessRule() = %+v, want nil (IDs positively differ)", got)
		}
	})

	t.Run("unknown prior ID still allows positional pairing (Create/import)", func(t *testing.T) {
		prior := []AccessRuleResourceModel{
			{ID: types.StringUnknown()},
		}
		got := matchPriorAccessRule(prior, AccessRule{ID: "api-new"}, 0, map[int]bool{})
		if got != &prior[0] {
			t.Errorf("matchPriorAccessRule() = %+v, want prior[0] (unknown prior ID, guard inert)", got)
		}
	})

	t.Run("empty API rule ID still allows positional pairing", func(t *testing.T) {
		prior := []AccessRuleResourceModel{
			{ID: types.StringValue("prior-0")},
		}
		got := matchPriorAccessRule(prior, AccessRule{}, 0, map[int]bool{})
		if got != &prior[0] {
			t.Errorf("matchPriorAccessRule() = %+v, want prior[0] (no API ID, guard inert)", got)
		}
	})
}

// TestValidateEnhancedSecuritySettingsConfig covers the conservative plan-time
// value check: pinned toggle keys reject values outside {enabled, disabled},
// unknown/unfamiliar keys are left unvalidated, and unknown/null leaves defer.
func TestValidateEnhancedSecuritySettingsConfig(t *testing.T) {
	rule := func(ess types.Map) AccessPolicyResourceModel {
		return AccessPolicyResourceModel{
			AccessRules: []AccessRuleResourceModel{
				{Restrictions: &RestrictionsResourceModel{EnhancedSecuritySettings: ess}},
			},
		}
	}

	t.Run("valid toggle value passes", func(t *testing.T) {
		d := validateEnhancedSecuritySettingsConfig(rule(essMap(t, map[string]string{
			"clipboardV1":                          "disabled",
			"insecure_content_allowed_for_urls_v1": "enabled",
		})))
		if d.HasError() {
			t.Errorf("unexpected diags: %v", d)
		}
	})

	t.Run("invalid toggle value is rejected", func(t *testing.T) {
		d := validateEnhancedSecuritySettingsConfig(rule(essMap(t, map[string]string{
			"clipboardV1": "on",
		})))
		if !d.HasError() {
			t.Errorf("expected an error for clipboardV1=on")
		}
	})

	t.Run("unpinned key is left unvalidated", func(t *testing.T) {
		// proxyTrafficV1 and _browserV1 have no pinned domain, so any value passes.
		d := validateEnhancedSecuritySettingsConfig(rule(essMap(t, map[string]string{
			"proxyTrafficV1": "secureBrowse",
			"_browserV1":     "embeddedBrowser",
		})))
		if d.HasError() {
			t.Errorf("unexpected diags for unpinned keys: %v", d)
		}
	})

	t.Run("nil restrictions and null map are no-ops", func(t *testing.T) {
		d := validateEnhancedSecuritySettingsConfig(AccessPolicyResourceModel{
			AccessRules: []AccessRuleResourceModel{
				{Restrictions: nil},
				{Restrictions: &RestrictionsResourceModel{EnhancedSecuritySettings: types.MapNull(types.StringType)}},
			},
		})
		if d.HasError() {
			t.Errorf("unexpected diags: %v", d)
		}
	})
}

// stripKeys returns the sorted key set of an attr.Value map for assertions.
func strippedKeys(m map[string]attr.Value) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

func TestStripServerInjectedEnhancedSecurity(t *testing.T) {
	t.Run("drops _browserV1 the API injected when config never set it", func(t *testing.T) {
		prior := restrictionsRule(types.BoolValue(false), essMap(t, map[string]string{
			"downloadV1": "disabled",
		}))
		got := map[string]attr.Value{
			"downloadV1": types.StringValue("disabled"),
			"_browserV1": types.StringValue("embeddedBrowser"),
		}
		stripServerInjectedEnhancedSecurity(prior, got)
		if keys := strippedKeys(got); keys["_browserV1"] || !keys["downloadV1"] {
			t.Fatalf("expected only downloadV1 to survive, got %v", keys)
		}
	})

	t.Run("keeps _browserV1 when the prior config set it explicitly", func(t *testing.T) {
		prior := restrictionsRule(types.BoolValue(false), essMap(t, map[string]string{
			"downloadV1": "disabled",
			"_browserV1": "embeddedBrowser",
		}))
		got := map[string]attr.Value{
			"downloadV1": types.StringValue("disabled"),
			"_browserV1": types.StringValue("embeddedBrowser"),
		}
		stripServerInjectedEnhancedSecurity(prior, got)
		if keys := strippedKeys(got); !keys["_browserV1"] || !keys["downloadV1"] {
			t.Fatalf("expected both keys to survive, got %v", keys)
		}
	})

	t.Run("nil prior keeps injected key (import has no prior state)", func(t *testing.T) {
		got := map[string]attr.Value{
			"clipboardV1": types.StringValue("disabled"),
			"_browserV1":  types.StringValue("embeddedBrowser"),
		}
		stripServerInjectedEnhancedSecurity(nil, got)
		if keys := strippedKeys(got); !keys["_browserV1"] || !keys["clipboardV1"] {
			t.Fatalf("expected _browserV1 kept with nil prior, got %v", keys)
		}
	})

	t.Run("empty map is a no-op", func(t *testing.T) {
		got := map[string]attr.Value{}
		stripServerInjectedEnhancedSecurity(nil, got)
		if len(got) != 0 {
			t.Fatalf("expected empty map to stay empty, got %v", got)
		}
	})
}
