package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &HybridConfigDataSource{}

func NewHybridConfigDataSource() datasource.DataSource {
	return &HybridConfigDataSource{}
}

// HybridConfigDataSource defines the data source implementation.
type HybridConfigDataSource struct {
	client SPAClient
}

// HybridConfigDataSourceModel describes the data source data model.
type HybridConfigDataSourceModel struct {
	FirstTime types.Bool `tfsdk:"first_time"`
	IsHybrid  types.Bool `tfsdk:"is_hybrid"`
}

func (d *HybridConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hybrid_config"
}

func (d *HybridConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Hybrid configuration data source",

		Attributes: map[string]schema.Attribute{
			"first_time": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the first time hybrid configuration is being accessed",
				Computed:            true,
			},
			"is_hybrid": schema.BoolAttribute{
				MarkdownDescription: "Whether hybrid configuration is enabled",
				Computed:            true,
			},
		},
	}
}

func (d *HybridConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HybridConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HybridConfigDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the hybrid configuration from the API
	hybridConfig, err := d.client.GetHybridConfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read hybrid config, got error: %s", err))
		return
	}

	// Update the model with the hybrid configuration data
	data.FirstTime = types.BoolValue(hybridConfig.FirstTime)
	data.IsHybrid = types.BoolValue(hybridConfig.IsHybrid)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
