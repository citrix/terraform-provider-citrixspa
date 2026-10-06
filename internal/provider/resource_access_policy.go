package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccessPolicyResource{}
var _ resource.ResourceWithImportState = &AccessPolicyResource{}
var _ resource.ResourceWithValidateConfig = &AccessPolicyResource{}
var _ resource.ResourceWithModifyPlan = &AccessPolicyResource{}

func NewAccessPolicyResource() resource.Resource {
	return &AccessPolicyResource{}
}

// AccessPolicyResource defines the resource implementation.
type AccessPolicyResource struct {
	client SPAClient
}

// AccessPolicyResourceModel describes the resource data model.
type AccessPolicyResourceModel struct {
	ID          types.String              `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	Description types.String              `tfsdk:"description"`
	Active      types.Bool                `tfsdk:"active"`
	Priority    types.Int64               `tfsdk:"priority"`
	Apps        types.Set                 `tfsdk:"apps"`
	AccessRules []AccessRuleResourceModel `tfsdk:"access_rules"`
}

type AccessRuleResourceModel struct {
	ID               types.String                   `tfsdk:"id"`
	Name             types.String                   `tfsdk:"name"`
	Description      types.String                   `tfsdk:"description"`
	Priority         types.Int64                    `tfsdk:"priority"`
	Active           types.Bool                     `tfsdk:"active"`
	Access           types.String                   `tfsdk:"access"`
	AccessNative     types.String                   `tfsdk:"access_native"`
	AdvancedSettings *AdvancedSettingsResourceModel `tfsdk:"advanced_settings"`
	Conditions       []ConditionResourceModel       `tfsdk:"conditions"`
	Restrictions     *RestrictionsResourceModel     `tfsdk:"restrictions"`
	Rules            []RuleResourceModel            `tfsdk:"rules"`
}

type AdvancedSettingsResourceModel struct {
	DomainOverrides []DomainOverrideResourceModel `tfsdk:"domain_overrides"`
}

type DomainOverrideResourceModel struct {
	FQDN        types.String `tfsdk:"fqdn"`
	LocationIDs types.List   `tfsdk:"location_ids"`
	Type        types.String `tfsdk:"type"`
}

type ConditionResourceModel struct {
	PlatformFilter types.String `tfsdk:"platform_filter"`
	UserAndGroups  types.Map    `tfsdk:"user_and_groups"`
}

type RestrictionsResourceModel struct {
	RedirectSBS              types.Bool `tfsdk:"redirect_sbs"`
	EnhancedSecuritySettings types.Map  `tfsdk:"enhanced_security_settings"`
}

type RuleResourceModel struct {
	Type      types.String `tfsdk:"type"`
	Operator  types.String `tfsdk:"operator"`
	TagSource types.String `tfsdk:"tag_source"`
	TagKey    types.String `tfsdk:"tag_key"`
	Values    types.List   `tfsdk:"values"`
	Metadata  types.Map    `tfsdk:"metadata"`
}

func (r *AccessPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Metadata - Setting resource metadata")
	resp.TypeName = req.ProviderTypeName + "_access_policy"
}

func (r *AccessPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Schema - Defining resource schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SPA access policy.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Access policy identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the access policy",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the access policy",
				Optional:            true,
				Computed:            true,
				// Default "" (not UseStateForUnknown) so an omitted description clears
				// any prior value and stays equivalent to the API's empty string.
				Default: stringdefault.StaticString(""),
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the access policy is active",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority of the access policy. The Console owns/normalizes this value on save, so it is Optional + Computed to avoid spurious drift.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"apps": schema.SetAttribute{
				MarkdownDescription: "Set of application IDs associated with the access policy",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"access_rules": schema.ListNestedAttribute{
				MarkdownDescription: "Access rules for the access policy",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Access rule ID",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Access rule name. Optional; the SPA service does not require a name for an access rule. When omitted it is preserved from prior state by matching the rule's content, so an unrelated policy update does not clear a name the Console/API previously assigned.",
							Optional:            true,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Access rule description",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "Access rule priority. The Console owns/normalizes this value on save, so it is Optional + Computed. When omitted it is preserved from prior state by matching the rule's content; for a newly added rule it is assigned the next integer above the highest known priority in the list (so all-omitted rules become 1, 2, 3\u2026 while a rule added next to an explicit priority 5 becomes 6).",
							Optional:            true,
							Computed:            true,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether the access rule is active",
							Required:            true,
						},
						"access": schema.StringAttribute{
							MarkdownDescription: "Access type (ACCESS_DENY, ACCESS_ALLOW)",
							Required:            true,
						},
						"access_native": schema.StringAttribute{
							MarkdownDescription: "Native access type (ACCESS_DENY, ACCESS_ALLOW)",
							Optional:            true,
							Computed:            true,
						},
						"advanced_settings": schema.SingleNestedAttribute{
							MarkdownDescription: "Advanced settings for the access rule",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"domain_overrides": schema.ListNestedAttribute{
									MarkdownDescription: "Domain override settings",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"fqdn": schema.StringAttribute{
												MarkdownDescription: "Fully qualified domain name",
												Required:            true,
											},
											"location_ids": schema.ListAttribute{
												MarkdownDescription: "Location IDs",
												Required:            true,
												ElementType:         types.StringType,
											},
											"type": schema.StringAttribute{
												MarkdownDescription: "Domain override type",
												Required:            true,
											},
										},
									},
								},
							},
						},
						"conditions": schema.ListNestedAttribute{
							MarkdownDescription: "Conditions for the access rule",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"platform_filter": schema.StringAttribute{
										MarkdownDescription: "Platform filter (PLATFORM_FILTER_MOBILE, PLATFORM_FILTER_PC, PLATFORM_FILTER_ANY)",
										Optional:            true,
									},
									"user_and_groups": schema.MapAttribute{
										MarkdownDescription: "**Deprecated.** User and group scope as `identity token => display name` pairs. " +
											"The SPA service is retiring this field and the provider never sends it. The provider " +
											"translates any entries into an equivalent rule with `type = \"TYPE_USERGROUP\"` " +
											"so existing configurations keep working, and emits a plan-time warning showing " +
											"the rule to move into `rules[]`. This attribute will be removed in a future release.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
						"restrictions": schema.SingleNestedAttribute{
							MarkdownDescription: "Restrictions for the access rule",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"redirect_sbs": schema.BoolAttribute{
									MarkdownDescription: "Whether to redirect SBS. Defaults to `false` when omitted.",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"enhanced_security_settings": schema.MapAttribute{
									MarkdownDescription: "Enhanced security settings, keyed by setting name. The toggle keys `clipboardV1`, `downloadV1`, `printingV1`, `uploadV1`, `keyLoggingV1`, `screenCaptureV1`, `watermarkV1`, and `insecure_content_allowed_for_urls_v1` accept only `enabled` or `disabled`.",
									Optional:            true,
									ElementType:         types.StringType,
								},
							},
						},
						"rules": schema.ListNestedAttribute{
							MarkdownDescription: "Rules within the access rule",
							Required:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "Rule type. Valid values: `TYPE_TAG`, `TYPE_USERGROUP`, `TYPE_PLATFORM`, `TYPE_MACHINEGROUP`, `TYPE_MULTIURLDOMAIN`.",
										Required:            true,
									},
									"operator": schema.StringAttribute{
										MarkdownDescription: "Rule operator. Valid values: `OPERATOR_EQ`, `OPERATOR_IN`, etc. When `type` is `TYPE_MULTIURLDOMAIN`, only `OPERATOR_IN` or `OPERATOR_NOT` are accepted.",
										Required:            true,
									},
									"tag_source": schema.StringAttribute{
										MarkdownDescription: "Source of data retrieval for `TYPE_TAG` rules. May be omitted (or set to `\"\"`) when not applicable (including when `type` is `TYPE_MULTIURLDOMAIN`). Valid values: `\"\"`, `NLS`, `CAS`, `EPA`, `ITM`, `ThirdPartyDevicePosture`, `CONTEXTUAL`.",
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString(""),
									},
									"tag_key": schema.StringAttribute{
										MarkdownDescription: "Tag key for `TYPE_TAG` rules (e.g., `location-geo-country-isocode`). May be omitted (or set to `\"\"`) when not applicable (including when `type` is `TYPE_MULTIURLDOMAIN`).",
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString(""),
									},
									"values": schema.ListAttribute{
										MarkdownDescription: "Rule values.",
										Required:            true,
										ElementType:         types.StringType,
									},
									"metadata": schema.MapAttribute{
										MarkdownDescription: "Rule metadata as key-value pairs, namely usernames-SID/OID pairs (see example above), used for UIX purposes. Required for `TYPE_USERGROUP` rules: `values` must be resolvable directory tokens (`SID:/...`, `OID:/ad/...`, `OID:/azuread/...`) and `metadata` must map a display name to the comma-joined list of those tokens. Omitting it (or using non-resolvable placeholder values) makes the policy's edit page in the SPA Console fail to render. When omitted it is preserved from the prior-state rule matched by content (type/operator/tag/values) rather than by list position, so reordering rules keeps each rule's own metadata.",
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ValidateConfig enforces that TYPE_TAG rules carry non-empty tag_source and
// tag_key. Those attributes are Optional+Computed (Default "") so non-tag rules
// can omit them, which removed the schema-level Required guard; this restores a
// plan-time error for tag rules that leave them empty.
func (r *AccessPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// access_rules (and its nested rules) are collections that Config.Get decodes
	// into Go slices, which fails with a conversion error when any is unknown
	// (e.g. driven by an unresolved expression). Defer to apply in that case.
	var accessRules types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("access_rules"), &accessRules)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if accessRulesHasUnknown(accessRules) {
		return
	}

	var data AccessPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateTagRuleConfig(data)...)
	resp.Diagnostics.Append(validateEnhancedSecuritySettingsConfig(data)...)
	resp.Diagnostics.Append(validateDeprecatedUserAndGroups(ctx, data)...)
}

// resolveRulePriorities returns a collision-free API priority for each access
// rule. A rule with a known priority (explicitly configured, or adopted from its
// matched prior-state rule) keeps that exact value. A rule that omits it is
// assigned the next integer above the highest known priority, in list order, so a
// filled-in priority can never duplicate an explicitly-set one when a policy mixes
// the two. When every rule omits its priority this yields 1, 2, 3, … (the rule's
// 1-based list position), matching the historical default.
func resolveRulePriorities(known []*int) []int {
	maxKnown := 0
	for _, k := range known {
		if k != nil && *k > maxKnown {
			maxKnown = *k
		}
	}
	out := make([]int, len(known))
	next := maxKnown + 1
	for i, k := range known {
		if k != nil {
			out[i] = *k
			continue
		}
		out[i] = next
		next++
	}
	return out
}

