package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ApplicationDataSource{}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

// ApplicationDataSource defines the data source implementation.
type ApplicationDataSource struct {
	client SPAClient
}

// ApplicationDataSourceModel describes the data source data model.
type ApplicationDataSourceModel struct {
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
	SSO                  SSOValue     `tfsdk:"sso"`
	State                types.String `tfsdk:"state"`
	PolicyCount          types.String `tfsdk:"policy_count"`
}

var ApplicationAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		MarkdownDescription: "Application identifier",
		Optional:            true,
		Computed:            true,
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Name of the application",
		Optional:            true,
		Computed:            true,
	},
	"type": schema.StringAttribute{
		MarkdownDescription: "Type of application",
		Computed:            true,
	},
	"description": schema.StringAttribute{
		MarkdownDescription: "Description of the application",
		Optional:            true,
		Computed:            true,
	},
	"url": schema.StringAttribute{
		MarkdownDescription: "Application URL",
		Optional:            true,
		Computed:            true,
	},
	"category": schema.StringAttribute{
		MarkdownDescription: "Category of the application",
		Optional:            true,
		Computed:            true,
	},
	"hidden": schema.BoolAttribute{
		MarkdownDescription: "Whether the application is hidden",
		Optional:            true,
		Computed:            true,
	},
	"agentless_access": schema.BoolAttribute{
		MarkdownDescription: "Agentless access enabled",
		Optional:            true,
		Computed:            true,
	},
	"mobile_security": schema.BoolAttribute{
		MarkdownDescription: "Mobile security enabled",
		Optional:            true,
		Computed:            true,
	},
	"sbs_only_launch": schema.BoolAttribute{
		MarkdownDescription: "SBS only launch enabled",
		Optional:            true,
		Computed:            true,
	},
	"using_template": schema.BoolAttribute{
		MarkdownDescription: "Using template",
		Optional:            true,
		Computed:            true,
	},
	"template_name": schema.StringAttribute{
		MarkdownDescription: "Template name",
		Optional:            true,
		Computed:            true,
	},
	"icon": schema.StringAttribute{
		MarkdownDescription: "Base64 encoded icon data",
		Optional:            true,
		Computed:            true,
	},
	"icon_url": schema.StringAttribute{
		MarkdownDescription: "Application icon URL",
		Optional:            true,
		Computed:            true,
	},
	"related_urls": schema.SetAttribute{
		MarkdownDescription: "Related URLs",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"keywords": schema.SetAttribute{
		MarkdownDescription: "Keywords",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"locations": schema.ListNestedAttribute{
		MarkdownDescription: "Locations",
		Optional:            true,
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Location name",
					Computed:            true,
				},
				"uuid": schema.StringAttribute{
					MarkdownDescription: "Location UUID",
					Computed:            true,
				},
			},
		},
	},
	"policies": schema.ListNestedAttribute{
		MarkdownDescription: "Policies",
		Optional:            true,
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "Policy type (e.g., capability)",
					Computed:            true,
				},
				"data": schema.MapAttribute{
					MarkdownDescription: "Policy data as key-value pairs",
					Computed:            true,
					ElementType:         types.StringType,
				},
			},
		},
	},
	"destination": schema.ListNestedAttribute{
		MarkdownDescription: "Destinations for ZTNA applications",
		Optional:            true,
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"destination": schema.StringAttribute{
					MarkdownDescription: "Destination (hostname or IP range)",
					Computed:            true,
				},
				"port": schema.StringAttribute{
					MarkdownDescription: "Port number",
					Computed:            true,
				},
				"protocol": schema.StringAttribute{
					MarkdownDescription: "Protocol (PROTOCOL_TCP, PROTOCOL_UDP)",
					Computed:            true,
				},
				"subtype": schema.StringAttribute{
					MarkdownDescription: "Subtype (SUBTYPE_HOSTNAME, SUBTYPE_IP_AND_CIDR, SUBTYPE_IP_RANGE)",
					Computed:            true,
				},
			},
		},
	},
	"custom_properties": schema.MapAttribute{
		MarkdownDescription: "Custom properties fields",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"customer_domain_fields": schema.MapAttribute{
		MarkdownDescription: "Customer domain fields",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"sso": schema.DynamicAttribute{
		MarkdownDescription: "SSO configuration - a map where values can be either strings or arrays of objects (each object is a map of string to string)",
		Optional:            true,
		Computed:            true,
		CustomType:          SSOType{},
	},
	"state": schema.StringAttribute{
		MarkdownDescription: "Application state",
		Optional:            true,
		Computed:            true,
	},
	"policy_count": schema.StringAttribute{
		MarkdownDescription: "Policy count",
		Optional:            true,
		Computed:            true,
	},
}

// ApplicationListAttributes defines the schema for applications in a listing (where SSO is a string)
var ApplicationListAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		MarkdownDescription: "Application identifier",
		Optional:            true,
		Computed:            true,
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Name of the application",
		Optional:            true,
		Computed:            true,
	},
	"type": schema.StringAttribute{
		MarkdownDescription: "Type of application",
		Computed:            true,
	},
	"description": schema.StringAttribute{
		MarkdownDescription: "Description of the application",
		Optional:            true,
		Computed:            true,
	},
	"url": schema.StringAttribute{
		MarkdownDescription: "Application URL",
		Optional:            true,
		Computed:            true,
	},
	"category": schema.StringAttribute{
		MarkdownDescription: "Category of the application",
		Optional:            true,
		Computed:            true,
	},
	"hidden": schema.BoolAttribute{
		MarkdownDescription: "Whether the application is hidden",
		Optional:            true,
		Computed:            true,
	},
	"agentless_access": schema.BoolAttribute{
		MarkdownDescription: "Agentless access enabled",
		Optional:            true,
		Computed:            true,
	},
	"mobile_security": schema.BoolAttribute{
		MarkdownDescription: "Mobile security enabled",
		Optional:            true,
		Computed:            true,
	},
	"sbs_only_launch": schema.BoolAttribute{
		MarkdownDescription: "SBS only launch enabled",
		Optional:            true,
		Computed:            true,
	},
	"using_template": schema.BoolAttribute{
		MarkdownDescription: "Using template",
		Optional:            true,
		Computed:            true,
	},
	"template_name": schema.StringAttribute{
		MarkdownDescription: "Template name",
		Optional:            true,
		Computed:            true,
	},
	"icon": schema.StringAttribute{
		MarkdownDescription: "Base64 encoded icon data",
		Optional:            true,
		Computed:            true,
	},
	"icon_url": schema.StringAttribute{
		MarkdownDescription: "Application icon URL",
		Optional:            true,
		Computed:            true,
	},
	"related_urls": schema.SetAttribute{
		MarkdownDescription: "Related URLs",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"keywords": schema.SetAttribute{
		MarkdownDescription: "Keywords",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"locations": schema.ListNestedAttribute{
		MarkdownDescription: "Locations",
		Optional:            true,
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Location name",
					Computed:            true,
				},
				"uuid": schema.StringAttribute{
					MarkdownDescription: "Location UUID",
					Computed:            true,
				},
			},
		},
	},
	"policies": schema.ListNestedAttribute{
		MarkdownDescription: "Policies",
		Optional:            true,
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					MarkdownDescription: "Policy type (e.g., capability)",
					Computed:            true,
				},
				"data": schema.MapAttribute{
					MarkdownDescription: "Policy data as key-value pairs",
					Computed:            true,
					ElementType:         types.StringType,
				},
			},
		},
	},
	"destination": schema.ListNestedAttribute{
		MarkdownDescription: "Destinations for ZTNA applications",
		Optional:            true,
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"destination": schema.StringAttribute{
					MarkdownDescription: "Destination (hostname or IP range)",
					Computed:            true,
				},
				"port": schema.StringAttribute{
					MarkdownDescription: "Port number",
					Computed:            true,
				},
				"protocol": schema.StringAttribute{
					MarkdownDescription: "Protocol (PROTOCOL_TCP, PROTOCOL_UDP)",
					Computed:            true,
				},
				"subtype": schema.StringAttribute{
					MarkdownDescription: "Subtype (SUBTYPE_HOSTNAME, SUBTYPE_IP_AND_CIDR, SUBTYPE_IP_RANGE)",
					Computed:            true,
				},
			},
		},
	},
	"custom_properties": schema.MapAttribute{
		MarkdownDescription: "Custom properties fields",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"customer_domain_fields": schema.MapAttribute{
		MarkdownDescription: "Customer domain fields",
		Optional:            true,
		Computed:            true,
		ElementType:         types.StringType,
	},
	"sso": schema.StringAttribute{
		MarkdownDescription: "SSO configuration (as string in application listings)",
		Optional:            true,
		Computed:            true,
	},
	"state": schema.StringAttribute{
		MarkdownDescription: "Application state",
		Optional:            true,
		Computed:            true,
	},
	"policy_count": schema.StringAttribute{
		MarkdownDescription: "Policy count",
		Optional:            true,
		Computed:            true,
	},
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a SPA application.",
		Attributes:          ApplicationAttributes,
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(SPAClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected SPAClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var app *Application
	var err error

	if !data.ID.IsNull() {
		// Get by ID
		app, err = d.client.GetApplication(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read application, got error: %s", err))
			return
		}
	} else if !data.Name.IsNull() {
		// Get by name - first search for the application, then get full details by ID
		applications, err := d.client.GetApplications(ctx, 0, -1, data.Name.ValueString(), "")
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read applications, got error: %s", err))
			return
		}

		if len(applications.Applications) == 0 {
			resp.Diagnostics.AddError("Application Not Found", fmt.Sprintf("No application found with name: %s", data.Name.ValueString()))
			return
		}

		if len(applications.Applications) > 1 {
			resp.Diagnostics.AddError("Multiple Applications Found", fmt.Sprintf("Multiple applications found with name: %s", data.Name.ValueString()))
			return
		}

		// Get full application details by ID
		app, err = d.client.GetApplication(ctx, applications.Applications[0].ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read application details, got error: %s", err))
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Required Field", "Either 'id' or 'name' must be specified")
		return
	}

	// Map API response to data source model
	data.ID = types.StringValue(app.ID)
	data.Name = types.StringValue(app.Name)
	data.Type = types.StringValue(app.Type)
	data.State = types.StringValue(app.State)
	data.PolicyCount = types.StringValue(app.PolicyCount)

	// Handle boolean fields
	data.Hidden = types.BoolValue(app.Hidden)
	data.AgentlessAccess = types.BoolValue(app.AgentlessAccess)
	data.MobileSecurity = types.BoolValue(app.MobileSecurity)
	data.SbsOnlyLaunch = types.BoolValue(app.SbsOnlyLaunch)
	data.UsingTemplate = types.BoolValue(app.UsingTemplate)

	// Handle optional string fields - set to null if empty string from API
	if app.Description != "" {
		data.Description = types.StringValue(app.Description)
	} else {
		data.Description = types.StringNull()
	}

	if app.URL != "" {
		data.URL = types.StringValue(app.URL)
	} else {
		data.URL = types.StringNull()
	}

	if app.Category != "" {
		data.Category = types.StringValue(app.Category)
	} else {
		data.Category = types.StringNull()
	}

	if app.TemplateName != "" {
		data.TemplateName = types.StringValue(app.TemplateName)
	} else {
		data.TemplateName = types.StringNull()
	}

	if app.Icon != "" {
		data.Icon = types.StringValue(app.Icon)
	} else {
		data.Icon = types.StringNull()
	}

	if app.IconURL != "" {
		data.IconURL = types.StringValue(app.IconURL)
	} else {
		data.IconURL = types.StringNull()
	}

	// Handle locations
	if len(app.Locations) > 0 {
		// Convert Location objects to LocationModel
		locationModels := make([]LocationModel, len(app.Locations))
		for i, loc := range app.Locations {
			locationModels[i] = LocationModel{
				Name: types.StringValue(loc.Name),
				UUID: types.StringValue(loc.UUID),
			}
		}
		data.Locations, _ = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name": types.StringType,
				"uuid": types.StringType,
			},
		}, locationModels)
	} else {
		data.Locations = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name": types.StringType,
				"uuid": types.StringType,
			},
		})
	}

	// Handle policies
	if len(app.Policies) > 0 {
		// Convert Policy objects to PolicyModel
		policyModels := make([]PolicyModel, len(app.Policies))
		for i, policy := range app.Policies {
			// Convert the data map from API to terraform map
			dataElements := make(map[string]attr.Value)
			for key, value := range policy.Data {
				dataElements[key] = types.StringValue(fmt.Sprintf("%v", value))
			}

			policyModels[i] = PolicyModel{
				Type: types.StringValue(policy.Type),
				Data: types.MapValueMust(types.StringType, dataElements),
			}
		}
		data.Policies, _ = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type": types.StringType,
				"data": types.MapType{ElemType: types.StringType},
			},
		}, policyModels)
	} else {
		data.Policies = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type": types.StringType,
				"data": types.MapType{ElemType: types.StringType},
			},
		})
	}

	// Handle keywords
	if len(app.Keywords) > 0 {
		data.Keywords, _ = types.SetValueFrom(ctx, types.StringType, app.Keywords)
	} else {
		data.Keywords = types.SetNull(types.StringType)
	}

	// Handle related URLs
	if len(app.RelatedURLs) > 0 {
		data.RelatedURLs, _ = types.SetValueFrom(ctx, types.StringType, app.RelatedURLs)
	} else {
		data.RelatedURLs = types.SetNull(types.StringType)
	}

	// Handle destination
	if len(app.Destination) > 0 {
		destinationModels := make([]DestinationModel, len(app.Destination))
		for i, dest := range app.Destination {
			destinationModels[i] = DestinationModel{
				Destination: types.StringValue(dest.Destination),
				Port:        types.StringValue(dest.Port),
				Protocol:    types.StringValue(dest.Protocol),
				Subtype:     types.StringValue(dest.Subtype),
			}
		}
		data.Destination, _ = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"destination": types.StringType,
				"port":        types.StringType,
				"protocol":    types.StringType,
				"subtype":     types.StringType,
			},
		}, destinationModels)
	} else {
		data.Destination = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"destination": types.StringType,
				"port":        types.StringType,
				"protocol":    types.StringType,
				"subtype":     types.StringType,
			},
		})
	}

	// Handle SSO map
	if len(app.SSO) > 0 {
		elements := make(map[string]attr.Value)
		elementTypes := make(map[string]attr.Type)

		for key, value := range app.SSO {
			// Skip the "customer" field - it's a backend-added field that shouldn't be in state
			if key == "customer" {
				continue
			}
			
			// Skip null values - they should not appear in the Terraform configuration
			if value == nil {
				continue
			}

			// Convert each value to appropriate Terraform types
			switch v := value.(type) {
			case string:
				elements[key] = types.StringValue(v)
				elementTypes[key] = types.StringType
			case bool:
				elements[key] = types.BoolValue(v)
				elementTypes[key] = types.BoolType
			case int:
				elements[key] = types.Int64Value(int64(v))
				elementTypes[key] = types.Int64Type
			case int64:
				elements[key] = types.Int64Value(v)
				elementTypes[key] = types.Int64Type
			case float64:
				elements[key] = types.NumberValue(big.NewFloat(v))
				elementTypes[key] = types.NumberType
			case []interface{}:
				// Handle arrays by converting them to tuple of objects
				// Using Tuple instead of List to match HCL's parsing behavior for heterogeneous collections
				tupleElements := make([]attr.Value, len(v))
				tupleTypes := make([]attr.Type, len(v))
				for i, elem := range v {
					if objMap, ok := elem.(map[string]interface{}); ok {
						// Convert map[string]interface{} to object type
						objElements := make(map[string]attr.Value)
						objElementTypes := make(map[string]attr.Type)
						for objKey, objVal := range objMap {
							objElements[objKey] = types.StringValue(fmt.Sprintf("%v", objVal))
							objElementTypes[objKey] = types.StringType
						}
						objValue, _ := types.ObjectValue(objElementTypes, objElements)
						tupleElements[i] = objValue
						tupleTypes[i] = objValue.Type(ctx)
					} else {
						// Fall back to string representation
						tupleElements[i] = types.StringValue(fmt.Sprintf("%v", elem))
						tupleTypes[i] = types.StringType
					}
				}
				// Create a tuple that can hold objects with different schemas
				// This matches how Terraform parses array literals in HCL
				tupleValue, _ := types.TupleValue(tupleTypes, tupleElements)
				elements[key] = tupleValue
				elementTypes[key] = tupleValue.Type(ctx)
			default:
				// For any other type, convert to string but skip if it's a string representation of null
				strValue := fmt.Sprintf("%v", v)
				if strValue == "<nil>" || strValue == "null" || strValue == "None" {
					continue // Skip null-like values
				}
				elements[key] = types.StringValue(strValue)
				elementTypes[key] = types.StringType
			}
		}

		// Create an object from the elements and wrap it in a custom SSO value
		objectValue, _ := types.ObjectValue(elementTypes, elements)
		data.SSO = SSOTypeValue(types.DynamicValue(objectValue))
	} else {
		data.SSO = SSOTypeNull()
	}

	// Handle CustomProperties map
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
		data.CustomProperties, _ = types.MapValue(types.StringType, elements)
	} else {
		data.CustomProperties = types.MapNull(types.StringType)
	}

	// Handle CustomerDomainFields map
	if len(app.CustomerDomainFields) > 0 {
		elements := make(map[string]attr.Value)
		for key, value := range app.CustomerDomainFields {
			elements[key] = types.StringValue(fmt.Sprintf("%v", value))
		}
		data.CustomerDomainFields, _ = types.MapValue(types.StringType, elements)
	} else {
		data.CustomerDomainFields = types.MapNull(types.StringType)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
