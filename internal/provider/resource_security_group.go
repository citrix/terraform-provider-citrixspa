package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SecurityGroupResource{}
var _ resource.ResourceWithImportState = &SecurityGroupResource{}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

type SecurityGroupResource struct {
	client SPAClient
}

type SecurityGroupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AppIds         types.Set    `tfsdk:"app_ids"`
	System         types.Object `tfsdk:"system"`
	UnpublishedApp types.Object `tfsdk:"unpublished_app"`
	Modified       types.Int64  `tfsdk:"modified"`
}

type ConfigurationSettingsModel struct {
	DataIn  types.String `tfsdk:"data_in"`
	DataOut types.String `tfsdk:"data_out"`
}

// configSettingsAttrTypes describes the attribute types of the data_in/data_out object.
var configSettingsAttrTypes = map[string]attr.Type{
	"data_in":  types.StringType,
	"data_out": types.StringType,
}

// dataFlowOrDefault returns the configured data-flow value, or "disabled" when the
// value is null, unknown, or empty. The SPA service treats data_in/data_out as
// optional and defaults them to "disabled"; because the API rejects an empty
// string against its enum, the provider substitutes the default before sending.
func dataFlowOrDefault(v types.String) string {
	if s := v.ValueString(); s != "" {
		return s
	}
	return "disabled"
}

// configSettingsFromObject decodes a data_in/data_out object into a resolved
// ConfigurationSettings, defaulting omitted (null/unknown/empty) values to
// "disabled" so the value sent to the SPA service is always a valid enum member.
func configSettingsFromObject(ctx context.Context, obj types.Object) (ConfigurationSettings, diag.Diagnostics) {
	var cfg ConfigurationSettingsModel
	diags := obj.As(ctx, &cfg, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	return ConfigurationSettings{
		DataIn:  dataFlowOrDefault(cfg.DataIn),
		DataOut: dataFlowOrDefault(cfg.DataOut),
	}, diags
}

// dataFlowAttribute returns the schema for a data_in/data_out flag. The value is
// optional and computed (defaulting to "disabled"); when set explicitly it must
// be one of the accepted enum members, so an empty string is rejected at plan
// time instead of being silently coerced.
func dataFlowAttribute(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.String{
			stringvalidator.OneOf("enabled", "disabled"),
		},
	}
}

// configSettingsSchema returns the schema for a system/unpublished_app data-flow block.
func configSettingsSchema(desc string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: desc,
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"data_in":  dataFlowAttribute("Data in configuration (`enabled` or `disabled`). Optional; defaults to `disabled` when omitted."),
			"data_out": dataFlowAttribute("Data out configuration (`enabled` or `disabled`). Optional; defaults to `disabled` when omitted."),
		},
	}
}

// configSettingsToObject builds the Terraform object value for a resolved
// ConfigurationSettings so computed data_in/data_out fields are stored as known values.
func configSettingsToObject(cs ConfigurationSettings) types.Object {
	// The attribute types are static and the values are always valid strings, so
	// construction cannot fail; ObjectValueMust panics loudly rather than
	// silently persisting an invalid object if that invariant is ever broken.
	return types.ObjectValueMust(configSettingsAttrTypes, map[string]attr.Value{
		"data_in":  types.StringValue(cs.DataIn),
		"data_out": types.StringValue(cs.DataOut),
	})
}

func (r *SecurityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SecurityGroupResource.Metadata - Setting resource metadata")
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SecurityGroupResource.Schema - Defining resource schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SPA security group.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Security group identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the security group",
				Required:            true,
			},
			"app_ids": schema.SetAttribute{
				MarkdownDescription: "List of application IDs associated with the security group",
				Required:            true,
				ElementType:         types.StringType,
			},
			"system":          configSettingsSchema("System configuration settings for data in/out"),
			"unpublished_app": configSettingsSchema("Unpublished app configuration settings for data in/out"),
			"modified": schema.Int64Attribute{
				MarkdownDescription: "Security group modified timestamp",
				Computed:            true,
			},
		},
	}
}