// accessRulesHasUnknown reports whether the access_rules tree contains any
// unknown list, set, object, or map-element value at any depth. Config.Get
// decodes collection/block boundaries into Go slices, structs, and pointers
// (rules, conditions, advanced_settings, restrictions, domain_overrides) that
// cannot hold an unknown value, so a block wired to an unresolved expression
// would otherwise raise a conversion diagnostic during validation.
// ValidateConfig defers to apply when this returns true.
//
// Maps are deliberately not walked. A types.Map field holds an unknown map, and
// an unknown element inside a known one, without a conversion error, so neither
// shape blocks the whole-model Get. Only ElementsAs into map[string]string
// cannot represent an unknown element, and the one reader that does that —
// userGroupScopeFromConditions, for user_and_groups — skips such a condition
// itself. Treating maps as unknown here instead would defer every unrelated
// validation (a TYPE_TAG rule missing tag_source, say) on any resource that
// wires one sub-rule's metadata to an apply-time expression.
func accessRulesHasUnknown(v attr.Value) bool {
	switch t := v.(type) {
	case types.List:
		if t.IsUnknown() {
			return true
		}
		return anyElemHasUnknown(t.Elements())
	case types.Set:
		if t.IsUnknown() {
			return true
		}
		return anyElemHasUnknown(t.Elements())
	case types.Object:
		if t.IsUnknown() {
			return true
		}
		for _, attrVal := range t.Attributes() {
			if accessRulesHasUnknown(attrVal) {
				return true
			}
		}
	}
	return false
}

func anyElemHasUnknown(elems []attr.Value) bool {
	for _, e := range elems {
		if accessRulesHasUnknown(e) {
			return true
		}
	}
	return false
}

// enhancedSecurityValueDomains pins the accepted values for the enhanced-
// security keys whose value domain has been empirically confirmed against the
// live backend (all enabled/disabled toggles). Keys NOT listed here — such as
// _browserV1, proxyTrafficV1, and any future setting — are intentionally left
// unconstrained so legitimate or unfamiliar values are never rejected.
var enhancedSecurityValueDomains = map[string][]string{
	"clipboardV1":                          {"enabled", "disabled"},
	"downloadV1":                           {"enabled", "disabled"},
	"printingV1":                           {"enabled", "disabled"},
	"uploadV1":                             {"enabled", "disabled"},
	"keyLoggingV1":                         {"enabled", "disabled"},
	"screenCaptureV1":                      {"enabled", "disabled"},
	"watermarkV1":                          {"enabled", "disabled"},
	"insecure_content_allowed_for_urls_v1": {"enabled", "disabled"},
}

// validateEnhancedSecuritySettingsConfig returns diagnostics for any
// enhanced_security_settings entry whose key has a known value domain but whose
// known, non-empty value falls outside it. Unknown values defer to apply, and
// keys with no pinned domain are left unvalidated so unfamiliar or future
// settings are never rejected. It is a pure function of the config model so it
// can be unit tested without constructing a full tfsdk.Config.
func validateEnhancedSecuritySettingsConfig(data AccessPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	for i, ar := range data.AccessRules {
		if ar.Restrictions == nil {
			continue
		}
		ess := ar.Restrictions.EnhancedSecuritySettings
		if ess.IsNull() || ess.IsUnknown() {
			continue
		}
		essPath := path.Root("access_rules").AtListIndex(i).AtName("restrictions").AtName("enhanced_security_settings")
		for key, elem := range ess.Elements() {
			allowed, pinned := enhancedSecurityValueDomains[key]
			if !pinned {
				continue
			}
			sv, ok := elem.(types.String)
			if !ok || sv.IsUnknown() || sv.IsNull() {
				continue
			}
			val := sv.ValueString()
			valid := false
			for _, a := range allowed {
				if val == a {
					valid = true
					break
				}
			}
			if !valid {
				diags.AddAttributeError(
					essPath.AtMapKey(key),
					"Invalid enhanced_security_settings value",
					fmt.Sprintf("enhanced_security_settings[%q] = %q is not valid. Allowed values: %s.",
						key, val, strings.Join(allowed, ", ")),
				)
			}
		}
	}
	return diags
}

// validateTagRuleConfig returns diagnostics for any TYPE_TAG rule missing a
// non-empty tag_source or tag_key. It is a pure function of the config model so
// it can be unit tested without constructing a full tfsdk.Config.
func validateTagRuleConfig(data AccessPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	for i, ar := range data.AccessRules {
		for j, rule := range ar.Rules {
			if rule.Type.ValueString() != "TYPE_TAG" {
				continue
			}
			rulePath := path.Root("access_rules").AtListIndex(i).AtName("rules").AtListIndex(j)
			if isEmptyConfigString(rule.TagSource) {
				diags.AddAttributeError(
					rulePath.AtName("tag_source"),
					"Missing tag_source for TYPE_TAG rule",
					"tag_source must be set to a non-empty value when a rule's type is TYPE_TAG.",
				)
			}
			if isEmptyConfigString(rule.TagKey) {
				diags.AddAttributeError(
					rulePath.AtName("tag_key"),
					"Missing tag_key for TYPE_TAG rule",
					"tag_key must be set to a non-empty value when a rule's type is TYPE_TAG.",
				)
			}
		}
	}
	return diags
}

// isEmptyConfigString reports whether a config string is known and empty (null
// or ""). Unknown values are treated as non-empty so validation defers to apply.
func isEmptyConfigString(s types.String) bool {
	if s.IsUnknown() {
		return false
	}
	return s.IsNull() || s.ValueString() == ""
}

// ModifyPlan reconciles backend-owned computed values on the access_rules list
// during an update: each rule's backend-assigned id and each sub-rule's Console
// metadata map. The backend honors whatever id/metadata it is sent, so filling an
// omitted value from the rule at the same list index reassigns rule identities and
// mis-associates metadata whenever the rules are reordered or a rule is inserted
// (verified against the live service). Instead, each plan rule is matched to its
// prior-state counterpart by content (with a positional fallback for an in-place
// edit); the matched rule's id and sub-rule metadata are adopted so identities and
// metadata follow the rule across a reorder. A rule with no prior match is new: its
// id is marked unknown (backend assigns one) and its omitted sub-rule metadata is
// marked unknown (backend materializes {}), which also avoids the "Provider produced
// inconsistent result after apply" error for genuinely new rules. An explicitly
// configured id or metadata value is always honored.
func (r *AccessPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only relevant when updating an existing resource (both state and plan present).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// access_rules (and its nested blocks) decode into Go slices/structs that
	// cannot hold an unknown value, so a collection wired to an unresolved
	// expression would make the whole-model Get below raise a conversion
	// diagnostic and block the plan. Defer reconciliation to apply in that case;
	// the id/metadata massaging here is only meaningful once the rules are known.
	var planRules, configRules types.List
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("access_rules"), &planRules)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("access_rules"), &configRules)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if accessRulesHasUnknown(planRules) || accessRulesHasUnknown(configRules) {
		// Full reconciliation is deferred to apply, but the nested computed `id`
		// must not keep the positional value UseStateForUnknown copied from the
		// old list index: on a reorder that locks a swapped identity into the plan
		// and, via Update, into the backend. For every rule whose id is not
		// explicitly configured, force the planned id unknown so apply assigns the
		// identity-correct id (matched by content or unique name in Update) without
		// a "provider produced inconsistent result after apply" error.
		if !planRules.IsNull() && !planRules.IsUnknown() {
			for i := range planRules.Elements() {
				idPath := path.Root("access_rules").AtListIndex(i).AtName("id")
				var cfgID types.String
				explicit := false
				if d := req.Config.GetAttribute(ctx, idPath, &cfgID); !d.HasError() {
					explicit = !cfgID.IsNull() && !cfgID.IsUnknown()
				}
				if !explicit {
					resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, idPath, types.StringUnknown())...)
				}
			}
		}
		return
	}

	var plan, state, config AccessPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Match each plan access rule to its prior-state counterpart so backend-owned
	// computed values (the rule id and each sub-rule's metadata) can be reconciled
	// by identity rather than by list position. The backend honors whatever rule id
	// it is sent, so filling an omitted id/metadata from the rule at the same index
	// reassigns identities and mis-associates metadata whenever the rules are
	// reordered or a rule is inserted. Matching is content-first (reorder-stable)
	// with a positional fallback for an in-place edit; a plan rule with no match is
	// genuinely new. An explicitly-configured rule id is authoritative and matched
	// first, so it is never overridden by a positional guess on a reorder+edit.
	trustedIDs := make([]string, len(plan.AccessRules))
	for i := range plan.AccessRules {
		if i < len(config.AccessRules) {
			if id := config.AccessRules[i].ID; !id.IsNull() && !id.IsUnknown() {
				trustedIDs[i] = id.ValueString()
			}
		}
	}
	priorMatch := matchAccessRulesToPrior(ctx, plan.AccessRules, state.AccessRules, trustedIDs)

	// Flag rules whose identity cannot be settled until apply: metadata wired to an
	// apply-time value is unknown now, so the metadata-inclusive match (Pass 0b) that
	// separates two otherwise-identical rules is skipped at plan time and only runs at
	// apply. When such a rule shares its metadata-independent content+name key with
	// another candidate identity, the plan-time (positional/FIFO) id and the apply-time
	// (metadata-matched) id can differ, yielding "provider produced inconsistent final
	// plan". That other candidate may be a sibling in the same plan (a reorder) or a
	// prior-state rule that is being deleted while this one survives: both cases leave
	// more than one prior id the unknown metadata could legitimately resolve to.
	// Deferring the id to unknown for the ambiguous rule lets apply assign the
	// metadata-correct id without that mismatch.
	ruleKey := func(r AccessRuleResourceModel) string {
		nm := ""
		if !r.Name.IsNull() && !r.Name.IsUnknown() {
			nm = r.Name.ValueString()
		}
		return ruleContentSignature(ctx, r, false) + "\x1fname=" + nm
	}
	keyCount := make(map[string]int)
	for i := range plan.AccessRules {
		keyCount[ruleKey(plan.AccessRules[i])]++
	}
	// Count prior-state rules sharing each metadata-independent key too, so a surviving
	// rule with unknown metadata is still flagged when its identical-content sibling is
	// deleted from the plan (plan-only counting would see just one and miss it).
	priorKeyCount := make(map[string]int)
	for i := range state.AccessRules {
		priorKeyCount[ruleKey(state.AccessRules[i])]++
	}
	metaHasUnknown := func(m types.Map) bool {
		if m.IsUnknown() {
			return true
		}
		if m.IsNull() {
			return false
		}
		for _, e := range m.Elements() {
			if e.IsUnknown() {
				return true
			}
		}
		return false
	}
	metaUnknownAmbiguous := make([]bool, len(plan.AccessRules))
	for i := range plan.AccessRules {
		k := ruleKey(plan.AccessRules[i])
		if keyCount[k] < 2 && priorKeyCount[k] < 2 {
			continue
		}
		for j := range plan.AccessRules[i].Rules {
			if metaHasUnknown(plan.AccessRules[i].Rules[j].Metadata) {
				metaUnknownAmbiguous[i] = true
				break
			}
		}
	}

	for i := range plan.AccessRules {
		var cfgRule *AccessRuleResourceModel
		if i < len(config.AccessRules) {
			cfgRule = &config.AccessRules[i]
		}
		prior := priorMatch[i] // matched prior-state rule, or nil for a new rule

		// Reconcile the backend-assigned id. An explicitly configured, known id is
		// authoritative and left as planned. A null (omitted) id adopts the matched
		// prior rule's id (identity retention across a reorder/insert), or is marked
		// unknown for a genuinely new rule. A configured-but-unknown id (wired to an
		// apply-time value) is a requested change, not an omission: keep it unknown so
		// the requested value materializes at apply instead of the positional id that
		// UseStateForUnknown copied from the old list index.
		cfgIDNull := cfgRule == nil || cfgRule.ID.IsNull()
		cfgIDUnknown := cfgRule != nil && cfgRule.ID.IsUnknown()
		if cfgIDNull || cfgIDUnknown || metaUnknownAmbiguous[i] {
			idPath := path.Root("access_rules").AtListIndex(i).AtName("id")
			if prior != nil && cfgIDNull && !metaUnknownAmbiguous[i] {
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, idPath, prior.ID)...)
			} else {
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, idPath, types.StringUnknown())...)
			}
		}

		// Reconcile each sub-rule's metadata within the matched prior rule only, so a
		// sub-rule can never adopt metadata from a different access rule that happens
		// to share the same values. An explicit metadata map is kept (but still
		// consumes its matching prior slot to keep the per-parent FIFO aligned); an
		// omitted map adopts the content-matched prior metadata, or is marked unknown
		// (backend materializes {}) when the sub-rule is new.
		var localMeta map[string][]types.Map
		if prior != nil {
			localMeta = make(map[string][]types.Map)
			for j := range prior.Rules {
				key := ruleMetadataMatchKey(ctx, prior.Rules[j])
				localMeta[key] = append(localMeta[key], prior.Rules[j].Metadata)
			}
		}
		for j := range plan.AccessRules[i].Rules {
			key := ruleMetadataMatchKey(ctx, plan.AccessRules[i].Rules[j])
			planMeta := plan.AccessRules[i].Rules[j].Metadata
			var matched *types.Map
			if localMeta != nil {
				matched, localMeta[key] = consumePriorMeta(localMeta[key], planMeta)
			}
			if !planMeta.IsNull() && !planMeta.IsUnknown() {
				continue // explicit known value kept; its equal prior slot consumed above
			}
			// A metadata map that is explicitly set in configuration but still
			// unknown at plan time (wired to an apply-time expression) is a
			// requested change, not an omission: keep it unknown so the new value
			// materializes at apply. Only a null (omitted) map adopts prior state.
			// A wholly-unknown map does not force ValidateConfig to defer, so it
			// reaches this known path and must be distinguished here.
			if cfgRule != nil && j < len(cfgRule.Rules) && !cfgRule.Rules[j].Metadata.IsNull() {
				metaPath := path.Root("access_rules").AtListIndex(i).AtName("rules").AtListIndex(j).AtName("metadata")
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, metaPath, types.MapUnknown(types.StringType))...)
				continue
			}
			metaPath := path.Root("access_rules").AtListIndex(i).AtName("rules").AtListIndex(j).AtName("metadata")
			if matched != nil {
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, metaPath, *matched)...)
			} else {
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, metaPath, types.MapUnknown(types.StringType))...)
			}
		}
	}
}

