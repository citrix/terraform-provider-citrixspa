package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccessPoliciesDataSource{}

func NewAccessPoliciesDataSource() datasource.DataSource {
	return &AccessPoliciesDataSource{}
}

// AccessPoliciesDataSource defines the data source implementation.
type AccessPoliciesDataSource struct {
	client SPAClient
}

// AccessPoliciesDataSourceModel describes the data source data model.
type AccessPoliciesDataSourceModel struct {
	AccessPolicies []AccessPolicyListDataSourceModel `tfsdk:"access_policies"`
	Offset         types.Int64                       `tfsdk:"offset"`
	Limit          types.Int64                       `tfsdk:"limit"`
	Name           types.String                      `tfsdk:"name"`
	OrderBy        types.String                      `tfsdk:"orderby"`
}

// AccessPolicyListDataSourceModel describes a single access policy in the list (without policy-level conditions/actions)
type AccessPolicyListDataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	Description types.String                `tfsdk:"description"`
	Active      types.Bool                  `tfsdk:"active"`
	Priority    types.Int64                 `tfsdk:"priority"`
	Modified    types.String                `tfsdk:"modified"`
	Apps        types.Set                   `tfsdk:"apps"`
	AccessRules []AccessRuleDataSourceModel `tfsdk:"access_rules"`
}

func (d *AccessPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_policies"
}

func (d *AccessPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This describes the data source and its expected configuration and attributes.
		MarkdownDescription: "Access policies data source provides a list of all access policies. " +
			"Detailed information fetching (including advanced_settings and other fields not available in the listing response) is controlled by the provider's 'fetch_details_on_list' configuration.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by policy name",
				Optional:            true,
			},
			"orderby": schema.StringAttribute{
				MarkdownDescription: "Order by field (e.g., 'name', 'id')",
				Optional:            true,
			},
			"access_policies": schema.ListNestedAttribute{
				MarkdownDescription: "List of access policies",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Access policy ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Access policy name",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Access policy description",
							Computed:            true,
						},

						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether the policy is active",
							Computed:            true,
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "Policy priority",
							Computed:            true,
						},
						"apps": schema.SetAttribute{
							MarkdownDescription: "Set of application IDs associated with the access policy",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"modified": schema.StringAttribute{
							MarkdownDescription: "Time the access policy was last modified (ISO 8601, e.g. 2026-05-11T09:49:40Z)",
							Computed:            true,
						},
						"access_rules": schema.ListNestedAttribute{
							MarkdownDescription: "Access rules for the access policy",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "Access rule ID",
										Computed:            true,
									},
									"name": schema.StringAttribute{
										MarkdownDescription: "Access rule name",
										Computed:            true,
									},
									"description": schema.StringAttribute{
										MarkdownDescription: "Access rule description",
										Computed:            true,
									},
									"priority": schema.Int64Attribute{
										MarkdownDescription: "Access rule priority",
										Computed:            true,
									},
									"active": schema.BoolAttribute{
										MarkdownDescription: "Whether the access rule is active",
										Computed:            true,
									},
									"access": schema.StringAttribute{
										MarkdownDescription: "Access type (ACCESS_DENY, ACCESS_ALLOW)",
										Computed:            true,
									},
									"access_native": schema.StringAttribute{
										MarkdownDescription: "Native access type (ACCESS_DENY, ACCESS_ALLOW)",
										Computed:            true,
									},
									"advanced_settings": schema.SingleNestedAttribute{
										MarkdownDescription: "Advanced settings for the access rule",
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"domain_overrides": schema.ListNestedAttribute{
												MarkdownDescription: "Domain override settings",
												Computed:            true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"fqdn": schema.StringAttribute{
															MarkdownDescription: "Fully qualified domain name",
															Computed:            true,
														},
														"location_ids": schema.ListAttribute{
															MarkdownDescription: "Location IDs",
															Computed:            true,
															ElementType:         types.StringType,
														},
														"type": schema.StringAttribute{
															MarkdownDescription: "Domain override type",
															Computed:            true,
														},
													},
												},
											},
										},
									},
									"conditions": schema.ListNestedAttribute{
										MarkdownDescription: "Conditions for the access rule",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"platform_filter": schema.StringAttribute{
													MarkdownDescription: "Platform filter (PLATFORM_FILTER_MOBILE, PLATFORM_FILTER_PC, PLATFORM_FILTER_ANY)",
													Computed:            true,
												},
												"user_and_groups": schema.MapAttribute{
													MarkdownDescription: "User and groups configuration",
													Computed:            true,
													ElementType:         types.StringType,
												},
											},
										},
									},
									"restrictions": schema.SingleNestedAttribute{
										MarkdownDescription: "Restrictions for the access rule",
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"redirect_sbs": schema.BoolAttribute{
												MarkdownDescription: "Whether to redirect SBS",
												Computed:            true,
											},
											"enhanced_security_settings": schema.MapAttribute{
												MarkdownDescription: "Enhanced security settings",
												Computed:            true,
												ElementType:         types.StringType,
											},
										},
									},
									"rules": schema.ListNestedAttribute{
										MarkdownDescription: "Rules within the access rule",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													MarkdownDescription: "Rule type (TYPE_TAG, TYPE_USERGROUP, TYPE_PLATFORM, TYPE_MACHINEGROUP, TYPE_MULTIURLDOMAIN)",
													Computed:            true,
												},
												"operator": schema.StringAttribute{
													MarkdownDescription: "Rule operator (OPERATOR_EQ, OPERATOR_IN, etc.)",
													Computed:            true,
												},
												"tag_source": schema.StringAttribute{
													MarkdownDescription: "Tag source (NLS, CAS, EPA, ITM, ThirdPartyDevicePosture, CONTEXTUAL)",
													Computed:            true,
												},
												"tag_key": schema.StringAttribute{
													MarkdownDescription: "Tag key",
													Computed:            true,
												},
												"values": schema.ListAttribute{
													MarkdownDescription: "Rule values",
													Computed:            true,
													ElementType:         types.StringType,
												},
												"metadata": schema.MapAttribute{
													MarkdownDescription: "Rule metadata",
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
				},
			},
		},
	}
}

