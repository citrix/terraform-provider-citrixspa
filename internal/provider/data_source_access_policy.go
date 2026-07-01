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

var _ datasource.DataSource = &AccessPolicyDataSource{}

func NewAccessPolicyDataSource() datasource.DataSource {
	return &AccessPolicyDataSource{}
}

type AccessPolicyDataSource struct {
	client SPAClient
}

type AccessPolicyDataSourceModel struct {
	ID           types.String                `tfsdk:"id"`
	Name         types.String                `tfsdk:"name"`
	Description  types.String                `tfsdk:"description"`
	Active       types.Bool                  `tfsdk:"active"`
	Priority     types.Int64                 `tfsdk:"priority"`
	Modified     types.String                `tfsdk:"modified"`
	Apps         types.Set                   `tfsdk:"apps"`
	AccessRules  []AccessRuleDataSourceModel `tfsdk:"access_rules"`
}

type AccessRuleDataSourceModel struct {
	ID               types.String                     `tfsdk:"id"`
	Name             types.String                     `tfsdk:"name"`
	Description      types.String                     `tfsdk:"description"`
	Priority         types.Int64                      `tfsdk:"priority"`
	Active           types.Bool                       `tfsdk:"active"`
	Access           types.String                     `tfsdk:"access"`
	AccessNative     types.String                     `tfsdk:"access_native"`
	AdvancedSettings *AdvancedSettingsDataSourceModel `tfsdk:"advanced_settings"`
	Conditions       []ConditionDataSourceModel       `tfsdk:"conditions"`
	Restrictions     *RestrictionsDataSourceModel     `tfsdk:"restrictions"`
	Rules            []RuleDataSourceModel            `tfsdk:"rules"`
}

type AdvancedSettingsDataSourceModel struct {
	DomainOverrides []DomainOverrideDataSourceModel `tfsdk:"domain_overrides"`
}

type DomainOverrideDataSourceModel struct {
	FQDN        types.String `tfsdk:"fqdn"`
	LocationIDs types.List   `tfsdk:"location_ids"`
	Type        types.String `tfsdk:"type"`
}

type ConditionDataSourceModel struct {
	PlatformFilter types.String `tfsdk:"platform_filter"`
	UserAndGroups  types.Map    `tfsdk:"user_and_groups"`
}

type RestrictionsDataSourceModel struct {
	RedirectSBS              types.Bool `tfsdk:"redirect_sbs"`
	EnhancedSecuritySettings types.Map  `tfsdk:"enhanced_security_settings"`
}

type RuleDataSourceModel struct {
	Type      types.String `tfsdk:"type"`
	Operator  types.String `tfsdk:"operator"`
	TagSource types.String `tfsdk:"tag_source"`
	TagKey    types.String `tfsdk:"tag_key"`
	Values    types.List   `tfsdk:"values"`
	Metadata  types.Map    `tfsdk:"metadata"`
}

func (d *AccessPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: AccessPolicyDataSource.Metadata - Setting data source metadata")
	resp.TypeName = req.ProviderTypeName + "_access_policy"
}

func (d *AccessPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a SPA access policy.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Access policy identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the access policy",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the access policy",
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the access policy is active",
				Computed:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority of the access policy",
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
							MarkdownDescription: "Access rule identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the access rule",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the access rule",
							Computed:            true,
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "Priority of the access rule",
							Computed:            true,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether the access rule is active",
							Computed:            true,
						},
						"access": schema.StringAttribute{
							MarkdownDescription: "Access type",
							Computed:            true,
						},
						"access_native": schema.StringAttribute{
							MarkdownDescription: "Native access type",
							Computed:            true,
						},
						"advanced_settings": schema.SingleNestedAttribute{
							MarkdownDescription: "Advanced settings for the access rule",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"domain_overrides": schema.ListNestedAttribute{
									MarkdownDescription: "Domain overrides",
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
												MarkdownDescription: "Type of domain override",
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
										MarkdownDescription: "Platform filter",
										Computed:            true,
									},
									"user_and_groups": schema.MapAttribute{
										MarkdownDescription: "User and groups",
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
									MarkdownDescription: "Redirect to SBS",
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
							MarkdownDescription: "Rules for the access rule",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "Rule type",
										Computed:            true,
									},
									"operator": schema.StringAttribute{
										MarkdownDescription: "Rule operator",
										Computed:            true,
									},
									"tag_source": schema.StringAttribute{
										MarkdownDescription: "Tag source",
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
	}
}

func (d *AccessPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(SPAClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected SPAClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *AccessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccessPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy *AccessPolicy
	var err error

	if !data.ID.IsNull() {
		// Get by ID
		policy, err = d.client.GetAccessPolicy(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read access policy, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() {
		// Get by name
		policies, err := d.client.GetAccessPolicies(ctx, 0, -1, data.Name.ValueString(), "name")
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read access policies, got error: %s", err))
			return
		}

		if len(policies.Policies) == 0 {
			resp.Diagnostics.AddError("Access Policy Not Found", fmt.Sprintf("No access policy found with name: %s", data.Name.ValueString()))
			return
		}

		if len(policies.Policies) > 1 {
			resp.Diagnostics.AddError("Multiple Access Policies Found", fmt.Sprintf("Multiple access policies found with name: %s", data.Name.ValueString()))
			return
		}

		policy = &policies.Policies[0]
	} else {
		resp.Diagnostics.AddError("Missing Required Field", "Either 'id' or 'name' must be specified")
		return
	}

	// Map API response to data source model
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Active = types.BoolValue(policy.Active)
	data.Priority = types.Int64Value(int64(policy.Priority))
	data.Description = types.StringValue(policy.Description)
	if policy.Modified != "" {
		data.Modified = types.StringValue(policy.Modified)
	} else {
		data.Modified = types.StringNull()
	}

	// Policy-level conditions and actions don't exist in the API - they are only at the access rule level

	// Map apps to Terraform list
	appsValues := make([]attr.Value, 0)
	if policy.Apps != nil {
		for _, app := range policy.Apps {
			appsValues = append(appsValues, types.StringValue(app))
		}
	}
	apps, diags := types.SetValue(types.StringType, appsValues)
	resp.Diagnostics.Append(diags...)
	data.Apps = apps

	// Map access_rules to complex nested structure
	accessRules := make([]AccessRuleDataSourceModel, 0)
	for _, rule := range policy.AccessRules {
		accessRule := AccessRuleDataSourceModel{
			ID:           types.StringValue(rule.ID),
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
		}

		// Convert Conditions
		conditions := make([]ConditionDataSourceModel, 0)
		for _, condition := range rule.Conditions {
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
	data.AccessRules = accessRules

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
