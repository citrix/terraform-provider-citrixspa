package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecurityGroupDataSource{}

func NewSecurityGroupDataSource() datasource.DataSource {
	return &SecurityGroupDataSource{}
}

type SecurityGroupDataSource struct {
	client SPAClient
}

type SecurityGroupDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AppIds         types.Set    `tfsdk:"app_ids"`
	System         types.Object `tfsdk:"system"`
	UnpublishedApp types.Object `tfsdk:"unpublished_app"`
	Modified       types.Int64  `tfsdk:"modified"`
}

func (d *SecurityGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (d *SecurityGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a SPA security group.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Security group identifier",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the security group",
				Optional:            true,
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
	}
}

func (d *SecurityGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var sg *SecurityGroup
	var err error

	if !data.ID.IsNull() {
		// Get by ID
		sg, err = d.client.GetSecurityGroup(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read security group, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() {
		// Get by name
		securityGroups, err := d.client.GetSecurityGroups(ctx, 0, -1, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read security groups, got error: %s", err))
			return
		}

		if len(securityGroups.SecurityGroups) == 0 {
			resp.Diagnostics.AddError("Security Group Not Found", fmt.Sprintf("No security group found with name: %s", data.Name.ValueString()))
			return
		}

		if len(securityGroups.SecurityGroups) > 1 {
			resp.Diagnostics.AddError("Multiple Security Groups Found", fmt.Sprintf("Multiple security groups found with name: %s", data.Name.ValueString()))
			return
		}

		sg = &securityGroups.SecurityGroups[0]
	} else {
		resp.Diagnostics.AddError("Missing Required Field", "Either 'id' or 'name' must be specified")
		return
	}

	// Map API response to data source model
	data.ID = types.StringValue(sg.ID)
	data.Name = types.StringValue(sg.Name)
	data.Modified = types.Int64Value(sg.Modified)

	// Handle app_ids list
	if len(sg.AppIds) > 0 {
		data.AppIds, _ = types.SetValueFrom(ctx, types.StringType, sg.AppIds)
	} else {
		data.AppIds, _ = types.SetValueFrom(ctx, types.StringType, []string{}) // Ensure it's an empty set if no app IDs
		//data.AppIds = types.SetNull(types.StringType)
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
	data.System, _ = types.ObjectValue(systemConfigAttrTypes, systemConfigAttrs)

	// Handle unpublished app configuration
	unpublishedAppConfigAttrs := map[string]attr.Value{
		"data_in":  types.StringValue(sg.UnpublishedApp.DataIn),
		"data_out": types.StringValue(sg.UnpublishedApp.DataOut),
	}
	data.UnpublishedApp, _ = types.ObjectValue(systemConfigAttrTypes, unpublishedAppConfigAttrs)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
