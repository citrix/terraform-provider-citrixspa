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
var _ datasource.DataSource = &CertificatesDataSource{}

func NewCertificatesDataSource() datasource.DataSource {
	return &CertificatesDataSource{}
}

// CertificatesDataSource defines the data source implementation.
type CertificatesDataSource struct {
	client SPAClient
}

// CertificatesDataSourceModel describes the data source data model.
type CertificatesDataSourceModel struct {
	Certificates []CertificateDataSourceModel `tfsdk:"certificates"`
	Offset       types.Int64                  `tfsdk:"offset"`
	Limit        types.Int64                  `tfsdk:"limit"`
}

// CertificateDataSourceModel describes the individual certificate data model.
type CertificateDataSourceModel struct {
	CertificateID   types.String `tfsdk:"certificate_id"`
	CertificateName types.String `tfsdk:"certificate_name"`
}

func (d *CertificatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: CertificatesDataSource.Metadata - Setting data source metadata")
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (d *CertificatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: CertificatesDataSource.Schema - Defining data source schema")
	resp.Schema = schema.Schema{
		// This describes the data source and its expected configuration and attributes.
		MarkdownDescription: "Certificates data source provides a list of all certificates.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"certificates": schema.ListNestedAttribute{
				MarkdownDescription: "List of certificates",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"certificate_id": schema.StringAttribute{
							MarkdownDescription: "Certificate ID",
							Computed:            true,
						},
						"certificate_name": schema.StringAttribute{
							MarkdownDescription: "Certificate name",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CertificatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: CertificatesDataSource.Configure - Configuring data source client")
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

func (d *CertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: CertificatesDataSource.Read - Reading certificates")
	var data CertificatesDataSourceModel

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
	tflog.Debug(ctx, "spa-terraform-provider: Reading certificates data source", map[string]any{
		"offset": offset,
		"limit":  limit,
	})

	// Get certificates from API
	certificates, err := d.client.GetCertificates(ctx, offset, limit)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read certificates, got error: %s", err))
		return
	}

	// Map API response to Terraform model
	data.Certificates = make([]CertificateDataSourceModel, len(certificates.Certificates))
	for i, cert := range certificates.Certificates {
		data.Certificates[i] = CertificateDataSourceModel{
			CertificateID:   types.StringValue(cert.CertificateID),
			CertificateName: types.StringValue(cert.CertificateName),
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
