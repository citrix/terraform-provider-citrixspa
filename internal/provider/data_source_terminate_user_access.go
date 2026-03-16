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
var _ datasource.DataSource = &TerminateUserAccessDataSource{}

func NewTerminateUserAccessDataSource() datasource.DataSource {
	return &TerminateUserAccessDataSource{}
}

// TerminateUserAccessDataSource defines the data source implementation.
type TerminateUserAccessDataSource struct {
	client SPAClient
}

// TerminateUserAccessDataSourceModel describes the data source data model.
type TerminateUserAccessDataSourceModel struct {
	Users  []TerminateUserAccessModel `tfsdk:"users"`
	Offset types.Int64                `tfsdk:"offset"`
	Limit  types.Int64                `tfsdk:"limit"`
}

// TerminateUserAccessModel describes the individual user data model.
type TerminateUserAccessModel struct {
	ID          types.String `tfsdk:"id"`
	AccountName types.String `tfsdk:"account_name"`
	Email       types.String `tfsdk:"email"`
	DomainName  types.String `tfsdk:"domain_name"`
	ObjectID    types.String `tfsdk:"object_id"`
	IDPType     types.String `tfsdk:"idp_type"`
	// CreatedTime types.String `tfsdk:"created_time"`
	Duration types.Int64 `tfsdk:"duration"`
}

func (d *TerminateUserAccessDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminate_user_access"
}

func (d *TerminateUserAccessDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This describes the data source and its expected configuration and attributes.
		MarkdownDescription: "Terminate user access data source provides a list of all users with terminated access.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users with terminated access",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "User ID",
							Computed:            true,
						},
						"account_name": schema.StringAttribute{
							MarkdownDescription: "User account name",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "User email",
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

func (d *TerminateUserAccessDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerminateUserAccessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerminateUserAccessDataSourceModel

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
	tflog.Debug(ctx, "spa-terraform-provider: Reading terminate user access data source", map[string]any{
		"offset": offset,
		"limit":  limit,
	})

	// Get terminate user access from API
	users, err := d.client.GetTerminateUserAccess(ctx, offset, limit)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read terminate user access, got error: %s", err))
		return
	}

	// Map API response to Terraform model
	data.Users = make([]TerminateUserAccessModel, len(users.Items))
	for i, user := range users.Items {
		data.Users[i] = TerminateUserAccessModel{
			ID:          types.StringValue(user.ID),
			AccountName: types.StringValue(user.AccountName),
			Email:       types.StringValue(user.Email),
			DomainName:  types.StringValue(user.DomainName),
			ObjectID:    types.StringValue(user.ObjectID),
			IDPType:     types.StringValue(user.IDPType),
			// CreatedTime: types.StringValue(user.CreatedTime),
			Duration: types.Int64Value(int64(user.Duration)),
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
