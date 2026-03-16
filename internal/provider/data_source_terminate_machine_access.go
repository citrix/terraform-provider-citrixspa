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
var _ datasource.DataSource = &TerminateMachineAccessDataSource{}

func NewTerminateMachineAccessDataSource() datasource.DataSource {
	return &TerminateMachineAccessDataSource{}
}

// TerminateMachineAccessDataSource defines the data source implementation.
type TerminateMachineAccessDataSource struct {
	client SPAClient
}

// TerminateMachineAccessDataSourceModel describes the data source data model.
type TerminateMachineAccessDataSourceModel struct {
	Machines []TerminateMachineAccessModel `tfsdk:"machines"`
	Offset   types.Int64                   `tfsdk:"offset"`
	Limit    types.Int64                   `tfsdk:"limit"`
}

// TerminateMachineAccessModel describes the individual machine data model.
type TerminateMachineAccessModel struct {
	ID          types.String `tfsdk:"id"`
	AccountName types.String `tfsdk:"account_name"`
	Name        types.String `tfsdk:"name"`
	DNSHostName types.String `tfsdk:"dns_host_name"`
	DomainName  types.String `tfsdk:"domain_name"`
	ObjectID    types.String `tfsdk:"object_id"`
	IDPType     types.String `tfsdk:"idp_type"`
	// CreatedTime types.String `tfsdk:"created_time"`
	Duration types.Int64 `tfsdk:"duration"`
}

func (d *TerminateMachineAccessDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminate_machine_access"
}

func (d *TerminateMachineAccessDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This describes the data source and its expected configuration and attributes.
		MarkdownDescription: "Terminate machine access data source provides a list of all machines with terminated access.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"machines": schema.ListNestedAttribute{
				MarkdownDescription: "List of machines with terminated access",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Machine ID",
							Computed:            true,
						},
						"account_name": schema.StringAttribute{
							MarkdownDescription: "Machine account name",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Machine name",
							Computed:            true,
						},
						"dns_host_name": schema.StringAttribute{
							MarkdownDescription: "DNS host name",
							Computed:            true,
						},
						"domain_name": schema.StringAttribute{
							MarkdownDescription: "Domain name",
							Computed:            true,
						},
						"object_id": schema.StringAttribute{
							MarkdownDescription: "Object ID",
							Computed:            true,
						},
						"idp_type": schema.StringAttribute{
							MarkdownDescription: "IDP type",
							Computed:            true,
						},
						// "created_time": schema.Int64Attribute{
						// 	MarkdownDescription: "Created time",
						// 	Computed:            true,
						// },
						"duration": schema.Int64Attribute{
							MarkdownDescription: "Duration",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TerminateMachineAccessDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerminateMachineAccessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerminateMachineAccessDataSourceModel

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
	tflog.Debug(ctx, "spa-terraform-provider: Reading terminate machine access data source", map[string]any{
		"offset": offset,
		"limit":  limit,
	})

	// Get terminate machine access from API
	machines, err := d.client.GetTerminateMachineAccess(ctx, offset, limit)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read terminate machine access, got error: %s", err))
		return
	}

	// Map API response to Terraform model
	data.Machines = make([]TerminateMachineAccessModel, len(machines.Items))
	for i, machine := range machines.Items {
		data.Machines[i] = TerminateMachineAccessModel{
			ID:          types.StringValue(machine.ID),
			AccountName: types.StringValue(machine.AccountName),
			Name:        types.StringValue(machine.Name),
			DNSHostName: types.StringValue(machine.DNSHostName),
			DomainName:  types.StringValue(machine.DomainName),
			ObjectID:    types.StringValue(machine.ObjectID),
			IDPType:     types.StringValue(machine.IDPType),
			// CreatedTime: types.StringValue(machine.CreatedTime),
			Duration: types.Int64Value(int64(machine.Duration)),
		}
	}

	// Set updated values
	data.Offset = types.Int64Value(int64(offset))
	if limit > 0 {
		data.Limit = types.Int64Value(int64(limit))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
