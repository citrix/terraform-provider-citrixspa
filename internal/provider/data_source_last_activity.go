package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &LastActivityDataSource{}

func NewLastActivityDataSource() datasource.DataSource {
	return &LastActivityDataSource{}
}

// LastActivityDataSource defines the data source implementation.
type LastActivityDataSource struct {
	client SPAClient
}

// LastActivityDataSourceModel describes the data source data model.
type LastActivityDataSourceModel struct {
	LastActivity types.Float64 `tfsdk:"last_activity"`
}

func (d *LastActivityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_last_activity"
}

func (d *LastActivityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Last activity data source",

		Attributes: map[string]schema.Attribute{
			"last_activity": schema.Float64Attribute{
				MarkdownDescription: "Last activity timestamp",
				Computed:            true,
			},
		},
	}
}

func (d *LastActivityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LastActivityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LastActivityDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the last activity from the API
	lastActivity, err := d.client.GetLastActivity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read last activity, got error: %s", err))
		return
	}

	// Update the model with the last activity data
	data.LastActivity = types.Float64Value(lastActivity.LastActivity)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
