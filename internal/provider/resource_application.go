package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
	Icon                 IconValue    `tfsdk:"icon"`
	IconURL              types.String `tfsdk:"icon_url"`
	RelatedURLs          types.Set    `tfsdk:"related_urls"`
	Keywords             types.Set    `tfsdk:"keywords"`
	Locations            types.List   `tfsdk:"locations"`
	Policies             types.List   `tfsdk:"policies"`
	Destination          types.List   `tfsdk:"destination"`
	CustomProperties     types.Map    `tfsdk:"custom_properties"`
	CustomerDomainFields types.Map    `tfsdk:"customer_domain_fields"`
	SSO                  types.Object `tfsdk:"sso"`
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
				Computed:            true,
				// The backend enforces a minimum length and retains the prior value when
				// the field is omitted, so a set description cannot be cleared. Keep the
				// prior value when omitted instead of drifting to "", and reject an
				// explicit "" at plan time (the backend rejects it with "too short").
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Application URL (required for non-ZTNA applications)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Category of the application",
				Optional:            true,
				Computed:            true,
				// Same backend constraint as description: minimum length, retained when
				// omitted, so it cannot be cleared once set. Reject an explicit "" at
				// plan time rather than letting it fail at apply.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide the application",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"agentless_access": schema.BoolAttribute{
				MarkdownDescription: "Enable agentless access",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"mobile_security": schema.BoolAttribute{
				MarkdownDescription: "Enable mobile security",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sbs_only_launch": schema.BoolAttribute{
				MarkdownDescription: "Enable SBS only launch",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"using_template": schema.BoolAttribute{
				MarkdownDescription: "Whether the application was provisioned from a built-in SPA catalog template (backend) — unrelated to the provider's `templates/` directory. The SPA Console owns and normalizes this flag on save, so it can be omitted; the Console-assigned value is adopted into state without producing drift.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: "Name of the template to use from the SPA service's built-in application catalog (backend). Applies only when `using_template` is `true`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"icon": schema.StringAttribute{
				CustomType:          IconType{},
				MarkdownDescription: "Base64 encoded icon data. A `data:<mime>;base64,` prefix and any surrounding whitespace are ignored, so wrapped or data-URI base64 is accepted and does not cause a spurious diff.",
				Optional:            true,
			},
			"icon_url": schema.StringAttribute{
				MarkdownDescription: "Application icon URL",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_domain_fields": schema.MapAttribute{
				MarkdownDescription: "Customer domain fields",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"sso": schema.SingleNestedAttribute{
				MarkdownDescription: "SSO configuration. Set `type` to one of: `saml`, `kerberos`, `basic`, `form`, `nosso`. Only include fields relevant to the chosen SSO type.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					// Universal
					"type": schema.StringAttribute{
						MarkdownDescription: "SSO type: `saml`, `kerberos`, `basic`, `form`, or `nosso`",
						Required:            true,
					},
					// SAML user-provided fields
					"saml_type": schema.StringAttribute{
						MarkdownDescription: "SAML role: `SP`, `IDP`, or `SP_IDP`",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sp_initiated_only": schema.BoolAttribute{
						MarkdownDescription: "Whether SSO is SP-initiated only",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"assertion_url": schema.StringAttribute{
						MarkdownDescription: "SAML assertion consumer service (ACS) URL",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"audience": schema.StringAttribute{
						MarkdownDescription: "SAML audience (entity ID of the service provider)",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"relay_state": schema.StringAttribute{
						MarkdownDescription: "SAML relay state URL",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sign_assertion": schema.StringAttribute{
						MarkdownDescription: "SAML signature scope: `ASSERTION`, `BOTH`, `NONE`, or `RESPONSE`",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name_id_source": schema.StringAttribute{
						MarkdownDescription: "SAML NameID source: `email`, `upn`, `name`, `guid_b64`, or `sam`",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name_id_format": schema.StringAttribute{
						MarkdownDescription: "SAML NameID format: `unspecified`, `emailAddress`, `persistent`, `transient`, `WindowsDomainQualifiedName`, or `X509SubjectName`",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"custom_attributes": schema.ListNestedAttribute{
						MarkdownDescription: "SAML custom attributes (max 16)",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"format": schema.StringAttribute{
									MarkdownDescription: "Attribute format: `uri`, `unspecified`, or `basic`",
									Optional:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Attribute name",
									Required:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Attribute value",
									Required:            true,
								},
								"prefix_expr": schema.BoolAttribute{
									MarkdownDescription: "Whether to use prefix expression",
									Optional:            true,
								},
							},
						},
					},
					// SAML server-computed fields (read-only)
					"saml_sso_login_url": schema.StringAttribute{
						MarkdownDescription: "SAML SSO login URL (computed by server)",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"saml_cert_issuer_name": schema.StringAttribute{
						MarkdownDescription: "SAML certificate issuer name (computed by server)",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					// Server-computed metadata
					"customer": schema.StringAttribute{
						MarkdownDescription: "Customer ID associated with the SSO configuration (computed by server)",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					// Form SSO fields
					"action_url": schema.StringAttribute{
						MarkdownDescription: "Form SSO action URL",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"logonform_url": schema.StringAttribute{
						MarkdownDescription: "Form SSO logon form URL",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"username_field": schema.StringAttribute{
						MarkdownDescription: "Form SSO username HTML field name",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"password_field": schema.StringAttribute{
						MarkdownDescription: "Form SSO password HTML field name",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"attribute": schema.StringAttribute{
						MarkdownDescription: "Form SSO attribute (e.g., `email`, `upn`, `name`)",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					// Shared fields (Form, Kerberos, Basic)
					"username_format": schema.StringAttribute{
						MarkdownDescription: "Username format (used by Form, Kerberos, and Basic SSO)",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					// Kerberos fields
					"user_realm": schema.StringAttribute{
						MarkdownDescription: "Kerberos user realm",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Application state",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("incomplete", "complete"),
				},
			},
			"policy_count": schema.StringAttribute{
				MarkdownDescription: "Number of policies associated with the application",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

// tryRecoverOrphanedApplication searches for an application that may have been partially created
// by the backend despite returning a 500 error. It uses applicationMatchesPlanned to identify
// the correct resource via a best-effort heuristic (see that function's comment for details).
func (r *ApplicationResource) tryRecoverOrphanedApplication(ctx context.Context, app *Application, usingTemplateExplicit bool) *ApplicationListItem {
	apps, err := r.client.GetApplications(ctx, 0, -1, app.Name, app.Type)
	if err != nil || apps == nil || len(apps.Applications) == 0 {
		return nil
	}

	for _, candidate := range apps.Applications {
		if r.applicationMatchesPlanned(ctx, &candidate, app, usingTemplateExplicit) {
			return &candidate
		}
	}
	return nil
}

// applicationMatchesPlanned uses a best-effort heuristic to decide whether a candidate
// ApplicationListItem (returned by the API) matches the Application that was sent in the
// create request. Name and Type must match exactly. For all other fields, a comparison is
// only performed when the candidate carries a non-zero value; if the candidate's field is
// the zero value (empty string, false, nil slice) the check is skipped, because partial
// creation may have left those fields unpersisted. This means the function can return true
// for a candidate that is missing some fields — it is intentionally permissive to avoid
// false negatives during orphan recovery.
func (r *ApplicationResource) applicationMatchesPlanned(ctx context.Context, candidate *ApplicationListItem, planned *Application, usingTemplateExplicit bool) bool {
	// Name and Type must always match exactly — these are the primary identifiers
	// used as query filters in tryRecoverOrphanedApplication.
	if candidate.Name != planned.Name || candidate.Type != planned.Type {
		return false
	}

	// For all remaining fields, only reject if the candidate has a non-zero value
	// that differs from the planned value. A zero value on the candidate (empty string,
	// false, nil slice) means the field may not have been persisted during partial
	// creation, so we skip that comparison rather than treating it as a mismatch.

	// String fields — empty string on candidate means potentially not persisted
	if candidate.Description != "" && candidate.Description != planned.Description {
		return false
	}
	if candidate.URL != "" && candidate.URL != planned.URL {
		return false
	}
	if candidate.Category != "" && candidate.Category != planned.Category {
		return false
	}
	if candidate.TemplateName != "" && candidate.TemplateName != planned.TemplateName {
		return false
	}
	if candidate.Icon != "" && candidate.Icon != planned.Icon {
		return false
	}

	// Boolean fields — false is the zero value, so we can only reject when the
	// candidate is true but planned is false (a definite mismatch). When the
	// candidate is false, we can't distinguish "not set" from "actually false".
	if candidate.Hidden && !planned.Hidden {
		return false
	}
	if candidate.AgentlessAccess && !planned.AgentlessAccess {
		return false
	}
	if candidate.MobileSecurity && !planned.MobileSecurity {
		return false
	}
	if candidate.SbsOnlyLaunch && !planned.SbsOnlyLaunch {
		return false
	}
	// using_template is Console-owned (Optional+Computed): when the user omitted it the
	// Console may set it true even though planned=false, so an omitted value is not a
	// reliable discriminator and is skipped. When the user set it explicitly, honor a
	// definite mismatch (candidate true, planned false) like the other boolean fields so
	// a contradicting candidate is not wrongly adopted.
	if usingTemplateExplicit && candidate.UsingTemplate && !planned.UsingTemplate {
		return false
	}

	// Slice fields — nil/empty on candidate means potentially not persisted
	if len(candidate.RelatedURLs) > 0 && !stringSlicesEqualSorted(candidate.RelatedURLs, planned.RelatedURLs) {
		return false
	}
	if len(candidate.Keywords) > 0 && !stringSlicesEqualSorted(candidate.Keywords, planned.Keywords) {
		return false
	}

	// Locations — only compare if candidate has locations (order-independent)
	if len(candidate.Locations) > 0 {
		if len(candidate.Locations) != len(planned.Locations) {
			return false
		}
		locSet := make(map[string]struct{}, len(planned.Locations))
		for _, loc := range planned.Locations {
			locSet[loc.UUID+"\x00"+loc.Name] = struct{}{}
		}
		for _, loc := range candidate.Locations {
			if _, ok := locSet[loc.UUID+"\x00"+loc.Name]; !ok {
				return false
			}
		}
	}

	// Destinations — only compare if candidate has destinations (order-independent)
	if len(candidate.Destination) > 0 {
		if len(candidate.Destination) != len(planned.Destination) {
			return false
		}
		destSet := make(map[string]struct{}, len(planned.Destination))
		for _, d := range planned.Destination {
			destSet[d.Destination+"\x00"+d.Port+"\x00"+d.Protocol+"\x00"+d.Subtype] = struct{}{}
		}
		for _, d := range candidate.Destination {
			if _, ok := destSet[d.Destination+"\x00"+d.Port+"\x00"+d.Protocol+"\x00"+d.Subtype]; !ok {
				return false
			}
		}
	}

	tflog.Debug(ctx, "spa-terraform-provider: Found matching orphaned application", map[string]any{
		"candidate_id":   candidate.ID,
		"candidate_name": candidate.Name,
	})
	return true
}

// stringSlicesEqualSorted compares two string slices for equality regardless of order.
func stringSlicesEqualSorted(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)
	sort.Strings(aCopy)
	sort.Strings(bCopy)
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
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
	}

	app.TemplateName = data.TemplateName.ValueString()
	app.Icon = normalizeIcon(data.Icon.ValueString())

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
	//
	// sentSSO records the SSO payload we send on create (nil when the config
	// leaves sso unset). It is used after the post-create read to detect and
	// repair the case where the backend does not echo the configured SSO back
	// immediately (see the read-back restore below).
	var sentSSO map[string]any
	if !data.SSO.IsNull() && !data.SSO.IsUnknown() {
		ssoModel, ssoDiags := ssoObjectToModel(ctx, data.SSO)
		resp.Diagnostics.Append(ssoDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ssoMap, ssoDiags := ssoToAPI(ctx, ssoModel)
		resp.Diagnostics.Append(ssoDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.SSO = ssoMap
		sentSSO = ssoMap
	} else {
		app.SSO = nil
	}

	// Include state in the creation request if specified
	if !data.State.IsNull() && !data.State.IsUnknown() {
		app.State = data.State.ValueString()
	}

	// Create the application
	createdApp, err := r.client.CreateApplication(ctx, app)
	if err != nil {
		// On 500 errors, the backend may have partially created the application.
		// Search by name+type to recover the orphan and save its ID to state,
		// preventing duplicate incomplete applications on the next apply.
		//
		// Classify on the parsed status, never on the error text: the message
		// embeds a verbatim echo of the payload we just sent, so a validation
		// error about a URL like "https://portal.example.com/status 500" would
		// otherwise start an orphan hunt and adopt whatever it found.
		if IsInternalServerError(err) {
			tflog.Warn(ctx, "spa-terraform-provider: Create returned 500, attempting to recover orphaned application", map[string]any{
				"app_name": app.Name,
				"app_type": app.Type,
			})
			recovered := r.tryRecoverOrphanedApplication(ctx, app, !data.UsingTemplate.IsNull() && !data.UsingTemplate.IsUnknown())
			if recovered != nil {
				data.ID = types.StringValue(recovered.ID)

				// Computed fields from the plan are still "unknown" at this point.
				// Terraform requires all values to be known after apply, so set
				// any remaining unknown computed fields to null before saving state.
				if data.IconURL.IsUnknown() {
					data.IconURL = types.StringNull()
				}
				if data.PolicyCount.IsUnknown() {
					data.PolicyCount = types.StringNull()
				}
				if data.State.IsUnknown() {
					data.State = types.StringNull()
				}
				if data.Policies.IsUnknown() {
					data.Policies = types.ListNull(types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"type": types.StringType,
							"data": types.MapType{ElemType: types.StringType},
						},
					})
				}
				if data.CustomProperties.IsUnknown() {
					data.CustomProperties = types.MapNull(types.StringType)
				}
				if data.CustomerDomainFields.IsUnknown() {
					data.CustomerDomainFields = types.MapNull(types.StringType)
				}
				if data.SSO.IsUnknown() {
					data.SSO = types.ObjectNull(ssoAttrTypes)
				}
				if data.UsingTemplate.IsUnknown() {
					data.UsingTemplate = types.BoolValue(recovered.UsingTemplate)
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				resp.Diagnostics.AddWarning(
					"Partial Creation Detected",
					fmt.Sprintf(
						"Application '%s' was partially created by the backend (ID: %s) despite returning a 500 error. "+
							"The resource has been saved to state to prevent duplicates. "+
							"Run 'terraform apply' again to update the application to the desired configuration. "+
							"Original error: %s",
						app.Name, recovered.ID, err),
				)
				return
			}
		}
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

	// Restore the configured SSO if the post-create read did not echo it back.
	//
	// Some application types (e.g. a web app whose SSO defaults to
	// {type:"nosso"}) do not return the SSO object on the GET immediately after
	// creation, so the read above sets sso to null. When the configuration
	// specified an SSO, returning null violates Terraform's plan==result
	// contract ("Provider produced inconsistent result after apply"). Restore
	// the SSO we sent so state matches the plan; a later refresh reconciles any
	// backend-computed sub-fields.
	r.restoreSentSSO(ctx, &resp.State, sentSSO, data.SSO, &resp.Diagnostics)
}

// restoreSentSSO re-applies the SSO payload sent to the backend when the
// post-operation read did not echo it back. Some application types omit the SSO
// object from the GET that immediately follows a create/update, so the read
// stores a null SSO; returning null when the configuration set an SSO violates
// Terraform's plan==result contract. Shared by Create and Update so both write
// paths stay in sync. plannedSSO carries the values Terraform planned for the
// resource, used to keep server-computed fields consistent (see
// preserveComputedSSOFields).
func (r *ApplicationResource) restoreSentSSO(ctx context.Context, state *tfsdk.State, sentSSO map[string]any, plannedSSO types.Object, diags *diag.Diagnostics) {
	if sentSSO == nil || diags.HasError() {
		return
	}
	var stateData ApplicationResourceModel
	diags.Append(state.Get(ctx, &stateData)...)
	if diags.HasError() || !stateData.SSO.IsNull() {
		return
	}
	ssoModel, ssoDiags := ssoFromAPI(ctx, sentSSO)
	diags.Append(ssoDiags...)
	if diags.HasError() {
		return
	}
	// ssoToAPI strips the server-computed fields from sentSSO, so restoring
	// solely from it would null them out. On update the plan keeps their prior
	// values known via UseStateForUnknown; preserve those to satisfy the
	// plan==result contract. On create the planned values are unknown and stay
	// null.
	if !plannedSSO.IsNull() && !plannedSSO.IsUnknown() {
		plannedModel, planDiags := ssoObjectToModel(ctx, plannedSSO)
		diags.Append(planDiags...)
		if diags.HasError() {
			return
		}
		preserveComputedSSOFields(ssoModel, plannedModel)
	}
	ssoObj, objDiags := ssoModelToObject(ctx, ssoModel)
	diags.Append(objDiags...)
	if diags.HasError() {
		return
	}
	stateData.SSO = ssoObj
	diags.Append(state.Set(ctx, &stateData)...)
}

// preserveComputedSSOFields copies the server-computed SSO fields that ssoToAPI
// omits from the write payload (saml_sso_login_url, saml_cert_issuer_name,
// customer) from planned into reconstructed whenever the planned value is known.
// This keeps the restored state aligned with a plan that carried those values
// forward via UseStateForUnknown (the update path); on create they are unknown
// and left untouched.
func preserveComputedSSOFields(reconstructed, planned *SSOModel) {
	if reconstructed == nil || planned == nil {
		return
	}
	keep := func(dst *types.String, src types.String) {
		if !src.IsNull() && !src.IsUnknown() {
			*dst = src
		}
	}
	keep(&reconstructed.SamlSSOLoginURL, planned.SamlSSOLoginURL)
	keep(&reconstructed.SamlCertIssuerName, planned.SamlCertIssuerName)
	keep(&reconstructed.Customer, planned.Customer)
}

// shouldAwaitSSO reports whether Read should use the bounded SSO re-fetch
// (GetApplicationAwaitSSO) instead of a single GET. It returns true on import —
// where ImportState passes through only the ID, so the prior name is null — and
// when the prior state already had an SSO object, so a transient empty read in
// the post-write consistency window is smoothed over instead of recorded as
// drift. A settled application whose prior state has no SSO returns false and is
// read with a single GET, avoiding extra GETs on every routine refresh.
func shouldAwaitSSO(priorName types.String, priorSSO types.Object) bool {
	return priorName.IsNull() || !priorSSO.IsNull()
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: ApplicationResource.Read - Reading application")
	var data ApplicationResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the application from the API. shouldAwaitSSO decides whether to use the
	// bounded SSO re-fetch (import, or when prior state already had an SSO object)
	// or a single GET (a settled application with no SSO), so routine refreshes of
	// SSO-less applications don't pay extra GETs.
	var app *Application
	var err error
	if shouldAwaitSSO(data.Name, data.SSO) {
		app, err = r.client.GetApplicationAwaitSSO(ctx, data.ID.ValueString())
	} else {
		app, err = r.client.GetApplication(ctx, data.ID.ValueString())
	}
	if err != nil {
		if IsNotFound(err) {
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

	// The backend never stores an empty description/category (min length 1), so an
	// empty API value means "unset" and reads back as null; a non-empty value reads
	// back as-is. UseStateForUnknown then keeps an omitted value at its prior state,
	// and the LengthAtLeast(1) validator rejects an explicit "". url is required and
	// always a known string.
	if app.Description != "" {
		data.Description = types.StringValue(app.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.URL = types.StringValue(app.URL)
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
		data.Icon = NewIconValue(app.Icon)
	} else {
		data.Icon = NewIconNull()
	}

	if app.IconURL != "" {
		data.IconURL = types.StringValue(app.IconURL)
	} else {
		data.IconURL = types.StringNull()
	}

	// Handle SSO fields
	if len(app.SSO) > 0 {
		ssoModel, ssoDiags := ssoFromAPI(ctx, app.SSO)
		resp.Diagnostics.Append(ssoDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ssoObj, objDiags := ssoModelToObject(ctx, ssoModel)
		resp.Diagnostics.Append(objDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.SSO = ssoObj
	} else {
		data.SSO = types.ObjectNull(ssoAttrTypes)
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
	app.Icon = normalizeIcon(data.Icon.ValueString())
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
	var sentSSO map[string]any
	if !data.SSO.IsNull() && !data.SSO.IsUnknown() {
		ssoModel, ssoDiags := ssoObjectToModel(ctx, data.SSO)
		resp.Diagnostics.Append(ssoDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ssoMap, ssoDiags := ssoToAPI(ctx, ssoModel)
		resp.Diagnostics.Append(ssoDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		app.SSO = ssoMap
		sentSSO = ssoMap
	} else {
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

	// Restore the configured SSO if the post-update read did not echo it back
	// (same plan==result contract as Create; see restoreSentSSO).
	r.restoreSentSSO(ctx, &resp.State, sentSSO, data.SSO, &resp.Diagnostics)
}

// urlHostname reduces a configured URL to the bare hostname the backend stores
// as a routing domain FQDN, or "" when none can be derived. It mirrors the
// backend's _extract_url_hostname (spaconfig, gatewaycore/adminapi/routing/
// graph_handlers.py): parse, then retry scheme-less input as an authority so
// that "host:port" yields "host" rather than being read as scheme plus opaque.
func urlHostname(raw string) string {
	if parsed, err := url.Parse(raw); err == nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	if !strings.Contains(raw, "://") {
		if parsed, err := url.Parse("//" + raw); err == nil && parsed.Hostname() != "" {
			return parsed.Hostname()
		}
	}
	return ""
}

// urlAuthority reduces a value whose hostname could not be parsed to its
// authority, dropping the scheme prefix and everything from the first "/", "?"
// or "#". The service stores any url that contains no bare "%" and no double
// quote, so a value Go cannot parse can still reach state; without this the
// delete warning would echo that value's path and query — which may carry a
// token — into customer-visible and CI output.
func urlAuthority(raw string) string {
	if scheme := strings.Index(raw, "://"); scheme >= 0 {
		raw = raw[scheme+len("://"):]
	}
	if cut := strings.IndexAny(raw, "/?#"); cut >= 0 {
		raw = raw[:cut]
	}
	return strings.TrimSpace(raw)
}

// applicationFQDNs returns the FQDNs an application references, in the form the
// backend stores routing domains. Never returns an error: a value whose hostname
// cannot be derived is reduced to its authority rather than dropped, so a delete
// is never blocked by a formatting problem in state.
func applicationFQDNs(ctx context.Context, data *ApplicationResourceModel) []string {
	fqdns := []string{}
	seen := make(map[string]bool)
	add := func(value string) {
		value = strings.TrimSpace(value)
		// An FQDN never contains "@", so drop any userinfo prefix rather than
		// echo embedded credentials back in a customer-visible warning.
		if at := strings.LastIndex(value, "@"); at >= 0 {
			value = value[at+1:]
		}
		// The backend lower-cases every FQDN it stores, so match it.
		value = strings.ToLower(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		fqdns = append(fqdns, value)
	}
	// addHost applies the backend's hostname reduction, falling back to the
	// value's authority when it yields nothing. Not used for ZTNA destinations,
	// which the backend stores verbatim — parsing those would truncate a CIDR
	// suffix.
	addHost := func(value string) {
		value = strings.TrimSpace(value)
		if host := urlHostname(value); host != "" {
			add(host)
			return
		}
		add(urlAuthority(value))
	}

	if !data.URL.IsNull() && !data.URL.IsUnknown() {
		if raw := strings.TrimSpace(data.URL.ValueString()); raw != "" {
			addHost(raw)
		}
	}

	if !data.RelatedURLs.IsNull() && !data.RelatedURLs.IsUnknown() {
		var relatedURLs []string
		if diags := data.RelatedURLs.ElementsAs(ctx, &relatedURLs, false); !diags.HasError() {
			for _, related := range relatedURLs {
				addHost(related)
			}
		}
	}

	if !data.Destination.IsNull() && !data.Destination.IsUnknown() {
		var destinations []DestinationModel
		if diags := data.Destination.ElementsAs(ctx, &destinations, false); !diags.HasError() {
			for _, dest := range destinations {
				add(dest.Destination.ValueString())
			}
		}
	}

	return fqdns
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

	// "may": the provider cannot see the tenant's other live applications, which
	// is what decides whether each FQDN's routing domain is retained.
	if fqdns := applicationFQDNs(ctx, &data); len(fqdns) > 0 {
		resp.Diagnostics.AddWarning(
			"Routing Domains May Also Be Deleted",
			fmt.Sprintf("Deleting this application may also delete the routing domains for: %s.\n\n"+
				"Citrix Secure Private Access deletes a routing domain whose FQDN matches one of this application's url, related_urls or destination values, "+
				"unless another application still references the same FQDN. See Note 3 in the citrixspa_routing_domain documentation for the full rule.",
				strings.Join(fqdns, ", ")),
		)
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
