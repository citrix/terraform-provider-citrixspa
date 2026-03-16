package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &RoutingDomainDataSource{}

func NewRoutingDomainDataSource() datasource.DataSource {
	return &RoutingDomainDataSource{}
}

type RoutingDomainDataSource struct {
	client SPAClient
}

type RoutingDomainDataSourceModel struct {
	FQDN        types.String `tfsdk:"fqdn"`
	Type        types.String `tfsdk:"type"`
	AppType     types.String `tfsdk:"app_type"`
	Comment     types.String `tfsdk:"comment"`
	Flag        types.String `tfsdk:"flag"`
	Error       types.String `tfsdk:"error"`
	IP          types.Bool   `tfsdk:"ip"`
	LocationIds types.List   `tfsdk:"location_ids"`
}

func (d *RoutingDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_domain"
}

func (d *RoutingDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a SPA routing domain.",

		Attributes: map[string]schema.Attribute{
			"fqdn": schema.StringAttribute{
				MarkdownDescription: "Fully qualified domain name",
				Required:            true,
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
	}
}

func (d *RoutingDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	tflog.Debug(ctx, "spa-terraform-provider: Configuring RoutingDomainDataSource")
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

func (d *RoutingDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoutingDomainDataSourceModel

	tflog.Debug(ctx, "spa-terraform-provider: Reading RoutingDomainDataSource")
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the routing domain by FQDN
	rd, err := d.client.GetRoutingDomain(ctx, data.FQDN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read routing domain, got error: %s", err))
		return
	}

	// Map API response to data source model
	data.FQDN = types.StringValue(rd.FQDN)
	data.Type = types.StringValue(rd.Type)
	data.Error = types.StringValue(rd.Error)
	data.IP = types.BoolValue(rd.IP)

	// Handle optional string fields - set to null if empty string from API
	if rd.AppType != "" {
		data.AppType = types.StringValue(rd.AppType)
	} else {
		data.AppType = types.StringNull()
	}

	if rd.Comment != "" {
		data.Comment = types.StringValue(rd.Comment)
	} else {
		data.Comment = types.StringNull()
	}

	if rd.Flag != "" {
		data.Flag = types.StringValue(rd.Flag)
	} else {
		data.Flag = types.StringNull()
	}

	// Handle location IDs
	if len(rd.LocationIds) > 0 {
		var locIds []string
		for _, locId := range rd.LocationIds {
			if locId != "" {
				locIds = append(locIds, locId)
			}
		}
		if len(locIds) == 0 {
			data.LocationIds = types.ListNull(types.StringType)
		} else {
			locationIds, diags := types.ListValueFrom(ctx, types.StringType, locIds)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.LocationIds = locationIds
		}
	} else if rd.LocationIds != nil {
		locationIds, diags := types.ListValueFrom(ctx, types.StringType, []string{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.LocationIds = locationIds
	} else {
		data.LocationIds = types.ListNull(types.StringType)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
