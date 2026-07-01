package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SessionPolicyDataSource{}

func NewSessionPolicyDataSource() datasource.DataSource {
	return &SessionPolicyDataSource{}
}

// SessionPolicyDataSource defines the data source implementation.
type SessionPolicyDataSource struct {
	client SPAClient
}

// SessionPolicyDataSourceModel describes the data source data model.
type SessionPolicyDataSourceModel struct {
	ID          types.String                       `tfsdk:"id"`
	Name        types.String                       `tfsdk:"name"`
	Description types.String                       `tfsdk:"description"`
	Active      types.Bool                         `tfsdk:"active"`
	Priority    types.Int64                        `tfsdk:"priority"`
	Rules       []SessionPolicyRuleDataSourceModel `tfsdk:"generic_rules"`
}

// SessionPolicyRuleDataSourceModel describes a single rule within a session policy.
type SessionPolicyRuleDataSourceModel struct {
	ID          types.String                            `tfsdk:"id"`
	Name        types.String                            `tfsdk:"name"`
	Description types.String                            `tfsdk:"description"`
	Priority    types.Int64                             `tfsdk:"priority"`
	Active      types.Bool                              `tfsdk:"active"`
	Actions     *SessionPolicyActionsDataSourceModel    `tfsdk:"actions"`
	Conditions  []SessionPolicyConditionDataSourceModel `tfsdk:"condition"`
}

// SessionPolicyActionsDataSourceModel describes the actions block within a rule.
type SessionPolicyActionsDataSourceModel struct {
	Routing               types.String `tfsdk:"routing"`
	DisableSecurityGroups types.String `tfsdk:"disable_security_groups"`
	LocalLanAccess        types.String `tfsdk:"local_lan_access"`
}

// SessionPolicyConditionDataSourceModel describes a single condition within a rule.
type SessionPolicyConditionDataSourceModel struct {
	Type      types.String `tfsdk:"type"`
	Operator  types.String `tfsdk:"operator"`
	TagSource types.String `tfsdk:"tag_source"`
	TagKey    types.String `tfsdk:"tag_key"`
	Values    types.List   `tfsdk:"values"`
	Metadata  types.Map    `tfsdk:"metadata"`
}

func (d *SessionPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyDataSource.Metadata - Setting data source metadata")
	resp.TypeName = req.ProviderTypeName + "_session_policy"
}

func (d *SessionPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single SPA session policy by ID or name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Session policy identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Session policy name (used for lookup when ID is not provided)",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Session policy description",
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the session policy is active",
				Computed:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Session policy priority",
				Computed:            true,
			},
			"generic_rules": schema.ListNestedAttribute{
				MarkdownDescription: "List of rules within the session policy",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: sessionPolicyRuleDataSourceAttributes(),
				},
			},
		},
	}
}

