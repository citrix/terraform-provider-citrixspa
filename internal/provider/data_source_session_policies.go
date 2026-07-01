package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SessionPoliciesDataSource{}

func NewSessionPoliciesDataSource() datasource.DataSource {
	return &SessionPoliciesDataSource{}
}

// SessionPoliciesDataSource defines the data source implementation.
type SessionPoliciesDataSource struct {
	client SPAClient
}

// SessionPoliciesDataSourceModel describes the data source data model.
type SessionPoliciesDataSourceModel struct {
	SessionPolicies []SessionPolicyDataSourceModel `tfsdk:"session_policies"`
	Offset          types.Int64                    `tfsdk:"offset"`
	Limit           types.Int64                    `tfsdk:"limit"`
	Name            types.String                   `tfsdk:"name"`
	OrderBy         types.String                   `tfsdk:"orderby"`
}

func (d *SessionPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPoliciesDataSource.Metadata - Setting data source metadata")
	resp.TypeName = req.ProviderTypeName + "_session_policies"
}

func (d *SessionPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all SPA session policies for the customer.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of policies to return (default: all)",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter policies by exact name (case-insensitive)",
				Optional:            true,
			},
			"orderby": schema.StringAttribute{
				MarkdownDescription: "Field to sort by ('name' or 'priority'). Default: 'name'.",
				Optional:            true,
			},
			"session_policies": schema.ListNestedAttribute{
				MarkdownDescription: "List of session policies",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Session policy identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Session policy name",
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
				},
			},
		},
	}
}

func (d *SessionPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SessionPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SessionPoliciesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}

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

	tflog.Debug(ctx, "spa-terraform-provider: Reading session policies", map[string]any{
		"offset":  offset,
		"limit":   limit,
		"name":    nameFilter,
		"orderby": orderBy,
	})

	result, err := d.client.GetSessionPolicies(ctx, offset, limit, nameFilter, orderBy)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read session policies, got error: %s", err))
		return
	}

	sessionPolicies := make([]SessionPolicyDataSourceModel, 0, len(result.Items))
	for _, policy := range result.Items {
		policyModel := SessionPolicyDataSourceModel{
			ID:          types.StringValue(policy.ID),
			Name:        types.StringValue(policy.Name),
			Description: types.StringValue(policy.Description),
			Active:      types.BoolValue(policy.Active),
		}
		if policy.Priority != nil {
			policyModel.Priority = types.Int64Value(int64(*policy.Priority))
		} else {
			policyModel.Priority = types.Int64Value(0)
		}

		rules := make([]SessionPolicyRuleDataSourceModel, 0, len(policy.GenericRules))
		for _, rule := range policy.GenericRules {
			ruleModel := sessionPolicyRuleToDataSourceModel(ctx, rule, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
			rules = append(rules, ruleModel)
		}
		policyModel.Rules = rules

		sessionPolicies = append(sessionPolicies, policyModel)
	}
	data.SessionPolicies = sessionPolicies

	data.Offset = types.Int64Value(int64(offset))
	data.Limit = types.Int64Value(int64(limit))
	data.Name = types.StringValue(nameFilter)
	data.OrderBy = types.StringValue(orderBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
