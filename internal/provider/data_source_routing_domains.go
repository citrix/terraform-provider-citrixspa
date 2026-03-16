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
var _ datasource.DataSource = &RoutingDomainsDataSource{}

func NewRoutingDomainsDataSource() datasource.DataSource {
	return &RoutingDomainsDataSource{}
}

// RoutingDomainsDataSource defines the data source implementation.
type RoutingDomainsDataSource struct {
	client SPAClient
}

// RoutingDomainsDataSourceModel describes the data source data model.
type RoutingDomainsDataSourceModel struct {
	RoutingDomains []RoutingDomainDataSourceModel `tfsdk:"routing_domains"`
	Offset         types.Int64                    `tfsdk:"offset"`
	Limit          types.Int64                    `tfsdk:"limit"`
	Total          types.Int64                    `tfsdk:"total"`
}

func (d *RoutingDomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainsDataSource.Metadata - Setting data source metadata")
	resp.TypeName = req.ProviderTypeName + "_routing_domains"
}

func (d *RoutingDomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainsDataSource.Schema - Defining data source schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Routing domains data source provides a list of all routing domains.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of routing domains",
				Computed:            true,
			},
			"routing_domains": schema.ListNestedAttribute{
				MarkdownDescription: "List of routing domains",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"fqdn": schema.StringAttribute{
							MarkdownDescription: "Fully qualified domain name",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of routing entry",
							Computed:            true,
						},
						"app_type": schema.StringAttribute{
							MarkdownDescription: "Type of app binding to this routing entry",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Admin description for the routing entry",
							Computed:            true,
						},
						"flag": schema.StringAttribute{
							MarkdownDescription: "Whether routing entry is enabled or disabled",
							Computed:            true,
						},
						"error": schema.StringAttribute{
							MarkdownDescription: "Any error associated with this routing entry",
							Computed:            true,
						},
						"ip": schema.BoolAttribute{
							MarkdownDescription: "Whether the secure access app has IP-based configuration",
							Computed:            true,
						},
						"location_ids": schema.ListAttribute{
							MarkdownDescription: "List of resource location UUIDs",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *RoutingDomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainsDataSource.Configure - Configuring data source client")
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

func (d *RoutingDomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainsDataSource.Read - Reading routing domains")
	var data RoutingDomainsDataSourceModel

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
	tflog.Debug(ctx, "spa-terraform-provider: Reading routing domains data source", map[string]any{
		"offset": offset,
		"limit":  limit,
	})

	// Get routing domains from API
	routingDomains, err := d.client.GetRoutingDomains(ctx, offset, limit)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read routing domains, got error: %s", err))
		return
	}

	// Map API response to Terraform model
	data.RoutingDomains = make([]RoutingDomainDataSourceModel, len(routingDomains.RoutingDomains))
	for i, rd := range routingDomains.RoutingDomains {
		rdModel := RoutingDomainDataSourceModel{
			FQDN:  types.StringValue(rd.FQDN),
			Type:  types.StringValue(rd.Type),
			Error: types.StringValue(rd.Error),
			IP:    types.BoolValue(rd.IP),
		}

		// Handle optional string fields - set to null if empty string from API
		if rd.AppType != "" {
			rdModel.AppType = types.StringValue(rd.AppType)
		} else {
			rdModel.AppType = types.StringNull()
		}

		if rd.Comment != "" {
			rdModel.Comment = types.StringValue(rd.Comment)
		} else {
			rdModel.Comment = types.StringNull()
		}

		if rd.Flag != "" {
			rdModel.Flag = types.StringValue(rd.Flag)
		} else {
			rdModel.Flag = types.StringNull()
		}

		data.RoutingDomains[i] = rdModel

		// Convert LocationIds slice to List type
		if len(rd.LocationIds) > 0 {
			elements := make([]attr.Value, len(rd.LocationIds))
			for j, locationId := range rd.LocationIds {
				elements[j] = types.StringValue(locationId)
			}
			data.RoutingDomains[i].LocationIds, _ = types.ListValue(types.StringType, elements)
		} else {
			data.RoutingDomains[i].LocationIds = types.ListNull(types.StringType)
		}
	}

	// Set pagination information
	data.Offset = types.Int64Value(int64(offset))
	data.Limit = types.Int64Value(int64(limit))
	data.Total = types.Int64Value(int64(routingDomains.Total))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