// matchAccessRulesToPrior pairs each plan access rule with a prior-state access
// rule so backend-owned computed values (rule id, sub-rule metadata) can be
// reconciled by identity across a reorder or insert. It returns a slice parallel
// to planRules; entry i is the matched prior rule or nil when the plan rule is new.
//
// trustedIDs, when non-nil, is parallel to planRules and carries the
// explicitly-configured backend id for each plan rule ("" when the id was
// omitted). Callers pass config-authoritative ids only: ModifyPlan derives them
// from configuration (never the positional value UseStateForUnknown fills in),
// and Update from the already-reconciled plan. Pass nil to disable id matching.
//
// Matching runs in five passes so identity survives the common edits:
//
//   - Pass 0 matches on an explicitly-configured id and consumes that prior first,
//     so an authoritative backend identity is never overridden by a later content
//     or positional guess. An id-anchored plan rule whose id matches no prior is a
//     genuinely new rule with a chosen id and is excluded from the remaining
//     passes, so it cannot steal another rule's prior by content or position.
//   - Pass 0b matches on a metadata-inclusive content signature plus name, but only
//     for rules whose metadata is fully known on both sides. It distinguishes two
//     otherwise-identical rules that carry different explicit metadata, so a reorder
//     does not swap their ids; rules with omitted/unknown metadata fall through to
//     the metadata-independent passes below.
//   - Pass 1 matches on the metadata-independent content signature PLUS the rule
//     name, but only for plan rules whose name and signature fields are all known
//     (set in config). The name is reorder-stable parent context: it distinguishes
//     two rules that target the same values but carry different metadata, so
//     reordering them does not swap their identities/metadata.
//   - Pass 2 matches remaining rules on the content signature alone (FIFO), so a
//     reordered rule that omits its name still keeps its identity. A rule with any
//     unknown signature field is skipped here (its signature would match a prior
//     rule whose corresponding field was merely null); it is recovered by unique
//     name in Pass 3 or left unmatched.
//   - Pass 3 recovers every still-unmatched rule with a unique known name, so a
//     rule that is reordered and content-edited together (or whose signature
//     fields are unknown) keeps its identity by name instead of falling through
//     to the positional fallback and swapping ids.
//   - Pass 3b matches a still-unmatched named rule to a leftover prior by content
//     signature alone, recovering a rule that was renamed AND reordered in the same
//     plan (both name passes miss) before the positional fallback can swap its id.
//     It runs after the name passes so a surviving name always wins, and after
//     Pass 2 so it only takes priors an omitted-name counterpart did not need.
//   - Pass 4 falls back to a positional match between still-unmatched plan and prior
//     rules, so an in-place content edit keeps the id at that position rather than
//     churning it. Rules with any unknown signature field are excluded here: a
//     positional guess could pair them with the wrong prior rule.
//
// A plan rule that matches in no pass (an inserted or appended rule, or an
// unnamed rule with unknown signature fields) stays nil.
func matchAccessRulesToPrior(ctx context.Context, planRules, stateRules []AccessRuleResourceModel, trustedIDs []string) []*AccessRuleResourceModel {
	out := make([]*AccessRuleResourceModel, len(planRules))
	priorUsed := make([]bool, len(stateRules))
	idAnchored := make([]bool, len(planRules))

	nameKnown := func(s types.String) (string, bool) {
		if s.IsNull() || s.IsUnknown() {
			return "", false
		}
		return s.ValueString(), true
	}

	// Pass 0: explicit-id match. An explicitly-configured id is authoritative, so
	// consume its prior before any content/name/positional pass can pair the rule
	// with a sibling at its new list index (which would keep the id but adopt the
	// wrong rule's omitted name, priority, and metadata).
	if trustedIDs != nil {
		byID := make(map[string][]int)
		for j := range stateRules {
			if id := stateRules[j].ID.ValueString(); id != "" {
				byID[id] = append(byID[id], j)
			}
		}
		for i := range planRules {
			if i >= len(trustedIDs) || trustedIDs[i] == "" {
				continue
			}
			idAnchored[i] = true
			q := byID[trustedIDs[i]]
			for len(q) > 0 && priorUsed[q[0]] {
				q = q[1:]
			}
			if len(q) > 0 {
				j := q[0]
				byID[trustedIDs[i]] = q[1:]
				priorUsed[j] = true
				out[i] = &stateRules[j]
			}
		}
	}

	// Pass 0b: metadata-inclusive content+name match. Two rules that are identical
	// except for distinct, explicitly-configured metadata share the
	// metadata-independent signature the later passes use and would FIFO-swap ids
	// on reorder. Disambiguate them first by a signature that includes metadata
	// plus the (possibly empty) name. Only rules whose metadata is fully known on
	// both sides participate; a rule with omitted/unknown metadata falls through to
	// the metadata-independent passes, where plan metadata is expected to be
	// unknown while it is reconciled.
	bySigMeta := make(map[string][]int)
	for j := range stateRules {
		if priorUsed[j] || !ruleMetadataAllKnown(stateRules[j]) {
			continue
		}
		pnm, _ := nameKnown(stateRules[j].Name)
		k := ruleContentSignature(ctx, stateRules[j], true) + "\x1fname=" + pnm
		bySigMeta[k] = append(bySigMeta[k], j)
	}
	for i := range planRules {
		if out[i] != nil || idAnchored[i] || ruleIdentityHasUnknown(planRules[i]) || !ruleMetadataAllKnown(planRules[i]) {
			continue
		}
		nm, _ := nameKnown(planRules[i].Name)
		k := ruleContentSignature(ctx, planRules[i], true) + "\x1fname=" + nm
		if q := bySigMeta[k]; len(q) > 0 {
			j := q[0]
			bySigMeta[k] = q[1:]
			priorUsed[j] = true
			out[i] = &stateRules[j]
		}
	}

	// Pass 1: content signature + name, for named plan rules.
	byContentName := make(map[string][]int)
	for j := range stateRules {
		// Skip priors already consumed by the explicit-id pass, else a differently
		// -positioned plan rule with the same name+content could re-match a used
		// prior and inherit its id, sending one backend identity twice.
		if priorUsed[j] {
			continue
		}
		if nm, ok := nameKnown(stateRules[j].Name); ok {
			k := ruleContentSignature(ctx, stateRules[j], false) + "\x1fname=" + nm
			byContentName[k] = append(byContentName[k], j)
		}
	}
	for i := range planRules {
		if out[i] != nil || idAnchored[i] || ruleIdentityHasUnknown(planRules[i]) {
			continue
		}
		nm, ok := nameKnown(planRules[i].Name)
		if !ok {
			continue
		}
		k := ruleContentSignature(ctx, planRules[i], false) + "\x1fname=" + nm
		if q := byContentName[k]; len(q) > 0 {
			j := q[0]
			byContentName[k] = q[1:]
			priorUsed[j] = true
			out[i] = &stateRules[j]
		}
	}

	// Pass 2: content signature alone, for still-unmatched rules.
	bySig := make(map[string][]int)
	for j := range stateRules {
		if priorUsed[j] {
			continue
		}
		sig := ruleContentSignature(ctx, stateRules[j], false)
		bySig[sig] = append(bySig[sig], j)
	}
	for i := range planRules {
		if out[i] != nil || idAnchored[i] || ruleIdentityHasUnknown(planRules[i]) {
			continue
		}
		// A plan rule with a known name is an identity signal: it matches a prior
		// only by that name (Pass 1/3), never by content alone here. Skipping it
		// stops a newly inserted, differently-named rule from stealing the content
		// match that an omitted-name existing counterpart needs.
		if _, named := nameKnown(planRules[i].Name); named {
			continue
		}
		sig := ruleContentSignature(ctx, planRules[i], false)
		if q := bySig[sig]; len(q) > 0 {
			j := q[0]
			bySig[sig] = q[1:]
			priorUsed[j] = true
			out[i] = &stateRules[j]
		}
	}

	// Pass 3: unique-name recovery for every still-unmatched named rule. A rule
	// that is reordered and content-edited in the same plan (or whose signature
	// fields are unknown) no longer matches prior state by content, so Pass 1/2
	// miss it; matching by a unique known name keeps its identity instead of
	// letting the positional fallback swap ids between two stably-named rules.
	for i := range planRules {
		if out[i] != nil || idAnchored[i] {
			continue
		}
		nm, ok := nameKnown(planRules[i].Name)
		if !ok {
			continue
		}
		cand, dup := -1, false
		for j := range stateRules {
			if priorUsed[j] {
				continue
			}
			if pnm, pok := nameKnown(stateRules[j].Name); pok && pnm == nm {
				if cand >= 0 {
					dup = true
					break
				}
				cand = j
			}
		}
		if cand >= 0 && !dup {
			priorUsed[cand] = true
			out[i] = &stateRules[cand]
		}
	}

	// Pass 3b: content signature for still-unmatched NAMED plan rules, against the
	// priors left after the name passes. A rule renamed AND reordered in the same
	// plan misses every name pass, and Pass 2 deliberately skips named rules; without
	// this it would fall through to the positional fallback and swap ids. It runs
	// after Pass 3 so a rule whose name still exists in state keeps its name match
	// rather than a content match to a different prior, and after Pass 2 so it only
	// claims leftover priors and cannot steal an omitted-name counterpart's match.
	bySigNamed := make(map[string][]int)
	for j := range stateRules {
		if priorUsed[j] {
			continue
		}
		sig := ruleContentSignature(ctx, stateRules[j], false)
		bySigNamed[sig] = append(bySigNamed[sig], j)
	}
	for i := range planRules {
		if out[i] != nil || idAnchored[i] || ruleIdentityHasUnknown(planRules[i]) {
			continue
		}
		if _, named := nameKnown(planRules[i].Name); !named {
			continue
		}
		sig := ruleContentSignature(ctx, planRules[i], false)
		if q := bySigNamed[sig]; len(q) > 0 {
			j := q[0]
			bySigNamed[sig] = q[1:]
			priorUsed[j] = true
			out[i] = &stateRules[j]
		}
	}

	// Pass 4: positional fallback for an in-place edit — but only for plan rules
	// whose signature fields are all known. A rule with any unknown signature
	// field is ambiguous; a positional match could pair it with the wrong prior
	// rule and swap backend ids, so leave it unmatched (its id is planned unknown
	// for the backend to resolve rather than reassigned to a sibling's id).
	for i := range planRules {
		if out[i] != nil || idAnchored[i] || ruleIdentityHasUnknown(planRules[i]) {
			continue
		}
		if i < len(stateRules) && !priorUsed[i] {
			priorUsed[i] = true
			out[i] = &stateRules[i]
		}
	}
	return out
}

