package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ApplicationsDataSource{}

func NewApplicationsDataSource() datasource.DataSource {
	return &ApplicationsDataSource{}
}

// ApplicationsDataSource defines the data source implementation.
type ApplicationsDataSource struct {
	client SPAClient
}

// ApplicationListDataSourceModel describes the data model for application list items (where SSO is a string)
type ApplicationListDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Description          types.String `tfsdk:"description"`
	URL                  types.String `tfsdk:"url"`
	Category             types.String `tfsdk:"category"`
	Hidden               types.Bool   `tfsdk:"hidden"`
	AgentlessAccess      types.Bool   `tfsdk:"agentless_access"`
	MobileSecurity       types.Bool   `tfsdk:"mobile_security"`
	SbsOnlyLaunch        types.Bool   `tfsdk:"sbs_only_launch"`
	UsingTemplate        types.Bool   `tfsdk:"using_template"`
	TemplateName         types.String `tfsdk:"template_name"`
	Icon                 types.String `tfsdk:"icon"`
	IconURL              types.String `tfsdk:"icon_url"`
	RelatedURLs          types.Set    `tfsdk:"related_urls"`
	Keywords             types.Set    `tfsdk:"keywords"`
	Locations            types.List   `tfsdk:"locations"`
	Policies             types.List   `tfsdk:"policies"`
	Destination          types.List   `tfsdk:"destination"`
	CustomProperties     types.Map    `tfsdk:"custom_properties"`
	CustomerDomainFields types.Map    `tfsdk:"customer_domain_fields"`
	SSO                  types.String `tfsdk:"sso"`
	State                types.String `tfsdk:"state"`
	PolicyCount          types.String `tfsdk:"policy_count"`
}

// ApplicationsDataSourceModel describes the data source data model.
type ApplicationsDataSourceModel struct {
	Applications []ApplicationListDataSourceModel `tfsdk:"applications"`
	Offset       types.Int64                      `tfsdk:"offset"`
	Limit        types.Int64                      `tfsdk:"limit"`
}

func (d *ApplicationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationsDataSource.Metadata - Setting data source metadata")
	resp.TypeName = req.ProviderTypeName + "_applications"
}

func (d *ApplicationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationsDataSource.Schema - Defining data source schema")
	resp.Schema = schema.Schema{
		// This describes the data source and its expected configuration and attributes.
		MarkdownDescription: "Applications data source provides a list of all applications. " +
			"Detailed information fetching is controlled by the provider's 'fetch_details_on_list' configuration.",

		Attributes: map[string]schema.Attribute{
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset for pagination",
				Optional:            true,
			},
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for pagination",
				Optional:            true,
			},
			"applications": schema.ListNestedAttribute{
				MarkdownDescription: "List of applications",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: ApplicationListAttributes,
				},
			},
		},
	}
}

