package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

// ApplicationResource defines the resource implementation.
type ApplicationResource struct {
	client SPAClient
}

// ApplicationResourceModel describes the resource data model.
type ApplicationResourceModel struct {
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

// DestinationModel represents a destination in the resource model
type DestinationModel struct {
	Destination types.String `tfsdk:"destination"`
	Port        types.String `tfsdk:"port"`
	Protocol    types.String `tfsdk:"protocol"`
	Subtype     types.String `tfsdk:"subtype"`
}

// LocationModel represents a location in the resource model
type LocationModel struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// PolicyModel represents a policy in the resource model
type PolicyModel struct {
	Type types.String `tfsdk:"type"`
	Data types.Map    `tfsdk:"data"`
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Metadata - Setting resource metadata")
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Schema - Defining resource schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SPA application.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Application identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the application",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of application (web, saas, ztna)",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the application",
				Optional:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Application URL (required for non-ZTNA applications)",
				Optional:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Category of the application",
				Optional:            true,
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide the application",
				Optional:            true,
			},
			"agentless_access": schema.BoolAttribute{
				MarkdownDescription: "Enable agentless access",
				Optional:            true,
			},
			"mobile_security": schema.BoolAttribute{
				MarkdownDescription: "Enable mobile security",
				Optional:            true,
			},
			"sbs_only_launch": schema.BoolAttribute{
				MarkdownDescription: "Enable SBS only launch",
				Optional:            true,
			},
			"using_template": schema.BoolAttribute{
				MarkdownDescription: "Whether using a template (required for non-ZTNA applications)",
				Optional:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: "Template name",
				Optional:            true,
			},
			"icon": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded icon data",
				Optional:            true,
			},
			"icon_url": schema.StringAttribute{
				MarkdownDescription: "Application icon URL",
				Computed:            true,
			},
			"related_urls": schema.SetAttribute{
				MarkdownDescription: "Related URLs (required for non-ZTNA applications)",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             setdefault.StaticValue(types.SetNull(types.StringType)),
			},
			"keywords": schema.SetAttribute{
				MarkdownDescription: "Keywords associated with the application",
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetNull(types.StringType)),
				ElementType:         types.StringType,
			},
			"locations": schema.ListNestedAttribute{
				MarkdownDescription: "Locations associated with the application",
				Optional:            true,
				Computed:            true,
				Default: listdefault.StaticValue(types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name": types.StringType,
						"uuid": types.StringType,
					},
				})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Location name",
							Required:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Location UUID",
							Required:            true,
						},
					},
				},
			},
			"policies": schema.ListNestedAttribute{
				MarkdownDescription: "Policies associated with the application",
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
				Default: listdefault.StaticValue(types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"destination": types.StringType,
						"port":        types.StringType,
						"protocol":    types.StringType,
						"subtype":     types.StringType,
					},
				})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"destination": schema.StringAttribute{
							MarkdownDescription: "Destination (hostname or IP range)",
							Optional:            true,
						},
						"port": schema.StringAttribute{
							MarkdownDescription: "Port number",
							Optional:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol (PROTOCOL_TCP, PROTOCOL_UDP)",
							Optional:            true,
						},
						"subtype": schema.StringAttribute{
							MarkdownDescription: "Subtype (SUBTYPE_HOSTNAME, SUBTYPE_IP_AND_CIDR, SUBTYPE_IP_RANGE)",
							Optional:            true,
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
				Validators: []validator.String{
					stringvalidator.OneOf("incomplete", "complete"),
				},
			},
			"policy_count": schema.StringAttribute{
				MarkdownDescription: "Number of policies associated with the application",
				Computed:            true,
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Configure - Configuring resource client")
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

// completeApplication makes a PUT request to complete an application
func (r *ApplicationResource) completeApplication(ctx context.Context, applicationID string) error {
	tflog.Debug(ctx, "spa-terraform-provider: Completing application", map[string]any{
		"application_id": applicationID,
	})

	err := r.client.CompleteApplication(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("failed to complete application: %w", err)
	}

	tflog.Debug(ctx, "spa-terraform-provider: Application completed successfully", map[string]any{
		"application_id": applicationID,
	})
	return nil
}

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Create - Creating application")
	var data ApplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert Terraform model to API model
	app := &Application{
		Name:        data.Name.ValueString(),
		Type:        data.Type.ValueString(),
		Description: data.Description.ValueString(),
		URL:         data.URL.ValueString(),
		Category:    data.Category.ValueString(),
	}

	if !data.Hidden.IsNull() {
		app.Hidden = data.Hidden.ValueBool()
	}
	if !data.AgentlessAccess.IsNull() {
		app.AgentlessAccess = data.AgentlessAccess.ValueBool()
	}
	if !data.MobileSecurity.IsNull() {
		app.MobileSecurity = data.MobileSecurity.ValueBool()
	}
	if !data.SbsOnlyLaunch.IsNull() {
		app.SbsOnlyLaunch = data.SbsOnlyLaunch.ValueBool()
	}
	if !data.UsingTemplate.IsNull() {
		app.UsingTemplate = data.UsingTemplate.ValueBool()
	} else {
		app.UsingTemplate = false // Default to false if not set
	}

	app.TemplateName = data.TemplateName.ValueString()
	app.Icon = data.Icon.ValueString()

	// Handle related URLs
	if !data.RelatedURLs.IsNull() {
		var relatedURLs []string
		resp.Diagnostics.Append(data.RelatedURLs.ElementsAs(ctx, &relatedURLs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.RelatedURLs = relatedURLs
	} else {
		// Set to empty list if no related URLs
		app.RelatedURLs = []string{}
	}

	// Handle keywords
	if !data.Keywords.IsNull() {
		var keywords []string
		resp.Diagnostics.Append(data.Keywords.ElementsAs(ctx, &keywords, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.Keywords = keywords
	} else {
		// Set to empty list if no keywords
		app.Keywords = []string{}
	}

	// Handle locations
	if !data.Locations.IsNull() {
		var locations []LocationModel
		resp.Diagnostics.Append(data.Locations.ElementsAs(ctx, &locations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert LocationModel to Location
		apiLocations := make([]Location, len(locations))
		for i, loc := range locations {
			apiLocations[i] = Location{
				Name: loc.Name.ValueString(),
				UUID: loc.UUID.ValueString(),
			}
		}
		app.Locations = apiLocations
	} else {
		// Set to empty list if no locations
		app.Locations = []Location{}
	}

	// Note: policies field is read-only and computed by backend, not sent in create request

	// Handle destinations
	if !data.Destination.IsNull() {
		var destinations []DestinationModel
		resp.Diagnostics.Append(data.Destination.ElementsAs(ctx, &destinations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		app.Destination = make([]Destination, len(destinations))
		for i, dest := range destinations {
			app.Destination[i] = Destination{
				Destination: dest.Destination.ValueString(),
				Port:        dest.Port.ValueString(),
				Protocol:    dest.Protocol.ValueString(),
				Subtype:     dest.Subtype.ValueString(),
			}
		}
	} else {
		// Set to empty list if no destinations
		app.Destination = []Destination{}
	}

	// Handle custom properties fields
	if !data.CustomProperties.IsNull() && !data.CustomProperties.IsUnknown() {
		var elements map[string]string
		resp.Diagnostics.Append(data.CustomProperties.ElementsAs(ctx, &elements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.CustomProperties = make(map[string]any, len(elements))
		for key, value := range elements {
			// Try to parse as JSON for complex objects
			var jsonValue any
			if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
				// Successfully parsed as JSON, use the parsed value
				app.CustomProperties[key] = jsonValue
			} else {
				// Not JSON, use as string
				app.CustomProperties[key] = value
			}
		}
	} else {
		// Set to empty map if no customer domain fields
		app.CustomProperties = nil
	}

	// Handle customer domain fields
	if !data.CustomerDomainFields.IsNull() && !data.CustomerDomainFields.IsUnknown() {
		var elements map[string]string
		resp.Diagnostics.Append(data.CustomerDomainFields.ElementsAs(ctx, &elements, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.CustomerDomainFields = make(map[string]any, len(elements))
		for key, value := range elements {
			app.CustomerDomainFields[key] = value
		}
	} else {
		// Set to empty map if no customer domain fields
		app.CustomerDomainFields = map[string]any{}
	}

	// Handle SSO fields
	if !data.SSO.IsNull() && !data.SSO.IsUnknown() {
		// Extract the underlying value from the Dynamic type
		underlyingValue := data.SSO.UnderlyingValue()

		// The underlying value should be a map-like structure
		if objectValue, ok := underlyingValue.(types.Object); ok {
			// Convert Object attributes to map[string]interface{}
			attributes := objectValue.Attributes()

			app.SSO = make(map[string]interface{}, len(attributes))
			for key, value := range attributes {
				// Convert each element back to a Go value
				if dynVal, ok := value.(types.Dynamic); ok {
					app.SSO[key] = dynVal.UnderlyingValue()
				} else {
					// Handle other attr.Value types
					switch v := value.(type) {
					case types.String:
						app.SSO[key] = v.ValueString()
					case types.Bool:
						app.SSO[key] = v.ValueBool()
					case types.Int64:
						app.SSO[key] = v.ValueInt64()
					case types.Number:
						f, _ := v.ValueBigFloat().Float64()
						app.SSO[key] = f
					case types.List:
						// Handle lists of objects
						listElements := v.Elements()
						goList := make([]interface{}, len(listElements))
						for i, elem := range listElements {
							if objElem, ok := elem.(types.Object); ok {
								objAttributes := objElem.Attributes()
								goMap := make(map[string]interface{})
								for objKey, objVal := range objAttributes {
									if strVal, ok := objVal.(types.String); ok {
										goMap[objKey] = strVal.ValueString()
									} else {
										goMap[objKey] = fmt.Sprintf("%v", objVal)
									}
								}
								goList[i] = goMap
							} else {
								goList[i] = fmt.Sprintf("%v", elem)
							}
						}
						app.SSO[key] = goList
					default:
						app.SSO[key] = fmt.Sprintf("%v", value)
					}
				}
			}
		} else {
			// Fallback: treat as a single value or other structure
			app.SSO = map[string]interface{}{"value": underlyingValue}
		}
	} else {
		// Set to empty map if no SSO fields
		app.SSO = nil
	}

	// Include state in the creation request if specified
	if !data.State.IsNull() && !data.State.IsUnknown() {
		app.State = data.State.ValueString()
	}

	// Create the application
	createdApp, err := r.client.CreateApplication(ctx, app)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create application, got error: %s", err))
		return
	}

	// Update the model with the created application data
	data.ID = types.StringValue(createdApp.ID)

	// Set the initial state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the full application data to populate all computed fields
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

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Read - Reading application")
	var data ApplicationResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the application from the API
	app, err := r.client.GetApplication(ctx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}

	// Update the model with the API response
	data.Name = types.StringValue(app.Name)
	data.Type = types.StringValue(app.Type)
	data.State = types.StringValue(app.State)
	data.PolicyCount = types.StringValue(app.PolicyCount)

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

	// Handle SSO fields
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
						objValue, objDiag := types.ObjectValue(objElementTypes, objElements)
						resp.Diagnostics.Append(objDiag...)
						if resp.Diagnostics.HasError() {
							return
						}
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
				tupleValue, tupleDiag := types.TupleValue(tupleTypes, tupleElements)
				resp.Diagnostics.Append(tupleDiag...)
				if resp.Diagnostics.HasError() {
					return
				}
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
		objectValue, objDiag := types.ObjectValue(elementTypes, elements)
		resp.Diagnostics.Append(objDiag...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.SSO = SSOTypeValue(types.DynamicValue(objectValue))
	} else {
		data.SSO = SSOTypeNull()
	}

	// Handle boolean fields (these should always have values)
	data.Hidden = types.BoolValue(app.Hidden)
	data.AgentlessAccess = types.BoolValue(app.AgentlessAccess)
	data.MobileSecurity = types.BoolValue(app.MobileSecurity)
	data.SbsOnlyLaunch = types.BoolValue(app.SbsOnlyLaunch)
	data.UsingTemplate = types.BoolValue(app.UsingTemplate)

	// Handle related URLs
	if len(app.RelatedURLs) > 0 {
		data.RelatedURLs, _ = types.SetValueFrom(ctx, types.StringType, app.RelatedURLs)
	} else {
		// Set to empty set if no related URLs
		data.RelatedURLs = types.SetNull(types.StringType)
	}

	// Handle keywords
	if len(app.Keywords) > 0 {
		data.Keywords, _ = types.SetValueFrom(ctx, types.StringType, app.Keywords)
	} else {
		// Set to empty set if no keywords
		data.Keywords = types.SetNull(types.StringType)
	}

	// Handle custom properties fields
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

	// Handle customer domain fields
	if len(app.CustomerDomainFields) > 0 {
		elements := make(map[string]attr.Value)
		for key, value := range app.CustomerDomainFields {
			if str, ok := value.(string); ok {
				elements[key] = types.StringValue(str)
			} else {
				// Convert to string if not already
				elements[key] = types.StringValue(fmt.Sprintf("%v", value))
			}
		}
		data.CustomerDomainFields, _ = types.MapValue(types.StringType, elements)
	} else {
		data.CustomerDomainFields = types.MapNull(types.StringType)
	}

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
		// Set to empty list if no locations
		data.Locations = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name": types.StringType,
				"uuid": types.StringType,
			},
		})
	}

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
		// Set to empty list if no policies
		data.Policies = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"type": types.StringType,
				"data": types.MapType{ElemType: types.StringType},
			},
		})
	}

	// Handle destinations
	if len(app.Destination) > 0 {
		// Convert Destination objects to DestinationModel
		destinationModels := make([]DestinationModel, len(app.Destination))
		for i, dest := range app.Destination {
			destinationModels[i] = DestinationModel{
				Destination: types.StringValue(dest.Destination),
				Port:        types.StringValue(dest.Port),
				Protocol:    types.StringValue(dest.Protocol),
				Subtype:     types.StringValue(dest.Subtype),
			}
		}
		destinationType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"destination": types.StringType,
				"port":        types.StringType,
				"protocol":    types.StringType,
				"subtype":     types.StringType,
			},
		}
		data.Destination, _ = types.ListValueFrom(ctx, destinationType, destinationModels)
	} else {
		destinationType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"destination": types.StringType,
				"port":        types.StringType,
				"protocol":    types.StringType,
				"subtype":     types.StringType,
			},
		}
		data.Destination = types.ListNull(destinationType)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Update - Updating application")
	var plan ApplicationResourceModel
	var state ApplicationResourceModel

	// Read Terraform plan and state data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use plan for the rest of the update
	data := plan

	// Convert Terraform model to API model
	app := &Application{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Type:        data.Type.ValueString(),
		Description: data.Description.ValueString(),
		URL:         data.URL.ValueString(),
		Category:    data.Category.ValueString(),
	}

	if !data.Hidden.IsNull() {
		app.Hidden = data.Hidden.ValueBool()
	}
	if !data.AgentlessAccess.IsNull() {
		app.AgentlessAccess = data.AgentlessAccess.ValueBool()
	}
	if !data.MobileSecurity.IsNull() {
		app.MobileSecurity = data.MobileSecurity.ValueBool()
	}
	if !data.SbsOnlyLaunch.IsNull() {
		app.SbsOnlyLaunch = data.SbsOnlyLaunch.ValueBool()
	}
	if !data.UsingTemplate.IsNull() {
		app.UsingTemplate = data.UsingTemplate.ValueBool()
	}

	app.TemplateName = data.TemplateName.ValueString()
	app.Icon = data.Icon.ValueString()
	app.IconURL = data.IconURL.ValueString()

	// Handle related URLs
	if !data.RelatedURLs.IsNull() {
		var relatedURLs []string
		resp.Diagnostics.Append(data.RelatedURLs.ElementsAs(ctx, &relatedURLs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.RelatedURLs = relatedURLs
	}

	// Handle keywords
	if !data.Keywords.IsNull() {
		var keywords []string
		resp.Diagnostics.Append(data.Keywords.ElementsAs(ctx, &keywords, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.Keywords = keywords
	}

	// Handle locations
	if !data.Locations.IsNull() {
		var locations []LocationModel
		resp.Diagnostics.Append(data.Locations.ElementsAs(ctx, &locations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert LocationModel to Location
		apiLocations := make([]Location, len(locations))
		for i, loc := range locations {
			apiLocations[i] = Location{
				Name: loc.Name.ValueString(),
				UUID: loc.UUID.ValueString(),
			}
		}
		app.Locations = apiLocations
	} else {
		// Set to empty list if no locations
		app.Locations = []Location{}
	}

	// Note: policies field is read-only and computed by backend, not sent in update request

	// Handle destinations
	if !data.Destination.IsNull() {
		var destinations []DestinationModel
		resp.Diagnostics.Append(data.Destination.ElementsAs(ctx, &destinations, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		app.Destination = make([]Destination, len(destinations))
		for i, dest := range destinations {
			app.Destination[i] = Destination{
				Destination: dest.Destination.ValueString(),
				Port:        dest.Port.ValueString(),
				Protocol:    dest.Protocol.ValueString(),
				Subtype:     dest.Subtype.ValueString(),
			}
		}
	}

	// Handle custom properties fields
	if !data.CustomProperties.IsNull() && !data.CustomProperties.IsUnknown() {
		var elements map[string]string
		resp.Diagnostics.Append(data.CustomProperties.ElementsAs(ctx, &elements, false)...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "Error reading customer domain fields")
			return
		}
		app.CustomProperties = make(map[string]any, len(elements))
		for key, value := range elements {
			// Try to parse as JSON for complex objects
			var jsonValue any
			if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
				// Successfully parsed as JSON, use the parsed value
				app.CustomProperties[key] = jsonValue
			} else {
				// Not JSON, use as string
				app.CustomProperties[key] = value
			}
		}
	} else {
		// Set to empty map if no customer domain fields
		app.CustomProperties = nil
	}

	// Handle customer domain fields
	if !data.CustomerDomainFields.IsNull() && !data.CustomerDomainFields.IsUnknown() {
		var elements map[string]string
		resp.Diagnostics.Append(data.CustomerDomainFields.ElementsAs(ctx, &elements, false)...)
		if resp.Diagnostics.HasError() {
			tflog.Error(ctx, "Error reading customer domain fields")
			return
		}
		app.CustomerDomainFields = make(map[string]any, len(elements))
		for key, value := range elements {
			app.CustomerDomainFields[key] = value
		}
	} else {
		// Set to empty map if no customer domain fields
		app.CustomerDomainFields = map[string]any{}
	}

	// Handle SSO fields
	if !data.SSO.IsNull() && !data.SSO.IsUnknown() {
		// Extract the underlying value from the Dynamic type
		underlyingValue := data.SSO.UnderlyingValue()

		// The underlying value should be a map-like structure
		if objectValue, ok := underlyingValue.(types.Object); ok {
			// Convert Object attributes to map[string]interface{}
			attributes := objectValue.Attributes()

			app.SSO = make(map[string]interface{}, len(attributes))
			for key, value := range attributes {
				// Convert each element back to a Go value
				if dynVal, ok := value.(types.Dynamic); ok {
					app.SSO[key] = dynVal.UnderlyingValue()
				} else {
					// Handle other attr.Value types
					switch v := value.(type) {
					case types.String:
						app.SSO[key] = v.ValueString()
					case types.Bool:
						app.SSO[key] = v.ValueBool()
					case types.Int64:
						app.SSO[key] = v.ValueInt64()
					case types.Number:
						f, _ := v.ValueBigFloat().Float64()
						app.SSO[key] = f
					case types.List:
						// Handle lists of objects
						listElements := v.Elements()
						goList := make([]interface{}, len(listElements))
						for i, elem := range listElements {
							if objElem, ok := elem.(types.Object); ok {
								objAttributes := objElem.Attributes()
								goMap := make(map[string]interface{})
								for objKey, objVal := range objAttributes {
									if strVal, ok := objVal.(types.String); ok {
										goMap[objKey] = strVal.ValueString()
									} else {
										goMap[objKey] = fmt.Sprintf("%v", objVal)
									}
								}
								goList[i] = goMap
							} else {
								goList[i] = fmt.Sprintf("%v", elem)
							}
						}
						app.SSO[key] = goList
					default:
						app.SSO[key] = fmt.Sprintf("%v", value)
					}
				}
			}
		} else {
			// Fallback: treat as a single value or other structure
			app.SSO = map[string]interface{}{"value": underlyingValue}
		}
	} else {
		// Set to empty map if no SSO fields
		app.SSO = nil
	}

	// Update the application
	err := r.client.UpdateApplication(ctx, data.ID.ValueString(), app)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update application, got error: %s", err))
		return
	}

	// Handle state transitions (incomplete -> complete)
	if !plan.State.IsNull() && !plan.State.IsUnknown() && !state.State.IsNull() && !state.State.IsUnknown() {
		oldState := state.State.ValueString()
		newState := plan.State.ValueString()

		// If user wants to transition from incomplete to complete
		if oldState == "incomplete" && newState == "complete" {
			tflog.Debug(ctx, "spa-terraform-provider: User requested state transition to complete")
			err := r.completeApplication(ctx, data.ID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to complete application",
					fmt.Sprintf("Could not complete application %s: %s", data.ID.ValueString(), err.Error()),
				)
				return
			}
		}
	} else if !plan.State.IsNull() && !plan.State.IsUnknown() && plan.State.ValueString() == "complete" {
		// If state wasn't tracked before but user wants complete now
		tflog.Debug(ctx, "spa-terraform-provider: User requested completion")
		err := r.completeApplication(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to complete application",
				fmt.Sprintf("Could not complete application %s: %s", data.ID.ValueString(), err.Error()),
			)
			return
		}
	}

	// Set the updated state with the ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the updated application to populate all computed fields
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

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Delete - Deleting application")
	var data ApplicationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the application
	err := r.client.DeleteApplication(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete application, got error: %s", err))
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.ImportState - Importing application", map[string]any{
		"application_id": req.ID,
	})
	// Import using the application ID

	// Set only the ID and let Read populate everything else
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