func (d *AccessPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(SPAClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected SPAClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *AccessPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccessPoliciesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values for pagination
	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}

	// Set limit to -1 (no limit) if not specified
	// Negative values mean no limit parameter will be sent to API
	limit := -1
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}

	nameFilter := ""
	if !data.Name.IsNull() {
		nameFilter = data.Name.ValueString()
	}

	orderBy := ""
	if !data.OrderBy.IsNull() {
		orderBy = data.OrderBy.ValueString()
	}

	tflog.Debug(ctx, "spa-terraform-provider: Reading access policies", map[string]any{
		"offset":  offset,
		"limit":   limit,
		"name":    nameFilter,
		"orderby": orderBy,
	})

	// Get access policies from API using detailed method
	// The AuthenticatedClient will automatically fetch details if fetch_details_on_list is enabled
	policies, err := d.client.GetAccessPoliciesDetailed(ctx, offset, limit, nameFilter, "name", false)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read access policies, got error: %s", err))
		return
	}

	// Map API response to Terraform data model
	data.Offset = types.Int64Value(int64(offset))
	data.Limit = types.Int64Value(int64(limit))
	data.Name = types.StringValue(nameFilter)

	// Convert access policies to Terraform format
	accessPolicies := make([]AccessPolicyListDataSourceModel, len(policies.Policies))
	for i, policy := range policies.Policies {

		// Map apps to Terraform list
		appsValues := make([]attr.Value, 0)
		if policy.Apps != nil {
			for _, app := range policy.Apps {
				appsValues = append(appsValues, types.StringValue(app))
			}
		}
		apps, _ := types.SetValue(types.StringType, appsValues)

		// Convert access rules using the complex nested structure
		accessRules := make([]AccessRuleDataSourceModel, 0)
		for _, rule := range policy.AccessRules {
			// Handle optional ID field - use null if empty
			var ruleID types.String
			if rule.ID != "" {
				ruleID = types.StringValue(rule.ID)
			} else {
				ruleID = types.StringNull()
			}

			accessRule := AccessRuleDataSourceModel{
				ID:           ruleID,
				Name:         types.StringValue(rule.Name),
				Description:  types.StringValue(rule.Description),
				Priority:     types.Int64Value(int64(rule.Priority)),
				Active:       types.BoolValue(rule.Active),
				Access:       types.StringValue(rule.Access),
				AccessNative: types.StringValue(rule.AccessNative),
			}

			// Convert AdvancedSettings
			if rule.AdvancedSettings != nil {
				advancedSettings := &AdvancedSettingsDataSourceModel{}

				// Convert DomainOverrides
				domainOverrides := make([]DomainOverrideDataSourceModel, 0)
				for _, override := range rule.AdvancedSettings.DomainOverrides {
					locationIDsValues := make([]attr.Value, 0)
					for _, locationID := range override.LocationIDs {
						locationIDsValues = append(locationIDsValues, types.StringValue(locationID))
					}
					locationIDs, diags := types.ListValue(types.StringType, locationIDsValues)
					resp.Diagnostics.Append(diags...)

					domainOverrides = append(domainOverrides, DomainOverrideDataSourceModel{
						FQDN:        types.StringValue(override.FQDN),
						LocationIDs: locationIDs,
						Type:        types.StringValue(override.Type),
					})
				}
				advancedSettings.DomainOverrides = domainOverrides
				accessRule.AdvancedSettings = advancedSettings
			} else {
				accessRule.AdvancedSettings = nil
			}

			// Convert Conditions
			conditions := make([]ConditionDataSourceModel, 0)
			for condIdx, condition := range rule.Conditions {
				tflog.Debug(ctx, "spa-terraform-provider: Processing access rule condition", map[string]any{
					"rule_id":          rule.ID,
					"rule_name":        rule.Name,
					"condition_index":  condIdx,
					"platform_filter":  condition.PlatformFilter,
					"user_and_groups":  condition.UserAndGroups,
					"condition_struct": fmt.Sprintf("%+v", condition),
				})

				userAndGroupsMap := make(map[string]attr.Value)
				if condition.UserAndGroups != nil {
					for k, v := range condition.UserAndGroups {
						userAndGroupsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
					}
				}
				userAndGroups, diags := types.MapValue(types.StringType, userAndGroupsMap)
				resp.Diagnostics.Append(diags...)

				conditions = append(conditions, ConditionDataSourceModel{
					PlatformFilter: types.StringValue(condition.PlatformFilter),
					UserAndGroups:  userAndGroups,
				})
			}
			accessRule.Conditions = conditions

			// Convert Restrictions
			if rule.Restrictions != nil {
				enhancedSecuritySettingsMap := make(map[string]attr.Value)
				if rule.Restrictions.EnhancedSecuritySettings != nil {
					for k, v := range rule.Restrictions.EnhancedSecuritySettings {
						enhancedSecuritySettingsMap[k] = types.StringValue(fmt.Sprintf("%v", v))
					}
				}
				enhancedSecuritySettings, diags := types.MapValue(types.StringType, enhancedSecuritySettingsMap)
				resp.Diagnostics.Append(diags...)

				accessRule.Restrictions = &RestrictionsDataSourceModel{
					RedirectSBS:              types.BoolValue(rule.Restrictions.RedirectSBS),
					EnhancedSecuritySettings: enhancedSecuritySettings,
				}
			}

			// Convert Rules
			rules := make([]RuleDataSourceModel, 0)
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

				rules = append(rules, RuleDataSourceModel{
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

		policyModel := AccessPolicyListDataSourceModel{
			ID:          types.StringValue(policy.ID),
			Name:        types.StringValue(policy.Name),
			Description: types.StringValue(policy.Description),
			Active:      types.BoolValue(policy.Active),
			Priority:    types.Int64Value(int64(policy.Priority)),
			Modified:    types.StringValue(policy.Modified),
			Apps:        apps,
			AccessRules: accessRules,
		}

		accessPolicies[i] = policyModel
	}

	// Set the structured data in the model
	data.AccessPolicies = accessPolicies

	tflog.Debug(ctx, "spa-terraform-provider: Successfully read access policies", map[string]any{
		"count": len(accessPolicies),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
