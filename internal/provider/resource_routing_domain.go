package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &RoutingDomainResource{}
var _ resource.ResourceWithImportState = &RoutingDomainResource{}

func NewRoutingDomainResource() resource.Resource {
	return &RoutingDomainResource{}
}

type RoutingDomainResource struct {
	client SPAClient
}

type RoutingDomainResourceModel struct {
	FQDN        types.String `tfsdk:"fqdn"`
	Type        types.String `tfsdk:"type"`
	AppType     types.String `tfsdk:"app_type"`
	Comment     types.String `tfsdk:"comment"`
	Flag        types.String `tfsdk:"flag"`
	Error       types.String `tfsdk:"error"`
	IP          types.Bool   `tfsdk:"ip"`
	LocationIds types.List   `tfsdk:"location_ids"`
}

func (r *RoutingDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Metadata - Setting resource metadata")
	resp.TypeName = req.ProviderTypeName + "_routing_domain"
}

func (r *RoutingDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Schema - Defining resource schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SPA routing domain.",

		Attributes: map[string]schema.Attribute{
			"fqdn": schema.StringAttribute{
				MarkdownDescription: "Fully qualified domain name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of routing entry (internal, external, conflicting, internal_bypass_proxy, internal_via_gateway, external_fixed_ip)",
				Required:            true,
			},
			"app_type": schema.StringAttribute{
				MarkdownDescription: "Type of app binding to this routing entry (ztna, web, saas)",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Admin description for the routing entry",
				Optional:            true,
			},
			"flag": schema.StringAttribute{
				MarkdownDescription: "Whether routing entry is enabled or disabled",
				Optional:            true,
			},
			"error": schema.StringAttribute{
				MarkdownDescription: "Any error associated with this routing entry",
				Optional:            true,
				Computed:            true,
			},
			"ip": schema.BoolAttribute{
				MarkdownDescription: "Whether the secure access app has IP-based configuration",
				Optional:            true,
			},
			"location_ids": schema.ListAttribute{
				MarkdownDescription: "List of resource location UUIDs",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *RoutingDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Configure - Configuring resource client")
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

func (r *RoutingDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Create - Creating routing domain")
	var data RoutingDomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	rd := &RoutingDomain{
		FQDN:        data.FQDN.ValueString(),
		Type:        data.Type.ValueString(),
		AppType:     data.AppType.ValueString(),
		Comment:     data.Comment.ValueString(),
		Flag:        data.Flag.ValueString(),
		Error:       "none",
		LocationIds: []string{}, // Initialize as empty array
	}

	if !data.IP.IsNull() {
		rd.IP = data.IP.ValueBool()
	}

	// Handle location IDs
	if !data.LocationIds.IsNull() && !data.LocationIds.IsUnknown() {
		var locationIds []string
		resp.Diagnostics.Append(data.LocationIds.ElementsAs(ctx, &locationIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		rd.LocationIds = locationIds
	}

	// Create the routing domain
	createdRD, err := r.client.CreateRoutingDomain(ctx, rd)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create routing domain, got error: %s", err))
		return
	}

	// Update the model with the created routing domain data
	data.Error = types.StringValue(createdRD.Error)

	// Set the initial state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the full routing domain data to populate all computed fields including
	// location_ids. postCreate is true: a 404 here is a read-back failure on a
	// routing domain that was just created, not out-of-band drift.
	r.read(ctx, resp.State, &resp.State, &resp.Diagnostics, true)
}

func (r *RoutingDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.read(ctx, req.State, &resp.State, &resp.Diagnostics, false)
}

// read refreshes a routing domain from the API into dst. It backs both Read and
// the post-create refresh in Create, which differ only in how a 404 is handled:
// on a refresh the routing domain is genuinely gone and must leave state, but
// straight after a successful create a 404 means the backend has not converged
// yet, and neither blaming an application cascade nor dropping the resource
// would be correct.
func (r *RoutingDomainResource) read(ctx context.Context, src tfsdk.State, dst *tfsdk.State, diags *diag.Diagnostics, postCreate bool) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Read - Reading routing domain")
	var data RoutingDomainResourceModel

	diags.Append(src.Get(ctx, &data)...)
	if diags.HasError() {
		return
	}

	// Get the routing domain from the API
	rd, err := r.client.GetRoutingDomain(ctx, data.FQDN.ValueString())
	if err != nil {
		if IsNotFound(err) {
			if postCreate {
				// Keep the state written by Create: the routing domain exists, the
				// API just has not made it readable yet. Removing it here would
				// fail the apply with "Missing Resource State After Create".
				diags.AddWarning(
					"Routing Domain Not Readable After Creation",
					fmt.Sprintf("Routing domain %q was created, but reading it back from Citrix Secure Private Access returned \"not found\". "+
						"This is usually a temporary delay in the service. The configured values have been recorded in the Terraform state, "+
						"and any computed values are refreshed by the next terraform plan.",
						data.FQDN.ValueString()),
				)
				return
			}
			// The backend deletes routing domains whose FQDN matches a deleted
			// application's URLs, so explain the likely drift before dropping state.
			diags.AddWarning(
				"Routing Domain Deleted Outside Terraform",
				fmt.Sprintf("Routing domain %q no longer exists in Citrix Secure Private Access. The most common cause is the deletion of an application: "+
					"Citrix Secure Private Access also deletes the routing domains whose FQDN matches that application's url, related_urls or destination "+
					"values, unless another application still references the same FQDN.\n\n"+
					"If this routing domain is no longer required, remove its resource block from the configuration. If it is still required, the next "+
					"terraform apply recreates it. See Note 3 in the citrixspa_routing_domain documentation for the full rule.",
					data.FQDN.ValueString()),
			)
			dst.RemoveResource(ctx)
			return
		}
		diags.AddError("Client Error", fmt.Sprintf("Unable to read routing domain, got error: %s", err))
		return
	}

	// Update the model with the API response
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

	// Handle location IDs - preserve what's in state if API returns empty
	if len(rd.LocationIds) > 0 {
		var locIds []string
		for _, locId := range rd.LocationIds {
			if locId != "" {
				locIds = append(locIds, locId)
			}
		}

		if len(locIds) > 0 {
			// Convert slice of strings to Terraform list type
			locationIds, listDiags := types.ListValueFrom(ctx, types.StringType, locIds)
			diags.Append(listDiags...)
			if diags.HasError() {
				return
			}
			data.LocationIds = locationIds
		}
	}

	// Save updated data into Terraform state
	diags.Append(dst.Set(ctx, &data)...)
}

func (r *RoutingDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Update - Updating routing domain")
	var data RoutingDomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the old FQDN from state to use in the API path (FQDN is the resource identifier)
	var stateData RoutingDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	oldFQDN := stateData.FQDN.ValueString()

	// Convert Terraform model to API model
	rd := &RoutingDomain{
		FQDN:        data.FQDN.ValueString(),
		Type:        data.Type.ValueString(),
		AppType:     data.AppType.ValueString(),
		Comment:     data.Comment.ValueString(),
		Flag:        data.Flag.ValueString(),
		Error:       "none",
		LocationIds: []string{}, // Initialize as empty array
	}

	if !data.IP.IsNull() {
		rd.IP = data.IP.ValueBool()
	}

	// Handle location IDs
	if !data.LocationIds.IsNull() && !data.LocationIds.IsUnknown() {
		var locationIds = []string{}
		resp.Diagnostics.Append(data.LocationIds.ElementsAs(ctx, &locationIds, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		rd.LocationIds = locationIds
	}

	// Update the routing domain using the old FQDN for the API paths
	err := r.client.UpdateRoutingDomain(ctx, oldFQDN, rd)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update routing domain, got error: %s", err))
		return
	}

	// Re-read the routing domain using the new FQDN
	updatedRD, err := r.client.GetRoutingDomain(ctx, data.FQDN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read routing domain after update, got error: %s", err))
		return
	}

	// Update only the computed error field
	data.Error = types.StringValue(updatedRD.Error)

	// Set the initial state with the plan data
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call Read to populate computed fields from API response
	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State, Diagnostics: resp.Diagnostics}
	r.Read(ctx, readReq, readResp)
	resp.State = readResp.State
	resp.Diagnostics = readResp.Diagnostics
}

func (r *RoutingDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Delete - Deleting routing domain")
	var data RoutingDomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// First, check if the routing domain needs to be disabled before deletion
	fqdn := data.FQDN.ValueString()

	// Read the current state from API to check the flag value
	currentRD, err := r.client.GetRoutingDomain(ctx, fqdn)
	if err != nil {
		// If 404, resource is already gone
		if IsNotFound(err) {
			tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Delete - Resource already deleted")
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read routing domain before deletion, got error: %s", err))
		return
	}

	// If flag is not "disabled", update it to "disabled" first
	if currentRD.Flag != "disabled" {
		tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Delete - Disabling routing domain before deletion")

		// Create update request with current values but flag set to disabled
		updateRD := &RoutingDomain{
			FQDN:        currentRD.FQDN,
			Type:        currentRD.Type,
			AppType:     currentRD.AppType,
			Comment:     currentRD.Comment,
			Flag:        "disabled",
			Error:       currentRD.Error,
			IP:          currentRD.IP,
			LocationIds: currentRD.LocationIds,
		}

		// Update the routing domain to disable it
		err = r.client.UpdateRoutingDomain(ctx, fqdn, updateRD)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable routing domain before deletion, got error: %s", err))
			return
		}

		tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Delete - Routing domain disabled successfully")
	} else {
		tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.Delete - Routing domain already disabled, proceeding with deletion")
	}

	// Now delete the routing domain
	err = r.client.DeleteRoutingDomain(ctx, fqdn)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete routing domain, got error: %s", err))
		return
	}
}

func (r *RoutingDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: RoutingDomainResource.ImportState - Importing routing domain", map[string]any{
		"fqdn": req.ID,
	})
	// Import using the FQDN
	emptyList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
	data := RoutingDomainResourceModel{
		FQDN:        types.StringValue(req.ID),
		LocationIds: emptyList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
