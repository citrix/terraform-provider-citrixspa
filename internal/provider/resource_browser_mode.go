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
var _ resource.Resource = &BrowserModeResource{}

func NewBrowserModeResource() resource.Resource {
	return &BrowserModeResource{}
}

// BrowserModeResource defines the resource implementation.
type BrowserModeResource struct {
	client SPAClient
}

// BrowserModeResourceModel describes the resource data model.
type BrowserModeResourceModel struct {
	BrowserMode types.String `tfsdk:"browser_mode"`
}

func (r *BrowserModeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_browser_mode"
}

func (r *BrowserModeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Browser mode configuration resource",

		Attributes: map[string]schema.Attribute{
			"browser_mode": schema.StringAttribute{
				MarkdownDescription: "Browser mode setting (CEB or CEP)",
				Required:            true,
			},
		},
	}
}

func (r *BrowserModeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BrowserModeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BrowserModeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: Browser mode is typically read-only, but we'll implement this for completeness
	// In a real implementation, this might involve calling a different endpoint to set browser mode

	// For now, we'll just read the current browser mode and verify it matches
	browserMode, err := r.client.GetBrowserMode(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read browser mode, got error: %s", err))
		return
	}

	// Verify the browser mode matches what's requested
	if browserMode.BrowserMode != data.BrowserMode.ValueString() {
		resp.Diagnostics.AddError(
			"Browser Mode Mismatch",
			fmt.Sprintf("Expected browser mode %s, but current mode is %s", data.BrowserMode.ValueString(), browserMode.BrowserMode),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BrowserModeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BrowserModeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the current browser mode from the API
	browserMode, err := r.client.GetBrowserMode(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read browser mode, got error: %s", err))
		return
	}

	// Update the model with the current browser mode
	data.BrowserMode = types.StringValue(browserMode.BrowserMode)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BrowserModeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BrowserModeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note: Browser mode updates would typically require a different API endpoint
	// For now, we'll just read the current browser mode and verify it matches
	browserMode, err := r.client.GetBrowserMode(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read browser mode, got error: %s", err))
		return
	}

	// Verify the browser mode matches what's requested
	if browserMode.BrowserMode != data.BrowserMode.ValueString() {
		resp.Diagnostics.AddError(
			"Browser Mode Mismatch",
			fmt.Sprintf("Expected browser mode %s, but current mode is %s", data.BrowserMode.ValueString(), browserMode.BrowserMode),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BrowserModeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Note: Browser mode is typically a system-level setting that can't be deleted
	// This is a no-op for this resource type
}

func (r *BrowserModeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: Importing Browser Mode Resource")
	data := BrowserModeResourceModel{}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Now read the full application data from the API
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
