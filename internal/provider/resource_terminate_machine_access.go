package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TerminateMachineAccessResource{}
var _ resource.ResourceWithImportState = &TerminateMachineAccessResource{}

func NewTerminateMachineAccessResource() resource.Resource {
	return &TerminateMachineAccessResource{}
}

// TerminateMachineAccessResource defines the resource implementation.
type TerminateMachineAccessResource struct {
	client SPAClient
}

// TerminateMachineAccessResourceModel describes the resource data model.
type TerminateMachineAccessResourceModel struct {
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

func (r *TerminateMachineAccessResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terminate_machine_access"
}

func (r *TerminateMachineAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terminate machine access resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Machine access termination ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_name": schema.StringAttribute{
				MarkdownDescription: "Machine account name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Machine name. Optional; the SPA service does not require it (only account_name, object_id and idp_type identify the machine).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dns_host_name": schema.StringAttribute{
				MarkdownDescription: "DNS host name",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "Domain name",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.StringAttribute{
				MarkdownDescription: "Object ID. Required; identifies the machine together with account_name and idp_type.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"idp_type": schema.StringAttribute{
				MarkdownDescription: "IDP type. Required; identifies the machine together with account_name and object_id.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// "created_time": schema.Int64Attribute{
			// 	MarkdownDescription: "Created time",
			// 	Optional:            true,
			// 	Computed:            true,
			// },
			"duration": schema.Int64Attribute{
				MarkdownDescription: "Duration",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *TerminateMachineAccessResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TerminateMachineAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TerminateMachineAccessResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the machine access termination record
	machine := &TerminateMachineAccess{
		AccountName: data.AccountName.ValueString(),
		Name:        data.Name.ValueString(),
		DNSHostName: data.DNSHostName.ValueString(),
		DomainName:  data.DomainName.ValueString(),
		ObjectID:    data.ObjectID.ValueString(),
		IDPType:     data.IDPType.ValueString(),
		// CreatedTime: data.CreatedTime.ValueString(),
		Duration: int(data.Duration.ValueInt64()),
	}

	createdMachine, err := r.client.CreateTerminateMachineAccess(ctx, machine)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create terminate machine access, got error: %s", err))
		return
	}

	// The create API only echoes back id and object_id, so re-read the full
	// stored record to resolve any omitted (unknown) optional/computed fields. If
	// the record is not yet visible in the list API (eventual consistency) or the
	// list call fails, fall back to the create response rather than failing apply.
	storedMachine, err := r.client.GetTerminateMachineAccessByID(ctx, createdMachine.ID)
	if err != nil {
		tflog.Warn(ctx, "spa-terraform-provider: terminate machine access not readable immediately after create; using create response", map[string]any{"id": createdMachine.ID})
		storedMachine = createdMachine
	}

	// The server-assigned ID is authoritative. Required identity fields keep their
	// configured (plan) values. For optional/computed fields, preserve the
	// user-configured value and only fall back to the stored record when the
	// configuration omitted the field (left it unknown).
	data.ID = types.StringValue(storedMachine.ID)
	if data.Name.IsUnknown() {
		data.Name = types.StringValue(storedMachine.Name)
	}
	if data.DNSHostName.IsUnknown() {
		data.DNSHostName = types.StringValue(storedMachine.DNSHostName)
	}
	if data.DomainName.IsUnknown() {
		data.DomainName = types.StringValue(storedMachine.DomainName)
	}
	if data.Duration.IsUnknown() {
		data.Duration = types.Int64Value(int64(storedMachine.Duration))
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminateMachineAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TerminateMachineAccessResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the machine access termination record from the API
	machine, err := r.client.GetTerminateMachineAccessByID(ctx, data.ID.ValueString())
	if err != nil {
		// If the resource is not found, remove it from state
		if err.Error() == fmt.Sprintf("terminate machine access with ID %s not found", data.ID.ValueString()) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read terminate machine access, got error: %s", err))
		return
	}

	// Update the model with the current machine access data
	data.AccountName = types.StringValue(machine.AccountName)
	data.Name = types.StringValue(machine.Name)
	data.DNSHostName = types.StringValue(machine.DNSHostName)
	data.DomainName = types.StringValue(machine.DomainName)
	data.ObjectID = types.StringValue(machine.ObjectID)
	data.IDPType = types.StringValue(machine.IDPType)
	// data.CreatedTime = types.StringValue(machine.CreatedTime)
	data.Duration = types.Int64Value(int64(machine.Duration))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminateMachineAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TerminateMachineAccessResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Machine access termination records cannot be updated via the API
	// They can only be created and deleted
	resp.Diagnostics.AddWarning(
		"Terminate Machine Access Update",
		"Machine access termination records cannot be updated via the API. Consider using 'terraform apply -replace' to recreate the resource if needed.",
	)

	// Save data into Terraform state (no actual API call needed)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerminateMachineAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TerminateMachineAccessResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the machine access termination record
	err := r.client.DeleteTerminateMachineAccess(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete terminate machine access, got error: %s", err))
		return
	}
}

func (r *TerminateMachineAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the terminate machine access ID
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Terminate machine access ID cannot be empty.")
		return
	}

	// Set the ID in the model so the Read method can retrieve the full data
	data := TerminateMachineAccessResourceModel{
		ID: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