func (d *ApplicationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationsDataSource.Configure - Configuring data source client")
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

func (d *ApplicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationsDataSource.Read - Reading applications")
	var data ApplicationsDataSourceModel

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
	tflog.Debug(ctx, "spa-terraform-provider: Reading applications data source", map[string]any{
		"offset": offset,
		"limit":  limit,
	})

	// Get applications from API using detailed method
	// The AuthenticatedClient will automatically fetch details if fetch_details_on_list is enabled
	applications, err := d.client.GetApplicationsDetailed(ctx, offset, limit, "", "", false)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read applications, got error: %s", err))
		return
	}

	// Map API response to Terraform model
	data.Applications = make([]ApplicationListDataSourceModel, len(applications.Applications))
	for i, app := range applications.Applications {
		appModel := ApplicationListDataSourceModel{
			ID:              types.StringValue(app.ID),
			Name:            types.StringValue(app.Name),
			Type:            types.StringValue(app.Type),
			Hidden:          types.BoolValue(app.Hidden),
			AgentlessAccess: types.BoolValue(app.AgentlessAccess),
			MobileSecurity:  types.BoolValue(app.MobileSecurity),
			SbsOnlyLaunch:   types.BoolValue(app.SbsOnlyLaunch),
			UsingTemplate:   types.BoolValue(app.UsingTemplate),
			State:           types.StringValue(app.State),
			PolicyCount:     types.StringValue(app.PolicyCount),
		}

		// Handle optional string fields - set to null if empty string from API
		if app.Description != "" {
			appModel.Description = types.StringValue(app.Description)
		} else {
			appModel.Description = types.StringNull()
		}

		if app.URL != "" {
			appModel.URL = types.StringValue(app.URL)
		} else {
			appModel.URL = types.StringNull()
		}

		if app.Category != "" {
			appModel.Category = types.StringValue(app.Category)
		} else {
			appModel.Category = types.StringNull()
		}

		if app.TemplateName != "" {
			appModel.TemplateName = types.StringValue(app.TemplateName)
		} else {
			appModel.TemplateName = types.StringNull()
		}

		if app.Icon != "" {
			appModel.Icon = types.StringValue(app.Icon)
		} else {
			appModel.Icon = types.StringNull()
		}

		if app.IconURL != "" {
			appModel.IconURL = types.StringValue(app.IconURL)
		} else {
			appModel.IconURL = types.StringNull()
		}

		// Handle SSO string field
		if app.SSO != "" {
			appModel.SSO = types.StringValue(app.SSO)
		} else {
			appModel.SSO = types.StringNull()
		}

		data.Applications[i] = appModel

		// Convert RelatedURLs slice to Set type
		if len(app.RelatedURLs) > 0 {
			elements := make([]attr.Value, len(app.RelatedURLs))
			for j, url := range app.RelatedURLs {
				elements[j] = types.StringValue(url)
			}
			data.Applications[i].RelatedURLs, _ = types.SetValue(types.StringType, elements)
		} else {
			data.Applications[i].RelatedURLs = types.SetNull(types.StringType)
		}

		// Convert Keywords slice to Set type
		if len(app.Keywords) > 0 {
			elements := make([]attr.Value, len(app.Keywords))
			for j, keyword := range app.Keywords {
				elements[j] = types.StringValue(keyword)
			}
			data.Applications[i].Keywords, _ = types.SetValue(types.StringType, elements)
		} else {
			data.Applications[i].Keywords = types.SetNull(types.StringType)
		}

		// Convert Locations slice to List type
		if len(app.Locations) > 0 {
			locationElements := make([]attr.Value, len(app.Locations))
			for j, location := range app.Locations {
				locationObj, _ := types.ObjectValue(
					map[string]attr.Type{
						"name": types.StringType,
						"uuid": types.StringType,
					},
					map[string]attr.Value{
						"name": types.StringValue(location.Name),
						"uuid": types.StringValue(location.UUID),
					},
				)
				locationElements[j] = locationObj
			}
			data.Applications[i].Locations, _ = types.ListValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name": types.StringType,
					"uuid": types.StringType,
				},
			}, locationElements)
		} else {
			data.Applications[i].Locations = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name": types.StringType,
					"uuid": types.StringType,
				},
			})
		}

		// Convert Policies slice to List type
		if len(app.Policies) > 0 {
			// Convert Policy objects to PolicyModel
			policyModels := make([]PolicyModel, len(app.Policies))
			for j, policy := range app.Policies {
				// Convert the data map from API to terraform map
				dataElements := make(map[string]attr.Value)
				for key, value := range policy.Data {
					dataElements[key] = types.StringValue(fmt.Sprintf("%v", value))
				}

				policyModels[j] = PolicyModel{
					Type: types.StringValue(policy.Type),
					Data: types.MapValueMust(types.StringType, dataElements),
				}
			}
			data.Applications[i].Policies, _ = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type": types.StringType,
					"data": types.MapType{ElemType: types.StringType},
				},
			}, policyModels)
		} else {
			data.Applications[i].Policies = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type": types.StringType,
					"data": types.MapType{ElemType: types.StringType},
				},
			})
		}

		// Convert Destination slice to List type
		if len(app.Destination) > 0 {
			destinationModels := make([]DestinationModel, len(app.Destination))
			for j, dest := range app.Destination {
				destinationModels[j] = DestinationModel{
					Destination: types.StringValue(dest.Destination),
					Port:        types.StringValue(dest.Port),
					Protocol:    types.StringValue(dest.Protocol),
					Subtype:     types.StringValue(dest.Subtype),
				}
			}
			data.Applications[i].Destination, _ = types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"destination": types.StringType,
					"port":        types.StringType,
					"protocol":    types.StringType,
					"subtype":     types.StringType,
				},
			}, destinationModels)
		} else {
			data.Applications[i].Destination = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"destination": types.StringType,
					"port":        types.StringType,
					"protocol":    types.StringType,
					"subtype":     types.StringType,
				},
			})
		}

		// Convert CustomProperties map to Map type
		if len(app.CustomProperties) > 0 {
			elements := make(map[string]attr.Value)
			for key, value := range app.CustomProperties {
				if str, ok := value.(string); ok {
					// Simple string value
					elements[key] = types.StringValue(str)
				} else {
					// Complex value (dict, list, etc.) - encode as JSON
					if jsonBytes, err := json.Marshal(value); err == nil {
						elements[key] = types.StringValue(string(jsonBytes))
					} else {
						// Fallback to string representation
						elements[key] = types.StringValue(fmt.Sprintf("%v", value))
					}
				}
			}
			data.Applications[i].CustomProperties, _ = types.MapValue(types.StringType, elements)
		} else {
			data.Applications[i].CustomProperties = types.MapNull(types.StringType)
		}

		// Convert CustomerDomainFields map to Map type
		if len(app.CustomerDomainFields) > 0 {
			elements := make(map[string]attr.Value)
			for key, value := range app.CustomerDomainFields {
				elements[key] = types.StringValue(fmt.Sprintf("%v", value))
			}
			data.Applications[i].CustomerDomainFields, _ = types.MapValue(types.StringType, elements)
		} else {
			data.Applications[i].CustomerDomainFields = types.MapNull(types.StringType)
		}
	}

	data.Offset = types.Int64Value(int64(offset))
	data.Limit = types.Int64Value(int64(limit))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
