package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &CertificateResource{}
var _ resource.ResourceWithImportState = &CertificateResource{}

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

type CertificateResource struct {
	client SPAClient
}

type CertificateResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CertificateID       types.String `tfsdk:"certificate_id"`
	CertificateName     types.String `tfsdk:"certificate_name"`
	Certificate         types.String `tfsdk:"certificate"`
	CertificatePassword types.String `tfsdk:"certificate_password"`
	ApplicationID       types.String `tfsdk:"application_id"`
	Domain              types.String `tfsdk:"domain"`
}

func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SPA certificate.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Certificate identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Certificate ID assigned by the system",
			},
			"certificate_name": schema.StringAttribute{
				MarkdownDescription: "Name of the certificate",
				Required:            true,
			},
			"certificate": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded certificate data",
				Required:            true,
				Sensitive:           true,
			},
			"certificate_password": schema.StringAttribute{
				MarkdownDescription: "Password for the certificate (if applicable)",
				Optional:            true,
				Sensitive:           true,
			},
			"application_id": schema.StringAttribute{
				MarkdownDescription: "Application ID to assign the certificate to",
				Optional:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "Domain to assign the certificate to",
				Optional:            true,
			},
		},
	}
}

func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertificateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	cert := &Certificate{
		CertificateName:     data.CertificateName.ValueString(),
		Certificate:         data.Certificate.ValueString(),
		CertificatePassword: data.CertificatePassword.ValueString(),
		ApplicationID:       data.ApplicationID.ValueString(),
		Domain:              data.Domain.ValueString(),
	}

	// Create the certificate
	createdCert, err := r.client.CreateCertificate(ctx, cert)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create certificate, got error: %s", err))
		return
	}

	// Update the model with the created certificate data
	data.ID = types.StringValue(createdCert.ID)
	data.CertificateID = types.StringValue(createdCert.CertificateID)

	// If application_id and domain are provided, assign the certificate to the application
	if !data.ApplicationID.IsNull() && !data.Domain.IsNull() {
		err = r.client.AssignCertificateToApplication(ctx, data.ApplicationID.ValueString(), data.Domain.ValueString(), createdCert)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to assign certificate to application, got error: %s", err))
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the certificates from the API (we'll need to find the specific one)
	certificates, err := r.client.GetCertificates(ctx, 0, -1)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read certificates, got error: %s", err))
		return
	}

	// Find the certificate with matching ID
	var cert *Certificate
	targetID := data.ID.ValueString()
	targetCertID := data.CertificateID.ValueString()

	for _, c := range certificates.Certificates {
		// Match by either the ID field or CertificateID field
		// Since API response shows certificateId is the primary identifier
		if c.CertificateID == targetID || c.CertificateID == targetCertID ||
			(c.ID != "" && (c.ID == targetID || c.ID == targetCertID)) {
			cert = &c
			break
		}
	}

	if cert == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the model with the API response
	data.CertificateID = types.StringValue(cert.CertificateID)
	data.CertificateName = types.StringValue(cert.CertificateName)
	// Note: We don't update the certificate data from the API response as it's sensitive
	// and may not be returned in the response

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CertificateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For certificates, we typically need to delete and recreate rather than update
	// This is because certificate data is usually immutable
	resp.Diagnostics.AddWarning(
		"Certificate Update",
		"Certificate updates typically require recreating the resource. Consider using 'terraform apply -replace' if needed.",
	)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If assigned to an application, unassign first
	if !data.ApplicationID.IsNull() && !data.Domain.IsNull() {
		err := r.client.UnassignCertificateFromApplication(ctx, data.ApplicationID.ValueString(), data.Domain.ValueString())
		if err != nil {
			// Log the error but don't fail the deletion
			resp.Diagnostics.AddWarning("Client Warning", fmt.Sprintf("Unable to unassign certificate from application, got error: %s", err))
		}
	}

	// Delete the certificate
	err := r.client.DeleteCertificate(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete certificate, got error: %s", err))
		return
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the certificate ID
	value := "null"

	if req.Identity != nil {
		value = req.Identity.Raw.String()
	}
	tflog.Debug(ctx, "spa-terraform-provider: Importing certificate with ID", map[string]interface{}{
		"certificate_id": req.ID,
		"value":          value,
	})

	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Certificate ID cannot be empty.")
		return
	}

	// Set both ID and CertificateID to the provided ID for proper matching in Read function
	data := CertificateResourceModel{
		ID:            types.StringValue(req.ID),
		CertificateID: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
