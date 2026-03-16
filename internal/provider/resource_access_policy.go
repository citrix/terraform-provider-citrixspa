package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccessPolicyResource{}
var _ resource.ResourceWithImportState = &AccessPolicyResource{}

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
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the access policy is active",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority of the access policy",
				Optional:            true,
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
							MarkdownDescription: "Access rule name",
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Access rule description",
							Optional:            true,
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "Access rule priority",
							Required:            true,
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
										MarkdownDescription: "User and groups configuration",
										Optional:            true,
										ElementType:         types.StringType,
									},
								},
							},
						},
						"restrictions": schema.SingleNestedAttribute{
							MarkdownDescription: "Restrictions for the access rule",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"redirect_sbs": schema.BoolAttribute{
									MarkdownDescription: "Whether to redirect SBS",
									Optional:            true,
								},
								"enhanced_security_settings": schema.MapAttribute{
									MarkdownDescription: "Enhanced security settings",
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
										MarkdownDescription: "Rule type (TYPE_TAG, TYPE_USERGROUP, TYPE_PLATFORM, TYPE_MACHINEGROUP, TYPE_MULTIURLDOMAIN)",
										Required:            true,
									},
									"operator": schema.StringAttribute{
										MarkdownDescription: "Rule operator (OPERATOR_EQ, OPERATOR_IN, etc.)",
										Required:            true,
									},
									"tag_source": schema.StringAttribute{
										MarkdownDescription: "Tag source (NLS, CAS, EPA, ITM, ThirdPartyDevicePosture, CONTEXTUAL)",
										Optional:            true,
									},
									"tag_key": schema.StringAttribute{
										MarkdownDescription: "Tag key",
										Optional:            true,
									},
									"values": schema.ListAttribute{
										MarkdownDescription: "Rule values",
										Required:            true,
										ElementType:         types.StringType,
									},
									"metadata": schema.MapAttribute{
										MarkdownDescription: "Rule metadata",
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.UseStateForUnknown(),
										},
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

func (r *AccessPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Configure - Configuring resource client")
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
	if !data.Priority.IsNull() {
		policy.Priority = int(data.Priority.ValueInt64())
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
		for _, ruleData := range data.AccessRules {
			rule := AccessRule{
				Name:         ruleData.Name.ValueString(),
				Description:  ruleData.Description.ValueString(),
				Priority:     int(ruleData.Priority.ValueInt64()),
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

			// Convert Conditions
			if len(ruleData.Conditions) > 0 {
				conditions := make([]Condition, 0, len(ruleData.Conditions))
				for _, condData := range ruleData.Conditions {
					cond := Condition{
						PlatformFilter: condData.PlatformFilter.ValueString(),
					}
					if !condData.UserAndGroups.IsNull() && !condData.UserAndGroups.IsUnknown() {
						// Convert map[string]string to map[string]interface{}
						userAndGroupsStr := make(map[string]string)
						resp.Diagnostics.Append(condData.UserAndGroups.ElementsAs(ctx, &userAndGroupsStr, false)...)
						if resp.Diagnostics.HasError() {
							return
						}
						userAndGroups := make(map[string]interface{})
						for k, v := range userAndGroupsStr {
							userAndGroups[k] = v
						}
						cond.UserAndGroups = userAndGroups
					}
					conditions = append(conditions, cond)
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
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read access policy, got error: %s", err))
		return
	}

	// Update the model with the API response
	data.Name = types.StringValue(policy.Name)
	data.Active = types.BoolValue(policy.Active)
	data.Priority = types.Int64Value(int64(policy.Priority))
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
			for _, condition := range rule.Conditions {
				var userAndGroups types.Map
				userAndGroupsMap := make(map[string]attr.Value)
				if condition.UserAndGroups != nil {
					for k, v := range condition.UserAndGroups {
						userAndGroupsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
					}
				}
				if len(userAndGroupsMap) > 0 {
					var diags diag.Diagnostics
					userAndGroups, diags = types.MapValue(types.StringType, userAndGroupsMap)
					resp.Diagnostics.Append(diags...)
				} else {
					userAndGroups = types.MapNull(types.StringType)
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
			if ruleIdx < len(priorAccessRules) && priorAccessRules[ruleIdx].Conditions != nil {
				// Prior state had an explicit empty list — preserve it.
				accessRule.Conditions = []ConditionResourceModel{}
			}
			// else: prior was nil (null) — leave accessRule.Conditions nil (null).
		}

		// Convert Restrictions - only set if present in API response
		if rule.Restrictions != nil {
			enhancedSecuritySettingsMap := make(map[string]attr.Value)
			if rule.Restrictions.EnhancedSecuritySettings != nil {
				for k, v := range rule.Restrictions.EnhancedSecuritySettings {
					enhancedSecuritySettingsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
				}
			}
			enhancedSecuritySettings, diags := types.MapValue(types.StringType, enhancedSecuritySettingsMap)
			resp.Diagnostics.Append(diags...)

			accessRule.Restrictions = &RestrictionsResourceModel{
				RedirectSBS:              types.BoolValue(rule.Restrictions.RedirectSBS),
				EnhancedSecuritySettings: enhancedSecuritySettings,
			}
		}

		// Convert Rules
		rules := make([]RuleResourceModel, 0)
		for _, r := range rule.Rules {
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

func (r *AccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyResource.Update - Updating access policy")
	var data AccessPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	policy := &AccessPolicy{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.Active.IsNull() {
		policy.Active = data.Active.ValueBool()
	}
	if !data.Priority.IsNull() {
		policy.Priority = int(data.Priority.ValueInt64())
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
	for _, terraformRule := range data.AccessRules {
		accessRule := AccessRule{
			Name:         terraformRule.Name.ValueString(),
			Description:  terraformRule.Description.ValueString(),
			Priority:     int(terraformRule.Priority.ValueInt64()),
			Active:       terraformRule.Active.ValueBool(),
			Access:       terraformRule.Access.ValueString(),
			AccessNative: terraformRule.AccessNative.ValueString(),
		}

		// Handle optional ID field
		if !terraformRule.ID.IsNull() {
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

		// Convert Conditions
		conditions := make([]Condition, 0, len(terraformRule.Conditions))
		for _, tfCondition := range terraformRule.Conditions {
			condition := Condition{
				PlatformFilter: tfCondition.PlatformFilter.ValueString(),
			}

			// Convert UserAndGroups map
			if !tfCondition.UserAndGroups.IsNull() && !tfCondition.UserAndGroups.IsUnknown() {
				userAndGroupsStringMap := make(map[string]string)
				diags := tfCondition.UserAndGroups.ElementsAs(ctx, &userAndGroupsStringMap, false)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				// Convert to map[string]interface{} for API compatibility
				userAndGroupsMap := make(map[string]interface{})
				for k, v := range userAndGroupsStringMap {
					userAndGroupsMap[k] = v
				}
				condition.UserAndGroups = userAndGroupsMap
			}

			conditions = append(conditions, condition)
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

		// Convert Rules
		rules := make([]Rule, 0, len(terraformRule.Rules))
		for _, tfRule := range terraformRule.Rules {
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

			// Convert Metadata map
			if !tfRule.Metadata.IsNull() && !tfRule.Metadata.IsUnknown() {
				metadataStringMap := make(map[string]string)
				diags := tfRule.Metadata.ElementsAs(ctx, &metadataStringMap, false)
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