func (r *SecurityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SecurityGroupResource.Configure - Configuring resource client")
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(SPAClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected SPAClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	sg := &SecurityGroup{
		Name: data.Name.ValueString(),
	}

	// Handle app_ids
	if !data.AppIds.IsNull() {
		var appIds []string
		resp.Diagnostics.Append(data.AppIds.ElementsAs(ctx, &appIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// slices.Sort(appIds) // Ensure app IDs are sorted for consistency
		sg.AppIds = appIds
	} else {
		sg.AppIds = []string{} // Ensure appIds is initialized to an empty slice
	}

	// Resolve system/unpublished_app data-flow settings (default to "disabled" when omitted)
	var diags diag.Diagnostics
	sg.System, diags = configSettingsFromObject(ctx, data.System)
	resp.Diagnostics.Append(diags...)
	sg.UnpublishedApp, diags = configSettingsFromObject(ctx, data.UnpublishedApp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the security group
	createdSG, err := r.client.CreateSecurityGroup(ctx, sg)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create security group, got error: %s", err))
		return
	}

	// Update the model with the created security group data
	data.ID = types.StringValue(createdSG.ID)
	data.Modified = types.Int64Value(createdSG.Modified)
	// Persist the resolved (defaulted) data-flow settings so computed fields are known.
	data.System = configSettingsToObject(sg.System)
	data.UnpublishedApp = configSettingsToObject(sg.UnpublishedApp)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the security group from the API
	sg, err := r.client.GetSecurityGroup(ctx, data.ID.ValueString())
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read security group, got error: %s", err))
		return
	}

	// Update the model with the API response
	data.Name = types.StringValue(sg.Name)
	data.Modified = types.Int64Value(sg.Modified)

	// Handle app_ids list
	if len(sg.AppIds) > 0 {
		data.AppIds, _ = types.SetValueFrom(ctx, types.StringType, sg.AppIds)
	} else {
		data.AppIds, _ = types.SetValueFrom(ctx, types.StringType, []string{})
		//data.AppIds = types.SetNull(types.StringType)
	}

	// Handle system / unpublished app configuration using the shared attribute types.
	data.System = configSettingsToObject(sg.System)
	data.UnpublishedApp = configSettingsToObject(sg.UnpublishedApp)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	sg := &SecurityGroup{
		ID:   data.ID.ValueString(),
		Name: data.Name.ValueString(),
	}

	// Handle app_ids
	if !data.AppIds.IsNull() {
		var appIds []string
		resp.Diagnostics.Append(data.AppIds.ElementsAs(ctx, &appIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		sg.AppIds = appIds
	} else {
		sg.AppIds = []string{} // Ensure appIds is initialized to an empty slice
	}

	// Resolve system/unpublished_app data-flow settings (default to "disabled" when omitted)
	var diags diag.Diagnostics
	sg.System, diags = configSettingsFromObject(ctx, data.System)
	resp.Diagnostics.Append(diags...)
	sg.UnpublishedApp, diags = configSettingsFromObject(ctx, data.UnpublishedApp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Persist the resolved (defaulted) data-flow settings so computed fields are known.
	data.System = configSettingsToObject(sg.System)
	data.UnpublishedApp = configSettingsToObject(sg.UnpublishedApp)

	// Update the security group
	err := r.client.UpdateSecurityGroup(ctx, data.ID.ValueString(), sg)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update security group, got error: %s", err))
		return
	}

	// Fetch the updated security group to get the current computed values
	updatedSG, err := r.client.GetSecurityGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated security group, got error: %s", err))
		return
	}

	// Update computed fields
	data.Modified = types.Int64Value(updatedSG.Modified)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecurityGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the security group
	err := r.client.DeleteSecurityGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete security group, got error: %s", err))
		return
	}
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the security group ID
	data := SecurityGroupResourceModel{
		ID:             types.StringValue(req.ID),
		AppIds:         types.SetNull(types.StringType),
		System:         types.ObjectNull(configSettingsAttrTypes),
		UnpublishedApp: types.ObjectNull(configSettingsAttrTypes),
	}

	// Set the initial state with just the ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Now read the full security group data from the API
	readReq := resource.ReadRequest{
		State: resp.State,
	}
	readResp := &resource.ReadResponse{
		State: resp.State,
	}

	r.Read(ctx, readReq, readResp)

	// Copy any diagnostics and the updated state
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}
