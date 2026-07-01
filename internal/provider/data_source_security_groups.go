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
var _ datasource.DataSource = &SecurityGroupsDataSource{}

func NewSecurityGroupsDataSource() datasource.DataSource {
	return &SecurityGroupsDataSource{}
}

// SecurityGroupsDataSource defines the data source implementation.
type SecurityGroupsDataSource struct {
	client SPAClient
}

// SecurityGroupsDataSourceModel describes the data source data model.
type SecurityGroupsDataSourceModel struct {
	SecurityGroups []SecurityGroupDataSourceModel `tfsdk:"security_groups"`
	Offset         types.Int64                    `tfsdk:"offset"`
	Limit          types.Int64                    `tfsdk:"limit"`
}

func (d *SecurityGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

func (d *SecurityGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This describes the data source and its expected configuration and attributes.
		MarkdownDescription: "Security groups data source provides a list of all security groups.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"security_groups": schema.ListNestedAttribute{
				MarkdownDescription: "List of security groups",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Security group identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the security group",
							Computed:            true,
						},
						"app_ids": schema.SetAttribute{
							MarkdownDescription: "List of application IDs associated with the security group",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"system": schema.SingleNestedAttribute{
							MarkdownDescription: "System configuration settings for data in/out",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"data_in": schema.StringAttribute{
									MarkdownDescription: "Data in configuration (enabled or disabled)",
									Computed:            true,
								},
								"data_out": schema.StringAttribute{
									MarkdownDescription: "Data out configuration (enabled or disabled)",
									Computed:            true,
								},
							},
						},
						"unpublished_app": schema.SingleNestedAttribute{
							MarkdownDescription: "Unpublished app configuration settings for data in/out",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"data_in": schema.StringAttribute{
									MarkdownDescription: "Data in configuration (enabled or disabled)",
									Computed:            true,
								},
								"data_out": schema.StringAttribute{
									MarkdownDescription: "Data out configuration (enabled or disabled)",
									Computed:            true,
								},
							},
						},
						"modified": schema.Int64Attribute{
							MarkdownDescription: "Security group modified timestamp",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SecurityGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityGroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values if not provided
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

	// Log the data source read operation
	tflog.Debug(ctx, "spa-terraform-provider: Reading security groups data source", map[string]any{
		"offset": offset,
		"limit":  limit,
	})

	// Get security groups from API
	securityGroups, err := d.client.GetSecurityGroups(ctx, offset, limit)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read security groups, got error: %s", err))
		return
	}

	// Map API response to Terraform model
	data.SecurityGroups = make([]SecurityGroupDataSourceModel, len(securityGroups.SecurityGroups))
	for i, sg := range securityGroups.SecurityGroups {
		data.SecurityGroups[i] = SecurityGroupDataSourceModel{
			ID:       types.StringValue(sg.ID),
			Name:     types.StringValue(sg.Name),
			Modified: types.Int64Value(sg.Modified),
		}

		// Handle app_ids list
		if len(sg.AppIds) > 0 {
			data.SecurityGroups[i].AppIds, _ = types.SetValueFrom(ctx, types.StringType, sg.AppIds)
		} else {
			data.SecurityGroups[i].AppIds, _ = types.SetValueFrom(ctx, types.StringType, []string{})
			//data.SecurityGroups[i].AppIds = types.SetNull(types.StringType)
		}

		// Handle system configuration
		systemConfigAttrTypes := map[string]attr.Type{
			"data_in":  types.StringType,
			"data_out": types.StringType,
		}
		systemConfigAttrs := map[string]attr.Value{
			"data_in":  types.StringValue(sg.System.DataIn),
			"data_out": types.StringValue(sg.System.DataOut),
		}
		data.SecurityGroups[i].System, _ = types.ObjectValue(systemConfigAttrTypes, systemConfigAttrs)

		// Handle unpublished app configuration
		unpublishedAppConfigAttrs := map[string]attr.Value{
			"data_in":  types.StringValue(sg.UnpublishedApp.DataIn),
			"data_out": types.StringValue(sg.UnpublishedApp.DataOut),
		}
		data.SecurityGroups[i].UnpublishedApp, _ = types.ObjectValue(systemConfigAttrTypes, unpublishedAppConfigAttrs)
	}

	// Set updated values
	data.Offset = types.Int64Value(int64(offset))
	if limit > 0 {
		data.Limit = types.Int64Value(int64(limit))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