func (d *SessionPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SessionPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SessionPolicyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var policy *SessionPolicy
	var err error

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		// Lookup by ID
		policy, err = d.client.GetSessionPolicy(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read session policy, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() && data.Name.ValueString() != "" {
		// Lookup by name via list filter
		policies, listErr := d.client.GetSessionPolicies(ctx, 0, -1, data.Name.ValueString(), "name")
		if listErr != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list session policies, got error: %s", listErr))
			return
		}
		if len(policies.Items) == 0 {
			resp.Diagnostics.AddError("Session Policy Not Found", fmt.Sprintf("No session policy found with name: %s", data.Name.ValueString()))
			return
		}
		if len(policies.Items) > 1 {
			resp.Diagnostics.AddError("Multiple Session Policies Found", fmt.Sprintf("Multiple session policies found with name: %s", data.Name.ValueString()))
			return
		}
		policy = &policies.Items[0]
	} else {
		resp.Diagnostics.AddError("Missing Required Field", "Either 'id' or 'name' must be specified")
		return
	}

	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Active = types.BoolValue(policy.Active)
	if policy.Priority != nil {
		data.Priority = types.Int64Value(int64(*policy.Priority))
	} else {
		data.Priority = types.Int64Value(0)
	}
	data.Description = types.StringValue(policy.Description)

	rules := make([]SessionPolicyRuleDataSourceModel, 0, len(policy.GenericRules))
	for _, rule := range policy.GenericRules {
		ruleModel := sessionPolicyRuleToDataSourceModel(ctx, rule, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		rules = append(rules, ruleModel)
	}
	data.Rules = rules

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ─── shared schema helpers ───────────────────────────────────────────────────

// sessionPolicyRuleDataSourceAttributes returns the nested attribute map shared by
// both the single and list data sources for a session policy rule.
func sessionPolicyRuleDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Rule identifier",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Rule name",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Rule description",
			Computed:            true,
		},
		"priority": schema.Int64Attribute{
			MarkdownDescription: "Rule priority",
			Computed:            true,
		},
		"active": schema.BoolAttribute{
			MarkdownDescription: "Whether the rule is active",
			Computed:            true,
		},
		"actions": schema.SingleNestedAttribute{
			MarkdownDescription: "Actions applied when the rule matches",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"routing": schema.StringAttribute{
					MarkdownDescription: "Routing direction",
					Computed:            true,
				},
				"disable_security_groups": schema.StringAttribute{
					MarkdownDescription: "Disable security groups flag",
					Computed:            true,
				},
				"local_lan_access": schema.StringAttribute{
					MarkdownDescription: "Local LAN access setting",
					Computed:            true,
				},
			},
		},
		"condition": schema.ListNestedAttribute{
			MarkdownDescription: "Conditions that must match for the rule to fire",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Condition type",
						Computed:            true,
					},
					"operator": schema.StringAttribute{
						MarkdownDescription: "Condition operator",
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
						MarkdownDescription: "Condition values",
						Computed:            true,
						ElementType:         types.StringType,
					},
					"metadata": schema.MapAttribute{
						MarkdownDescription: "Condition metadata",
						Computed:            true,
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

// sessionPolicyRuleToDataSourceModel converts an API SessionPolicyRule to a data source model.
func sessionPolicyRuleToDataSourceModel(ctx context.Context, rule SessionPolicyRule, diags *diag.Diagnostics) SessionPolicyRuleDataSourceModel {
	var ruleDescription types.String
	if rule.Description != "" {
		ruleDescription = types.StringValue(rule.Description)
	} else {
		ruleDescription = types.StringNull()
	}
	var ruleID types.String
	if rule.ID != "" {
		ruleID = types.StringValue(rule.ID)
	} else {
		ruleID = types.StringNull()
	}
	ruleModel := SessionPolicyRuleDataSourceModel{
		ID:          ruleID,
		Name:        types.StringValue(rule.Name),
		Description: ruleDescription,
		Priority:    types.Int64Value(int64(rule.Priority)),
		Active:      types.BoolValue(rule.Active),
	}

	// Map actions
	actionsModel := &SessionPolicyActionsDataSourceModel{
		Routing:               types.StringValue(rule.Actions.Routing),
		DisableSecurityGroups: types.StringValue(rule.Actions.DisableSecurityGroups),
		LocalLanAccess:        types.StringValue(rule.Actions.LocalLanAccess),
	}
	ruleModel.Actions = actionsModel

	// Map conditions
	conditions := make([]SessionPolicyConditionDataSourceModel, 0, len(rule.Conditions))
	for _, cond := range rule.Conditions {
		valuesVals := make([]attr.Value, 0, len(cond.Values))
		for _, v := range cond.Values {
			valuesVals = append(valuesVals, types.StringValue(v))
		}
		valuesList, d := types.ListValue(types.StringType, valuesVals)
		diags.Append(d...)

		metaAttrMap := make(map[string]attr.Value)
		if cond.Metadata != nil {
			for k, v := range cond.Metadata {
				metaAttrMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
		}
		var metadata types.Map
		if len(metaAttrMap) > 0 {
			var mapDiags diag.Diagnostics
			metadata, mapDiags = types.MapValue(types.StringType, metaAttrMap)
			diags.Append(mapDiags...)
		} else {
			metadata = types.MapNull(types.StringType)
		}

		tagSource := types.StringValue(cond.TagSource)
		tagKey := types.StringValue(cond.TagKey)
		conditions = append(conditions, SessionPolicyConditionDataSourceModel{
			Type:      types.StringValue(cond.Type),
			Operator:  types.StringValue(cond.Operator),
			TagSource: tagSource,
			TagKey:    tagKey,
			Values:    valuesList,
			Metadata:  metadata,
		})
	}
	ruleModel.Conditions = conditions

	return ruleModel
}
