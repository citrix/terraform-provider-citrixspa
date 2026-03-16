package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TerminateUserAccessResource{}
var _ resource.ResourceWithImportState = &TerminateUserAccessResource{}

func NewTerminateUserAccessResource() resource.Resource {
	return &TerminateUserAccessResource{}
}

// TerminateUserAccessResource defines the resource implementation.
type TerminateUserAccessResource struct {
	client SPAClient
}

// TerminateUserAccessResourceModel describes the resource data model.
type TerminateUserAccessResourceModel struct {
	ID          types.String `tfsdk:"id"`
	AccountName types.String `tfsdk:"account_name"`
	Email       types.String `tfsdk:"email"`
	DomainName  types.String `tfsdk:"domain_name"`
	ObjectID    types.String `tfsdk:"object_id"`
	IDPType     types.String `tfsdk:"idp_type"`
	Duration    types.Int64  `tfsdk:"duration"`
}

func (r *TerminateUserAccessResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminate_user_access"
}

func (r *TerminateUserAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This describes the resource and its expected configuration and attributes.
		MarkdownDescription: "Terminate user access resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the terminate user access record",
			},
			"account_name": schema.StringAttribute{
				MarkdownDescription: "Account name for the user access termination",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the user",
				Required:            true,
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "Domain name for the user access termination",
				Required:            true,
			},
			"object_id": schema.StringAttribute{
				MarkdownDescription: "Object ID for the user access termination",
				Required:            true,
			},
			"idp_type": schema.StringAttribute{
				MarkdownDescription: "Identity provider type (e.g., AD, AAD)",
				Required:            true,
			},
			"duration": schema.Int64Attribute{
				MarkdownDescription: "Duration in days for the user access termination",
				Required:            true,
			},
		},
	}
}

func (r *TerminateUserAccessResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(SPAClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected SPAClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *TerminateUserAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TerminateUserAccessResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the user access termination record
	user := &TerminateUserAccess{
		AccountName: data.AccountName.ValueString(),
		Email:       data.Email.ValueString(),
		DomainName:  data.DomainName.ValueString(),
		ObjectID:    data.ObjectID.ValueString(),
		IDPType:     data.IDPType.ValueString(),
		Duration:    int(data.Duration.ValueInt64()),
	}

	createdUser, err := r.client.CreateTerminateUserAccess(ctx, user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create terminate user access, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.ID = types.StringValue(createdUser.ID)
	data.AccountName = types.StringValue(createdUser.AccountName)
	data.Email = types.StringValue(createdUser.Email)
	data.DomainName = types.StringValue(createdUser.DomainName)
	data.ObjectID = types.StringValue(createdUser.ObjectID)
	data.IDPType = types.StringValue(createdUser.IDPType)
	data.Duration = types.Int64Value(int64(createdUser.Duration))

	// Write logs using the tflog package
	tflog.Trace(ctx, "spa-terraform-provider: created a terminate user access resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminateUserAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TerminateUserAccessResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the user access termination record from the API
	user, err := r.client.GetTerminateUserAccessByID(ctx, data.ID.ValueString())
	if err != nil {
		// If the resource is not found, remove it from state
		if err.Error() == fmt.Sprintf("terminate user access with ID %s not found", data.ID.ValueString()) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read terminate user access, got error: %s", err))
		return
	}

	// Update the model with the current user access data
	data.AccountName = types.StringValue(user.AccountName)
	data.Email = types.StringValue(user.Email)
	data.DomainName = types.StringValue(user.DomainName)
	data.ObjectID = types.StringValue(user.ObjectID)
	data.IDPType = types.StringValue(user.IDPType)
	data.Duration = types.Int64Value(int64(user.Duration))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminateUserAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TerminateUserAccessResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the user access termination record
	user := &TerminateUserAccess{
		ID:          data.ID.ValueString(),
		AccountName: data.AccountName.ValueString(),
		Email:       data.Email.ValueString(),
		DomainName:  data.DomainName.ValueString(),
		ObjectID:    data.ObjectID.ValueString(),
		IDPType:     data.IDPType.ValueString(),
		Duration:    int(data.Duration.ValueInt64()),
	}

	err := r.client.UpdateTerminateUserAccess(ctx, data.ID.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update terminate user access, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminateUserAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TerminateUserAccessResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the user access termination record
	err := r.client.DeleteTerminateUserAccess(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete terminate user access, got error: %s", err))
		return
	}
}

func (r *TerminateUserAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the terminate user access ID
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Terminate user access ID cannot be empty.")
		return
	}

	// Set the ID in the model so the Read method can retrieve the full data
	data := TerminateUserAccessResourceModel{
		ID: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