// consumePriorMeta reconciles one sub-rule's metadata against the per-parent
// queue of prior metadata maps sharing the same match key. When the plan omitted
// metadata it adopts (and removes) the first queued prior value (FIFO). When the
// plan supplied metadata explicitly it keeps that value and removes the prior
// entry equal to it, or—when the explicit value changed and no prior is equal—
// removes the queue head (its positional slot), so an omitted same-key sibling
// still adopts the other prior value instead of this sibling's old one. It
// returns the value to adopt (nil when the explicit plan value is kept or nothing
// matched) and the updated queue.
func consumePriorMeta(q []types.Map, planMeta types.Map) (adopt *types.Map, rest []types.Map) {
	if !planMeta.IsNull() && !planMeta.IsUnknown() {
		for i := range q {
			if q[i].Equal(planMeta) {
				rest = append(append([]types.Map{}, q[:i]...), q[i+1:]...)
				return nil, rest
			}
		}
		// Explicit value with no equal prior = a changed sibling. Consume its
		// positional slot (the queue head) so a following omitted same-key sibling
		// adopts the next prior value rather than this sibling's old metadata.
		if len(q) > 0 {
			return nil, q[1:]
		}
		return nil, q
	}
	if len(q) > 0 {
		m := q[0]
		return &m, q[1:]
	}
	return nil, q
}

// ruleIdentityHasUnknown reports whether any field the content signature depends
// on is unknown (a scalar, or a list/map that is unknown or has an unknown
// element). It deliberately mirrors ruleContentSignature so a rule the signature
// cannot reliably compare is also barred from the positional fallback. Sub-rule
// metadata is excluded: it is reconciled by ModifyPlan and is expected to be
// unknown during matching. An unknown signature field decodes to an empty/partial
// value (the conversion diagnostic is intentionally discarded when building the
// signature), so its plan signature would not match the concrete prior-state
// signature; callers must not positionally reconcile such a rule.
func ruleIdentityHasUnknown(rule AccessRuleResourceModel) bool {
	listU := func(l types.List) bool {
		if l.IsUnknown() {
			return true
		}
		if l.IsNull() {
			return false
		}
		for _, e := range l.Elements() {
			if e.IsUnknown() {
				return true
			}
		}
		return false
	}
	mapU := func(m types.Map) bool {
		if m.IsUnknown() {
			return true
		}
		if m.IsNull() {
			return false
		}
		for _, e := range m.Elements() {
			if e.IsUnknown() {
				return true
			}
		}
		return false
	}

	if rule.Access.IsUnknown() || rule.Active.IsUnknown() || rule.Description.IsUnknown() {
		return true
	}
	for _, rr := range rule.Rules {
		if rr.Type.IsUnknown() || rr.Operator.IsUnknown() || rr.TagSource.IsUnknown() || rr.TagKey.IsUnknown() || listU(rr.Values) {
			return true
		}
	}
	for _, c := range rule.Conditions {
		if c.PlatformFilter.IsUnknown() || mapU(c.UserAndGroups) {
			return true
		}
	}
	if rule.AdvancedSettings != nil {
		for _, do := range rule.AdvancedSettings.DomainOverrides {
			if do.FQDN.IsUnknown() || do.Type.IsUnknown() || listU(do.LocationIDs) {
				return true
			}
		}
	}
	if rule.Restrictions != nil {
		if rule.Restrictions.RedirectSBS.IsUnknown() || mapU(rule.Restrictions.EnhancedSecuritySettings) {
			return true
		}
	}
	return false
}

// ruleMetadataMatchKey builds a metadata-independent key for a sub-rule from the
// fields that identify the group it targets (type, operator, tag_source, tag_key
// and the sorted values). It lets an omitted metadata map be reconciled against
// the matching prior-state sub-rule by content rather than by list position, so
// reordering access rules does not carry one rule's metadata onto another.
// ruleMetadataAllKnown reports whether the rule has at least one sub-rule and
// every sub-rule's metadata map is set and fully known, so a metadata-inclusive
// signature can be compared against prior state. A rule with no sub-rules, or any
// null/unknown metadata map, returns false and is matched by the
// metadata-independent passes instead.
func ruleMetadataAllKnown(rule AccessRuleResourceModel) bool {
	if len(rule.Rules) == 0 {
		return false
	}
	for _, rr := range rule.Rules {
		if rr.Metadata.IsNull() || rr.Metadata.IsUnknown() {
			return false
		}
	}
	return true
}

func ruleMetadataMatchKey(ctx context.Context, rr RuleResourceModel) string {
	const sentinel = "\x00"
	strSig := func(s types.String) string {
		if s.IsNull() || s.IsUnknown() {
			return sentinel
		}
		return s.ValueString()
	}
	valuesSig := sentinel
	if !rr.Values.IsNull() && !rr.Values.IsUnknown() {
		var vals []string
		rr.Values.ElementsAs(ctx, &vals, false)
		sort.Strings(vals)
		valuesSig = encodeStringSlice(vals)
	}
	return strings.Join([]string{
		strSig(rr.Type),
		strSig(rr.Operator),
		strSig(rr.TagSource),
		strSig(rr.TagKey),
		valuesSig,
	}, "\x1f")
}

// encodeStringSlice returns a collision-free encoding of a string slice by
// length-prefixing each element. A plain delimiter join is ambiguous because a
// value may itself contain the delimiter (["a,b"] and ["a","b"] would collide);
// length-prefixing makes every distinct slice map to a distinct string.
func encodeStringSlice(vals []string) string {
	var b strings.Builder
	for _, v := range vals {
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteByte(':')
		b.WriteString(v)
	}
	return b.String()
}

// encodeStringMap returns a collision-free encoding of a string map by
// length-prefixing each key and value over sorted keys, so map strings that
// contain a delimiter or `=` cannot make two different maps collide.
func encodeStringMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		v := m[k]
		b.WriteString(strconv.Itoa(len(k)))
		b.WriteByte(':')
		b.WriteString(k)
		b.WriteString(strconv.Itoa(len(v)))
		b.WriteByte(':')
		b.WriteString(v)
	}
	return b.String()
}

func (r *AccessPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(SPAClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected SPAClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *AccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Create - Creating access policy")
	var data AccessPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	policy := &AccessPolicy{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.Active.IsNull() {
		policy.Active = data.Active.ValueBool()
	} else {
		// Default to false if not specified (active is required by API)
		policy.Active = false
	}
	// Only send a policy priority when the user set an explicit, known value.
	// When omitted (null/unknown), leave it nil so the request omits the key and
	// the backend assigns the next available priority.
	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		p := int(data.Priority.ValueInt64())
		policy.Priority = &p
	}

	// Convert apps from Terraform Set to API []string
	if !data.Apps.IsNull() && !data.Apps.IsUnknown() {
		apps := make([]string, 0, len(data.Apps.Elements()))
		resp.Diagnostics.Append(data.Apps.ElementsAs(ctx, &apps, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		policy.Apps = apps
	}

	// Convert access rules from Terraform model to API model
	if len(data.AccessRules) > 0 {
		accessRules := make([]AccessRule, 0, len(data.AccessRules))
		// Access-rule priority is required by the API. Resolve every rule's priority
		// up front so an omitted one is placed above all explicit priorities instead
		// of the raw 1-based index, which could otherwise duplicate an explicit value
		// when a policy mixes explicit and omitted priorities.
		knownPrio := make([]*int, len(data.AccessRules))
		for i, rd := range data.AccessRules {
			if !rd.Priority.IsNull() && !rd.Priority.IsUnknown() {
				p := int(rd.Priority.ValueInt64())
				knownPrio[i] = &p
			}
		}
		resolvedPrio := resolveRulePriorities(knownPrio)
		for ruleIdx, ruleData := range data.AccessRules {
			rulePriority := resolvedPrio[ruleIdx]
			rule := AccessRule{
				Name:         ruleData.Name.ValueString(),
				Description:  ruleData.Description.ValueString(),
				Priority:     rulePriority,
				Active:       ruleData.Active.ValueBool(),
				Access:       ruleData.Access.ValueString(),
				AccessNative: ruleData.AccessNative.ValueString(),
				Conditions:   make([]Condition, 0), // Initialize as empty array
			}

			// Set ID if provided
			if !ruleData.ID.IsNull() && !ruleData.ID.IsUnknown() {
				rule.ID = ruleData.ID.ValueString()
			}

			// Convert AdvancedSettings
			if ruleData.AdvancedSettings != nil {
				advSettings := &AdvancedSettings{}
				if len(ruleData.AdvancedSettings.DomainOverrides) > 0 {
					domainOverrides := make([]DomainOverride, 0, len(ruleData.AdvancedSettings.DomainOverrides))
					for _, doData := range ruleData.AdvancedSettings.DomainOverrides {
						locationIDs := make([]string, 0)
						if !doData.LocationIDs.IsNull() && !doData.LocationIDs.IsUnknown() {
							resp.Diagnostics.Append(doData.LocationIDs.ElementsAs(ctx, &locationIDs, false)...)
							if resp.Diagnostics.HasError() {
								return
							}
						}
						domainOverrides = append(domainOverrides, DomainOverride{
							FQDN:        doData.FQDN.ValueString(),
							LocationIDs: locationIDs,
							Type:        doData.Type.ValueString(),
						})
					}
					advSettings.DomainOverrides = domainOverrides
				}
				rule.AdvancedSettings = advSettings
			}

			// Convert Conditions. The deprecated user_and_groups map is not sent;
			// it is translated into a TYPE_USERGROUP rule below.
			if len(ruleData.Conditions) > 0 {
				conditions := make([]Condition, 0, len(ruleData.Conditions))
				for _, condData := range ruleData.Conditions {
					conditions = append(conditions, Condition{
						PlatformFilter: condData.PlatformFilter.ValueString(),
					})
				}
				rule.Conditions = conditions
			}

			// Convert Restrictions
			if ruleData.Restrictions != nil {
				restrictions := &Restrictions{
					RedirectSBS: ruleData.Restrictions.RedirectSBS.ValueBool(),
				}
				if !ruleData.Restrictions.EnhancedSecuritySettings.IsNull() && !ruleData.Restrictions.EnhancedSecuritySettings.IsUnknown() {
					// Convert map[string]string to map[string]interface{}
					enhancedSettingsStr := make(map[string]string)
					resp.Diagnostics.Append(ruleData.Restrictions.EnhancedSecuritySettings.ElementsAs(ctx, &enhancedSettingsStr, false)...)
					if resp.Diagnostics.HasError() {
						return
					}
					enhancedSettings := make(map[string]interface{})
					for k, v := range enhancedSettingsStr {
						enhancedSettings[k] = v
					}
					restrictions.EnhancedSecuritySettings = enhancedSettings
				}
				rule.Restrictions = restrictions
			}

			// Convert Rules
			if len(ruleData.Rules) > 0 {
				rules := make([]Rule, 0, len(ruleData.Rules))
				for _, rData := range ruleData.Rules {
					r := Rule{
						Type:      rData.Type.ValueString(),
						Operator:  rData.Operator.ValueString(),
						TagSource: rData.TagSource.ValueString(),
						TagKey:    rData.TagKey.ValueString(),
					}
					if !rData.Values.IsNull() && !rData.Values.IsUnknown() {
						values := make([]string, 0)
						resp.Diagnostics.Append(rData.Values.ElementsAs(ctx, &values, false)...)
						if resp.Diagnostics.HasError() {
							return
						}
						r.Values = values
					}
					if !rData.Metadata.IsNull() && !rData.Metadata.IsUnknown() {
						// Convert map[string]string to map[string]interface{}
						metadataStr := make(map[string]string)
						resp.Diagnostics.Append(rData.Metadata.ElementsAs(ctx, &metadataStr, false)...)
						if resp.Diagnostics.HasError() {
							return
						}
						metadata := make(map[string]interface{})
						for k, v := range metadataStr {
							metadata[k] = v
						}
						r.Metadata = metadata
					}
					rules = append(rules, r)
				}
				rule.Rules = rules
			}

			// Translate the deprecated conditions[].user_and_groups maps into a
			// TYPE_USERGROUP rule, mirroring what the SPA Console does on save.
			ugScope, hasUG, ugDiags := userGroupScopeFromConditions(ctx, ruleData.Conditions)
			resp.Diagnostics.Append(ugDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if hasUG {
				rule.Rules = appendUserGroupRule(rule.Rules, ugScope)
			}

			accessRules = append(accessRules, rule)
		}
		policy.AccessRules = accessRules
	}

	// Create the policy
	tflog.Debug(ctx, "spa-terraform-provider: About to create access policy", map[string]any{
		"policy_name":        policy.Name,
		"apps_count":         len(policy.Apps),
		"access_rules_count": len(policy.AccessRules),
	})

	// Log each access rule for debugging
	for i, rule := range policy.AccessRules {
		tflog.Debug(ctx, "spa-terraform-provider: Access rule details", map[string]any{
			"rule_index": i,
			"rule_name":  rule.Name,
			"active":     rule.Active,
			"priority":   rule.Priority,
		})
	}

	createdPolicy, err := r.client.CreateAccessPolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create access policy, got error: %s", err))
		return
	}

	// Update the model with the created policy data
	data.ID = types.StringValue(createdPolicy.ID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the created policy to refresh computed values (like access_rules[].id and metadata)
	readReq := resource.ReadRequest{
		State: resp.State,
	}
	readResp := &resource.ReadResponse{
		State: resp.State,
	}

	r.Read(ctx, readReq, readResp)

	// Copy any diagnostics and the updated state
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

func (r *AccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Read - Reading access policy")
	var data AccessPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the policy from the API
	policy, err := r.client.GetAccessPolicy(ctx, data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read access policy, got error: %s", err))
		return
	}

	// Update the model with the API response
	data.Name = types.StringValue(policy.Name)
	data.Active = types.BoolValue(policy.Active)
	if policy.Priority != nil {
		data.Priority = types.Int64Value(int64(*policy.Priority))
	} else {
		data.Priority = types.Int64Value(0)
	}
	data.Description = types.StringValue(policy.Description)

	// Map apps to Terraform list
	appsValues := make([]attr.Value, 0)
	if policy.Apps != nil {
		for _, app := range policy.Apps {
			appsValues = append(appsValues, types.StringValue(app))
		}
	}
	appsList, appsDiags := types.SetValue(types.StringType, appsValues)
	resp.Diagnostics.Append(appsDiags...)
	data.Apps = appsList

	// Policy-level conditions and actions don't exist in the API - they are only at the access rule level

	// Snapshot the prior state's access rules so we can preserve the null vs
	// empty-list distinction for optional list fields (e.g. conditions) that the
	// API omits when empty. A nil Go slice → Terraform null; a non-nil empty
	// slice → Terraform empty list. The two are not interchangeable and the
	// framework will raise an inconsistency error if we return the wrong one.
	priorAccessRules := data.AccessRules

	// Convert access rules using the complex nested structure
	accessRules := make([]AccessRuleResourceModel, 0)
	// Tracks which prior rules have already been paired with an API rule, so a
	// prior rule is never matched twice (e.g. two rules sharing a name).
	usedPriorRules := make(map[int]bool)
	for ruleIdx, rule := range policy.AccessRules {
		// Handle optional ID field - use null if empty
		var ruleID types.String
		if rule.ID != "" {
			ruleID = types.StringValue(rule.ID)
		} else {
			ruleID = types.StringNull()
		}

		accessRule := AccessRuleResourceModel{
			ID:           ruleID,
			Name:         types.StringValue(rule.Name),
			Description:  types.StringValue(rule.Description),
			Priority:     types.Int64Value(int64(rule.Priority)),
			Active:       types.BoolValue(rule.Active),
			Access:       types.StringValue(rule.Access),
			AccessNative: types.StringValue(rule.AccessNative),
		}

		// Convert AdvancedSettings - only set if it has domain overrides
		if rule.AdvancedSettings != nil && len(rule.AdvancedSettings.DomainOverrides) > 0 {
			advancedSettings := &AdvancedSettingsResourceModel{}

			// Convert DomainOverrides
			domainOverrides := make([]DomainOverrideResourceModel, 0, len(rule.AdvancedSettings.DomainOverrides))
			for _, override := range rule.AdvancedSettings.DomainOverrides {
				locationIDsValues := make([]attr.Value, 0)
				for _, locationID := range override.LocationIDs {
					locationIDsValues = append(locationIDsValues, types.StringValue(locationID))
				}
				locationIDs, diags := types.ListValue(types.StringType, locationIDsValues)
				resp.Diagnostics.Append(diags...)

				domainOverrides = append(domainOverrides, DomainOverrideResourceModel{
					FQDN:        types.StringValue(override.FQDN),
					LocationIDs: locationIDs,
					Type:        types.StringValue(override.Type),
				})
			}
			advancedSettings.DomainOverrides = domainOverrides
			accessRule.AdvancedSettings = advancedSettings
		}

		// Resolve the prior-state rule that corresponds to this API rule by
		// identity (ID, then Name, then positional) so that conditions and
		// restrictions are both reconciled against the same prior rule even if
		// the backend returns the rules in a different order.
		priorRule := matchPriorAccessRule(priorAccessRules, rule, ruleIdx, usedPriorRules)

		// Resolve the scope the deprecated user_and_groups maps described on the
		// last write, and whether the service still carries it. Both the
		// conditions and the rules below depend on the answer.
		var ugScope userGroupScope
		var hasUG bool
		scopeEnforced := true
		declaredScopeRules := 0
		if priorRule != nil {
			var ugDiags diag.Diagnostics
			ugScope, hasUG, ugDiags = userGroupScopeFromConditions(ctx, priorRule.Conditions)
			resp.Diagnostics.Append(ugDiags...)
			if hasUG && !ugDiags.HasError() {
				declaredScopeRules, ugDiags = countDeclaredScopeMatches(ctx, priorRule.Rules, ugScope)
				resp.Diagnostics.Append(ugDiags...)
				scopeEnforced = countScopeMatches(rule.Rules, ugScope) > 0
			}
		}

		// Convert Conditions.
		//
		// The API omits conditions when there are none, so the slice will be
		// empty. Terraform distinguishes between null and an empty list:
		//   null  → attribute absent in config  (nil Go slice)
		//   []    → attribute present but empty (non-nil empty Go slice)
		// We must return whichever variant the prior state held; otherwise the
		// framework raises "was cty.ListValEmpty, but now null" (or vice-versa).
		if len(rule.Conditions) > 0 {
			// API returned real conditions — rebuild from API data.
			conditions := make([]ConditionResourceModel, 0, len(rule.Conditions))
			for condIdx, condition := range rule.Conditions {
				// user_and_groups is deprecated and is not sent, so it is
				// carried forward from the prior state verbatim — that keeps
				// null null, {} as {}, and a populated map populated. On import
				// there is no prior state and it stays null: the imported
				// configuration expresses the scope as a TYPE_USERGROUP rule.
				//
				// The exception is drift. If the translated rule was deleted
				// out of band the scope is no longer enforced, yet the
				// remaining rules still equal the configuration, so nothing
				// else in the response would reveal it. Clearing the map here
				// makes the next plan show user_and_groups coming back, which
				// re-synthesizes the rule on apply.
				userAndGroups := types.MapNull(types.StringType)
				if priorRule != nil && condIdx < len(priorRule.Conditions) && scopeEnforced {
					if prior := priorRule.Conditions[condIdx].UserAndGroups; !prior.IsUnknown() {
						userAndGroups = prior
					}
				}
				conditions = append(conditions, ConditionResourceModel{
					PlatformFilter: types.StringValue(condition.PlatformFilter),
					UserAndGroups:  userAndGroups,
				})
			}
			accessRule.Conditions = conditions
		} else {
			// API returned no conditions. Mirror the prior state so that null
			// stays null and [] stays [] — each maps to a distinct Terraform value.
			if priorRule != nil && priorRule.Conditions != nil {
				// Prior state had an explicit empty list — preserve it.
				accessRule.Conditions = []ConditionResourceModel{}
			}
			// else: prior was nil (null) — leave accessRule.Conditions nil (null).
		}

		// Convert Restrictions.
		//
		// The API omits any enhanced-security key whose value equals the backend
		// default, and omits the ENTIRE restrictions object when every setting is
		// at its default. Terraform then sees the configured block "vanish" and
		// raises an inconsistent-result error. It is reconciled against the prior
		// rule resolved above (matched by identity, not slice position) so that
		// restored settings stay attached to the correct rule even if the backend
		// reorders rules.
		if rule.Restrictions != nil {
			enhancedSecuritySettingsMap := make(map[string]attr.Value)
			if rule.Restrictions.EnhancedSecuritySettings != nil {
				for k, v := range rule.Restrictions.EnhancedSecuritySettings {
					enhancedSecuritySettingsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			stripServerInjectedEnhancedSecurity(priorRule, enhancedSecuritySettingsMap)
			restoreOmittedEnhancedSecurity(ctx, priorRule, enhancedSecuritySettingsMap, &resp.Diagnostics)

			var enhancedSecuritySettings types.Map
			switch {
			case len(enhancedSecuritySettingsMap) == 0 && priorRule != nil && priorRule.Restrictions != nil &&
				(priorRule.Restrictions.EnhancedSecuritySettings.IsNull() || priorRule.Restrictions.EnhancedSecuritySettings.IsUnknown()):
				// API returned the block but no enhanced-security keys (e.g. only
				// redirect_sbs was set); mirror the prior null/unknown map rather
				// than forcing an empty {} that would differ from the plan.
				enhancedSecuritySettings = priorRule.Restrictions.EnhancedSecuritySettings
			case len(enhancedSecuritySettingsMap) == 0 && rule.Restrictions.EnhancedSecuritySettings == nil &&
				(priorRule == nil || priorRule.Restrictions == nil):
				// API omitted the enhanced_security_settings field and there is no
				// prior map to preserve (e.g. a fresh import of a redirect_sbs-only
				// block, where priorRule is nil). Record null to match an omitted
				// optional map rather than an empty {}, which Terraform treats as
				// different and would show as a spurious post-import diff.
				enhancedSecuritySettings = types.MapNull(types.StringType)
			default:
				var diags diag.Diagnostics
				enhancedSecuritySettings, diags = types.MapValue(types.StringType, enhancedSecuritySettingsMap)
				resp.Diagnostics.Append(diags...)
			}

			accessRule.Restrictions = &RestrictionsResourceModel{
				RedirectSBS:              types.BoolValue(rule.Restrictions.RedirectSBS),
				EnhancedSecuritySettings: enhancedSecuritySettings,
			}
		} else if priorRestrictionsAllDefault(priorRule) {
			// The whole restrictions object was omitted because every setting is
			// at its default. Rebuild it from prior state so the block does not
			// vanish. Preserve the prior enhanced_security_settings map exactly —
			// including a null/unknown map, e.g. a restrictions block that only
			// set redirect_sbs — and, when it is a known map, restore the
			// default-valued keys the API dropped. redirect_sbs is preserved from
			// prior so an unset (null) value stays null.
			//
			// Reconstruction is gated on the prior block being semantically all
			// default: if a non-default block (e.g. redirect_sbs = true) is removed
			// out of band, the API also returns no restrictions, and rebuilding
			// would mask that drift — so in that case the block is left absent.
			enhancedSecuritySettings := priorRule.Restrictions.EnhancedSecuritySettings
			if !enhancedSecuritySettings.IsNull() && !enhancedSecuritySettings.IsUnknown() {
				enhancedSecuritySettingsMap := make(map[string]attr.Value)
				restoreOmittedEnhancedSecurity(ctx, priorRule, enhancedSecuritySettingsMap, &resp.Diagnostics)

				var diags diag.Diagnostics
				enhancedSecuritySettings, diags = types.MapValue(types.StringType, enhancedSecuritySettingsMap)
				resp.Diagnostics.Append(diags...)
			}
			accessRule.Restrictions = &RestrictionsResourceModel{
				RedirectSBS:              priorRule.Restrictions.RedirectSBS,
				EnhancedSecuritySettings: enhancedSecuritySettings,
			}
		}

		// Convert Rules.
		//
		// The write path translates the deprecated conditions[].user_and_groups
		// into an extra TYPE_USERGROUP rule. rules[] is Required (not Computed),
		// so mapping that synthesized rule back into state would make the result
		// differ from the configuration and trip the framework's consistency
		// check. Strip it here, mirroring the write path's own decision about
		// whether it appended anything at all.
		apiRules := rule.Rules
		if hasUG {
			apiRules = reconcileSynthesizedUserGroupRule(apiRules, declaredScopeRules, ugScope)
		}

		rules := make([]RuleResourceModel, 0)
		for _, r := range apiRules {
			valuesValues := make([]attr.Value, 0)
			for _, value := range r.Values {
				valuesValues = append(valuesValues, types.StringValue(value))
			}
			values, diags := types.ListValue(types.StringType, valuesValues)
			resp.Diagnostics.Append(diags...)

			metadataMap := make(map[string]attr.Value)
			if r.Metadata != nil {
				for k, v := range r.Metadata {
					metadataMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			metadata, diags := types.MapValue(types.StringType, metadataMap)
			resp.Diagnostics.Append(diags...)

			rules = append(rules, RuleResourceModel{
				Type:      types.StringValue(r.Type),
				Operator:  types.StringValue(r.Operator),
				TagSource: types.StringValue(r.TagSource),
				TagKey:    types.StringValue(r.TagKey),
				Values:    values,
				Metadata:  metadata,
			})
		}
		accessRule.Rules = rules

		accessRules = append(accessRules, accessRule)
	}
	data.AccessRules = accessRules

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ruleContentSignature builds a reorder-stable key for an access rule from every
// semantically meaningful field that is known in the plan, excluding the Console-owned
// fields (id, name, priority). It lets an updated rule be matched to
// its prior-state counterpart so an omitted name/priority can be preserved even after the
// rules are reordered. Every discriminating field that a user can set in config is
// included so two rules that differ only in one of them (advanced_settings, restrictions,
// conditions[].user_and_groups, or rule metadata) do not share a signature and get their
// preserved name/priority swapped on reorder. Null and unknown scalar/collection values
// are normalized to a single sentinel so the signature computed from a prior-state rule
// matches the one computed from its (config-omitted) plan counterpart.
//
// access_native is deliberately excluded: it is Optional+Computed with no
// UseStateForUnknown, so on an update that omits it the plan value is unknown while the
// prior-state value is concrete. Including it would make an unchanged rule's plan and
// state signatures differ and break the very name/priority preservation this key exists
// for. In practice access_native is backend-derived, so two rules identical in all other
// config fields resolve to the same access_native and are not distinguished by it.
//
// includeMetadata controls whether the nested rule metadata participates in the key.
// Name/priority preservation (Update) passes true so two rules that differ only in
// metadata keep distinct signatures. Reorder-stable rule matching in ModifyPlan passes
// false: there the plan metadata is still unknown (about to be reconciled), so a
// metadata-inclusive key computed on the plan would never match the concrete prior-state
// key.
func ruleContentSignature(ctx context.Context, rule AccessRuleResourceModel, includeMetadata bool) string {
	const sentinel = "\x00"
	strSig := func(s types.String) string {
		if s.IsNull() || s.IsUnknown() {
			return sentinel
		}
		return s.ValueString()
	}
	boolSig := func(b types.Bool) string {
		if b.IsNull() || b.IsUnknown() {
			return sentinel
		}
		return strconv.FormatBool(b.ValueBool())
	}
	listSig := func(l types.List) string {
		if l.IsNull() || l.IsUnknown() {
			return sentinel
		}
		var vals []string
		l.ElementsAs(ctx, &vals, false)
		sort.Strings(vals)
		return encodeStringSlice(vals)
	}
	mapSig := func(m types.Map) string {
		if m.IsNull() || m.IsUnknown() {
			return sentinel
		}
		var mm map[string]string
		m.ElementsAs(ctx, &mm, false)
		return encodeStringMap(mm)
	}

	ruleParts := make([]string, 0, len(rule.Rules))
	for _, rr := range rule.Rules {
		parts := []string{
			strSig(rr.Type),
			strSig(rr.Operator),
			strSig(rr.TagSource),
			strSig(rr.TagKey),
			listSig(rr.Values),
		}
		if includeMetadata {
			parts = append(parts, mapSig(rr.Metadata))
		}
		ruleParts = append(ruleParts, strings.Join(parts, "\x1f"))
	}
	sort.Strings(ruleParts)

	condParts := make([]string, 0, len(rule.Conditions))
	for _, c := range rule.Conditions {
		condParts = append(condParts, strings.Join([]string{
			strSig(c.PlatformFilter),
			mapSig(c.UserAndGroups),
		}, "\x1f"))
	}
	sort.Strings(condParts)

	advSig := sentinel
	if rule.AdvancedSettings != nil {
		doParts := make([]string, 0, len(rule.AdvancedSettings.DomainOverrides))
		for _, do := range rule.AdvancedSettings.DomainOverrides {
			doParts = append(doParts, strings.Join([]string{
				strSig(do.FQDN),
				strSig(do.Type),
				listSig(do.LocationIDs),
			}, "\x1f"))
		}
		sort.Strings(doParts)
		advSig = strings.Join(doParts, "\x1e")
	}

	restrSig := sentinel
	if rule.Restrictions != nil {
		restrSig = strings.Join([]string{
			boolSig(rule.Restrictions.RedirectSBS),
			mapSig(rule.Restrictions.EnhancedSecuritySettings),
		}, "\x1f")
	}

	return strings.Join([]string{
		"access=" + strSig(rule.Access),
		"active=" + boolSig(rule.Active),
		"desc=" + strSig(rule.Description),
		"rules=" + strings.Join(ruleParts, "\x1e"),
		"conds=" + strings.Join(condParts, "\x1e"),
		"adv=" + advSig,
		"restr=" + restrSig,
	}, "|")
}

func (r *AccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Update - Updating access policy")
	var data AccessPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve Console-owned rule name/priority for rules that already exist. An omitted
	// name or priority is unknown in the plan; matching the rule to its prior-state
	// counterpart lets us re-send the previously stored value instead of clearing the
	// name or resetting the priority to the list index on an unrelated update. Match by
	// the ModifyPlan-reconciled rule id first (so an ordinary content edit still finds
	// its prior rule), then by content signature, then treat as new (positional).
	var priorState AccessPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Read configuration so an explicitly-configured rule id can be treated as
	// authoritative. Only a config-set id is trusted: the plan's id may be the
	// positional value UseStateForUnknown copied from the old list index when
	// ModifyPlan deferred (unknown values at plan time), which must not be matched
	// as if it were reconciled or it would lock in a reorder swap.
	var configState AccessPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Reconcile each plan rule to its prior-state counterpart with the same
	// content/name matcher ModifyPlan uses, so Console-owned fields (id, name,
	// priority, sub-rule metadata) survive an insert, reorder, or edit. This also
	// covers the case where the whole access_rules list was unknown at plan time
	// (ModifyPlan deferred): a reordered rule whose own value is the unresolved
	// output is recovered here by its unique name. An explicitly-configured rule
	// id is matched first so an authoritative identity is never re-paired
	// positionally; an omitted id is reconciled by content/name instead.
	trustedIDs := make([]string, len(data.AccessRules))
	for i := range data.AccessRules {
		if i < len(configState.AccessRules) {
			if id := configState.AccessRules[i].ID; !id.IsNull() && !id.IsUnknown() {
				trustedIDs[i] = id.ValueString()
			}
		}
	}
	priorMatch := matchAccessRulesToPrior(ctx, data.AccessRules, priorState.AccessRules, trustedIDs)

	// Convert Terraform model to API model
	policy := &AccessPolicy{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.Active.IsNull() {
		policy.Active = data.Active.ValueBool()
	}
	// On update the backend requires a policy priority; the schema keeps the
	// prior (Console-assigned) value known via UseStateForUnknown, so send it.
	// Only omit when genuinely null/unknown.
	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		p := int(data.Priority.ValueInt64())
		policy.Priority = &p
	}

	// Convert Apps from Terraform set to string slice
	if !data.Apps.IsNull() && !data.Apps.IsUnknown() {
		var apps []string
		diags := data.Apps.ElementsAs(ctx, &apps, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		policy.Apps = apps
	}

	// Convert AccessRules from Terraform model to API model
	accessRules := make([]AccessRule, 0, len(data.AccessRules))
	// Resolve every rule's priority up front so an omitted one is placed above all
	// explicit priorities instead of the raw 1-based index, which could otherwise
	// duplicate an explicit value when a policy mixes explicit and omitted
	// priorities. A rule that omits its priority adopts its matched prior-state
	// value first (preserving the Console-assigned priority across an update).
	knownPrio := make([]*int, len(data.AccessRules))
	for i, tr := range data.AccessRules {
		pv := tr.Priority
		if priorMatch[i] != nil && pv.IsUnknown() {
			pv = priorMatch[i].Priority
		}
		if !pv.IsNull() && !pv.IsUnknown() {
			p := int(pv.ValueInt64())
			knownPrio[i] = &p
		}
	}
	resolvedPrio := resolveRulePriorities(knownPrio)
	for i, terraformRule := range data.AccessRules {
		// The prior-state rule this plan rule maps to (nil for a genuinely new rule).
		priorRule := priorMatch[i]
		ruleName := terraformRule.Name
		if priorRule != nil && ruleName.IsUnknown() {
			ruleName = priorRule.Name
		}
		rulePriority := resolvedPrio[i]
		accessRule := AccessRule{
			Name:         ruleName.ValueString(),
			Description:  terraformRule.Description.ValueString(),
			Priority:     rulePriority,
			Active:       terraformRule.Active.ValueBool(),
			Access:       terraformRule.Access.ValueString(),
			AccessNative: terraformRule.AccessNative.ValueString(),
		}

		// Handle optional ID field. An explicitly-configured id is authoritative.
		// Otherwise prefer the reconciled prior rule's id (matched by content/name)
		// rather than the plan's own id, which may be the positional value
		// UseStateForUnknown copied from the old list index when ModifyPlan
		// deferred; falling back to the plan id only when there is no prior match,
		// and leaving it empty so the backend assigns a fresh one for a new rule.
		var cfgID types.String
		if i < len(configState.AccessRules) {
			cfgID = configState.AccessRules[i].ID
		}
		switch {
		case !cfgID.IsNull() && !cfgID.IsUnknown():
			accessRule.ID = cfgID.ValueString()
		case priorRule != nil:
			accessRule.ID = priorRule.ID.ValueString()
		case !terraformRule.ID.IsNull() && !terraformRule.ID.IsUnknown():
			accessRule.ID = terraformRule.ID.ValueString()
		}

		// Convert AdvancedSettings
		if terraformRule.AdvancedSettings != nil {
			advancedSettings := &AdvancedSettings{}

			// Convert DomainOverrides
			domainOverrides := make([]DomainOverride, 0, len(terraformRule.AdvancedSettings.DomainOverrides))
			for _, tfOverride := range terraformRule.AdvancedSettings.DomainOverrides {
				var locationIDs []string
				diags := tfOverride.LocationIDs.ElementsAs(ctx, &locationIDs, false)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}

				domainOverrides = append(domainOverrides, DomainOverride{
					FQDN:        tfOverride.FQDN.ValueString(),
					LocationIDs: locationIDs,
					Type:        tfOverride.Type.ValueString(),
				})
			}
			advancedSettings.DomainOverrides = domainOverrides
			accessRule.AdvancedSettings = advancedSettings
		}

		// Convert Conditions. The deprecated user_and_groups map is not sent; it
		// is translated into a TYPE_USERGROUP rule below.
		conditions := make([]Condition, 0, len(terraformRule.Conditions))
		for _, tfCondition := range terraformRule.Conditions {
			conditions = append(conditions, Condition{
				PlatformFilter: tfCondition.PlatformFilter.ValueString(),
			})
		}
		accessRule.Conditions = conditions

		// Convert Restrictions
		if terraformRule.Restrictions != nil {
			restrictions := &Restrictions{
				RedirectSBS: terraformRule.Restrictions.RedirectSBS.ValueBool(),
			}

			// Convert EnhancedSecuritySettings map
			if !terraformRule.Restrictions.EnhancedSecuritySettings.IsNull() && !terraformRule.Restrictions.EnhancedSecuritySettings.IsUnknown() {
				enhancedSecurityStringMap := make(map[string]string)
				diags := terraformRule.Restrictions.EnhancedSecuritySettings.ElementsAs(ctx, &enhancedSecurityStringMap, false)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				// Convert to map[string]interface{} for API compatibility
				enhancedSecurityMap := make(map[string]interface{})
				for k, v := range enhancedSecurityStringMap {
					enhancedSecurityMap[k] = v
				}
				restrictions.EnhancedSecuritySettings = enhancedSecurityMap
			}

			accessRule.Restrictions = restrictions
		}

		// When ModifyPlan deferred (whole access_rules unknown), a sub-rule's
		// metadata is still unknown here; recover it from the matched prior rule by
		// content so an existing rule keeps its metadata instead of the PUT dropping
		// it. Keyed FIFO by ruleMetadataMatchKey, scoped to this rule's prior match.
		var priorSubMeta map[string][]types.Map
		if priorRule != nil {
			priorSubMeta = make(map[string][]types.Map)
			for j := range priorRule.Rules {
				key := ruleMetadataMatchKey(ctx, priorRule.Rules[j])
				priorSubMeta[key] = append(priorSubMeta[key], priorRule.Rules[j].Metadata)
			}
		}

		// Convert Rules
		rules := make([]Rule, 0, len(terraformRule.Rules))
		for j, tfRule := range terraformRule.Rules {
			rule := Rule{
				Type:      tfRule.Type.ValueString(),
				Operator:  tfRule.Operator.ValueString(),
				TagSource: tfRule.TagSource.ValueString(),
				TagKey:    tfRule.TagKey.ValueString(),
			}

			// Convert Values list
			var values []string
			diags := tfRule.Values.ElementsAs(ctx, &values, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			rule.Values = values

			// Resolve metadata: prefer an explicit plan value; when omitted/unknown
			// (e.g. ModifyPlan deferred), adopt the content-matched prior sub-rule's
			// metadata so an existing rule does not lose it on the PUT. An explicit
			// value still consumes its matching prior slot to keep the FIFO aligned.
			effectiveMeta := tfRule.Metadata
			if priorRule != nil {
				key := ruleMetadataMatchKey(ctx, tfRule)
				if q := priorSubMeta[key]; len(q) > 0 {
					// Adopt the content-matched prior metadata when omitted; when the plan
					// value is explicit, keep it and consume only its equal prior slot so
					// an omitted same-key sibling still adopts the other prior value.
					adopt, rest := consumePriorMeta(q, tfRule.Metadata)
					priorSubMeta[key] = rest
					if adopt != nil {
						effectiveMeta = *adopt
					}
				} else if (effectiveMeta.IsNull() || effectiveMeta.IsUnknown()) && j < len(priorRule.Rules) {
					// The sub-rule's own value is unknown (its content key can't match a
					// prior slot), so recover metadata positionally within the
					// name-matched prior rule instead of dropping it on the PUT.
					effectiveMeta = priorRule.Rules[j].Metadata
				}
			}
			if !effectiveMeta.IsNull() && !effectiveMeta.IsUnknown() {
				metadataStringMap := make(map[string]string)
				diags := effectiveMeta.ElementsAs(ctx, &metadataStringMap, false)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
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

		// Translate the deprecated conditions[].user_and_groups maps into a
		// TYPE_USERGROUP rule, mirroring what the SPA Console does on save.
		ugScope, hasUG, ugDiags := userGroupScopeFromConditions(ctx, terraformRule.Conditions)
		resp.Diagnostics.Append(ugDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if hasUG {
			rules = appendUserGroupRule(rules, ugScope)
		}

		accessRule.Rules = rules

		accessRules = append(accessRules, accessRule)
	}
	policy.AccessRules = accessRules

	// Update the policy
	err := r.client.UpdateAccessPolicy(ctx, data.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update access policy, got error: %s", err))
		return
	}

	// Save the updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the updated policy to refresh computed values
	readReq := resource.ReadRequest{
		State: resp.State,
	}
	readResp := &resource.ReadResponse{
		State: resp.State,
	}

	r.Read(ctx, readReq, readResp)

	// Copy any diagnostics and the updated state
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

func (r *AccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Delete - Deleting access policy")
	var data AccessPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the policy
	err := r.client.DeleteAccessPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete access policy, got error: %s", err))
		return
	}
}

func (r *AccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.ImportState - Importing access policy", map[string]any{
		"import_id": req.ID,
	})
	// Import using the policy ID
	data := AccessPolicyResourceModel{
		ID:   types.StringValue(req.ID),
		Apps: types.SetNull(types.StringType),
	}

	// Set the initial state with just the ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Now read the full access policy data from the API
	readReq := resource.ReadRequest{
		State: resp.State,
	}
	readResp := &resource.ReadResponse{
		State: resp.State,
	}

	r.Read(ctx, readReq, readResp)

	// Copy any diagnostics and the updated state
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

// enhancedSecurityDefaults holds the backend default value for each enhanced-
// security setting whose backend field elides its default (EmitDefaultValue =
// false). The API omits such a key from its response when the stored value
// equals this default (and omits the whole restrictions object when every
// setting is at its default). The set matches the backend States-enum fields on
// EnhancedSecuritySettingsModel and was confirmed empirically against the live
// API: keyLoggingV1/screenCaptureV1/uploadV1 are dropped at "enabled", so
// "enabled" is their default. _browserV1 is intentionally absent because its
// backend field is always emitted (no default-elision to reconcile).
// proxyTrafficV1 is also absent: it has no enabled/disabled default (its values
// are direct/secureBrowse) and is echoed whenever set, so it is only ever
// omitted when cleared to null — which must surface as drift, not be treated as
// a reconcilable default. insecure_content_allowed_for_urls_v1 (snake_case, the
// backend's own key name) defaults to "disabled" and the API drops it from the
// response at that value, so it is included here; at "enabled" it is echoed.
var enhancedSecurityDefaults = map[string]string{
	"clipboardV1":                          "enabled",
	"downloadV1":                           "enabled",
	"printingV1":                           "enabled",
	"uploadV1":                             "enabled",
	"keyLoggingV1":                         "enabled",
	"screenCaptureV1":                      "enabled",
	"watermarkV1":                          "disabled",
	"insecure_content_allowed_for_urls_v1": "disabled",
}

// serverInjectedEnhancedSecurityKeys lists enhanced-security keys the backend
// always adds to its response even when the configuration never set them: the
// API injects _browserV1 whenever a restrictions block carries any enhanced-
// security setting. Left in the read-back map, such a key appears as a value the
// configuration lacks, so every plan shows a phantom removal and every apply
// fails with "inconsistent result after apply" (new element "_browserV1" has
// appeared). They are dropped from state unless the prior configuration set them
// explicitly, in which case the user-supplied value is preserved.
var serverInjectedEnhancedSecurityKeys = map[string]bool{
	"_browserV1": true,
}

// priorRestrictionsAllDefault reports whether a prior-state rule's restrictions
// block is semantically at its backend default: redirect_sbs is unset/false and
// every enhanced_security_settings key holds its known default. A key without a
// known default is treated as NON-default, so the block is not reconstructed and
// out-of-band drift surfaces instead of being masked. This is safe against
// re-introducing the vanish error: the only keys not in enhancedSecurityDefaults
// are _browserV1 (always emitted by the API) and proxyTrafficV1 (echoed whenever
// set), so a block that contains either with a real value is never fully elided
// and this reconstruction path is never reached for it.
func priorRestrictionsAllDefault(priorRule *AccessRuleResourceModel) bool {
	if priorRule == nil || priorRule.Restrictions == nil {
		return false
	}
	r := priorRule.Restrictions
	if !r.RedirectSBS.IsNull() && !r.RedirectSBS.IsUnknown() && r.RedirectSBS.ValueBool() {
		return false
	}
	ess := r.EnhancedSecuritySettings
	if ess.IsNull() || ess.IsUnknown() {
		return true
	}
	for k, v := range ess.Elements() {
		sv, ok := v.(types.String)
		if !ok || sv.IsNull() || sv.IsUnknown() {
			continue
		}
		def, known := enhancedSecurityDefaults[k]
		if !known || sv.ValueString() != def {
			return false
		}
	}
	return true
}

// matchPriorAccessRule returns the prior-state access rule that corresponds to
// an API-returned rule, matching by ID first, then Name, and finally falling
// back to positional index. Matching by identity keeps restored settings
// attached to the correct rule even if the backend reorders rules.
//
// Each prior rule is claimed at most once via the caller-supplied used set, so
// two rules sharing the same (schema-permitted, non-unique) name are matched
// one-to-one in order instead of both resolving to the first prior rule — which
// would otherwise cross-apply restored defaults between them. During Create the
// planned rules carry unknown IDs, so the ID pass is skipped and this name/
// positional disambiguation is what keeps each API rule paired with its own
// prior rule.
func matchPriorAccessRule(prior []AccessRuleResourceModel, apiRule AccessRule, idx int, used map[int]bool) *AccessRuleResourceModel {
	claim := func(i int) *AccessRuleResourceModel {
		if used != nil {
			used[i] = true
		}
		return &prior[i]
	}
	if apiRule.ID != "" {
		for i := range prior {
			if used[i] {
				continue
			}
			if !prior[i].ID.IsNull() && !prior[i].ID.IsUnknown() && prior[i].ID.ValueString() == apiRule.ID {
				return claim(i)
			}
		}
	}
	if apiRule.Name != "" {
		for i := range prior {
			if used[i] {
				continue
			}
			if !prior[i].Name.IsNull() && !prior[i].Name.IsUnknown() && prior[i].Name.ValueString() == apiRule.Name {
				// If the API rule and this prior rule both carry known, non-empty IDs
				// that differ, they are different rules (e.g. one was replaced out of
				// band) even though their names coincide, so do not pair them —
				// otherwise stale conditions or restrictions would be cross-applied to
				// the wrong rule. During Create/import the prior ID is unknown, so this
				// guard is inert and name matching still works.
				if apiRule.ID != "" && !prior[i].ID.IsNull() && !prior[i].ID.IsUnknown() && prior[i].ID.ValueString() != "" && prior[i].ID.ValueString() != apiRule.ID {
					continue
				}
				return claim(i)
			}
		}
	}
	if idx >= 0 && idx < len(prior) && !used[idx] {
		// Positional fallback. If the API rule and the positional prior rule both
		// carry known, non-empty IDs that differ, they are different rules (e.g.
		// one was replaced out of band), so do not pair them — otherwise stale
		// conditions or restrictions would be cross-applied to the wrong rule.
		// During Create/import the prior ID is unknown, so this guard is inert and
		// positional pairing still works.
		p := &prior[idx]
		if apiRule.ID != "" && !p.ID.IsNull() && !p.ID.IsUnknown() && p.ID.ValueString() != "" && p.ID.ValueString() != apiRule.ID {
			return nil
		}
		return claim(idx)
	}
	return nil
}

// restoreOmittedEnhancedSecurity re-adds enhanced-security keys the user
// configured in prior state but that the API dropped from its response because
// their value equals the backend default. Only keys with a known default are
// restored, and to that literal default so genuine out-of-band drift to a non-
// default value is still detected on refresh. A key without a known default
// (e.g. proxyTrafficV1) is only ever omitted when its value was cleared out of
// band, so it is deliberately left absent — restoring the prior value would
// mask that drift and stop Terraform from re-applying the intended setting.
func restoreOmittedEnhancedSecurity(ctx context.Context, priorRule *AccessRuleResourceModel, essMap map[string]attr.Value, diags *diag.Diagnostics) {
	if priorRule == nil || priorRule.Restrictions == nil {
		return
	}
	prior := priorRule.Restrictions.EnhancedSecuritySettings
	if prior.IsNull() || prior.IsUnknown() {
		return
	}
	priorESS := make(map[string]string)
	diags.Append(prior.ElementsAs(ctx, &priorESS, false)...)
	for k := range priorESS {
		if _, ok := essMap[k]; ok {
			continue
		}
		if def, known := enhancedSecurityDefaults[k]; known {
			essMap[k] = types.StringValue(def)
		}
	}
}

// stripServerInjectedEnhancedSecurity removes backend-injected enhanced-security
// keys (see serverInjectedEnhancedSecurityKeys) from essMap so state mirrors the
// configuration. A key is kept only when the prior rule set it explicitly, so a
// user who deliberately configures such a key is unaffected.
func stripServerInjectedEnhancedSecurity(priorRule *AccessRuleResourceModel, essMap map[string]attr.Value) {
	if len(essMap) == 0 {
		return
	}
	// On import there is no prior state to consult (priorRule == nil), so keep
	// the backend-injected keys as-is; stripping here would drop an
	// explicitly-set value and make ImportStateVerify diverge from apply.
	if priorRule == nil {
		return
	}
	priorKeys := make(map[string]bool)
	if priorRule.Restrictions != nil {
		prior := priorRule.Restrictions.EnhancedSecuritySettings
		if !prior.IsNull() && !prior.IsUnknown() {
			for k := range prior.Elements() {
				priorKeys[k] = true
			}
		}
	}
	for k := range serverInjectedEnhancedSecurityKeys {
		if !priorKeys[k] {
			delete(essMap, k)
		}
	}
}
