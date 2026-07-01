package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SessionPolicyResource{}
var _ resource.ResourceWithImportState = &SessionPolicyResource{}

func NewSessionPolicyResource() resource.Resource {
	return &SessionPolicyResource{}
}

// SessionPolicyResource defines the resource implementation.
type SessionPolicyResource struct {
	client SPAClient
}

// SessionPolicyResourceModel describes the resource data model.
type SessionPolicyResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	Name        types.String             `tfsdk:"name"`
	Description types.String             `tfsdk:"description"`
	Active      types.Bool               `tfsdk:"active"`
	Priority    types.Int64              `tfsdk:"priority"`
	Rules       []SessionPolicyRuleModel `tfsdk:"generic_rules"`
}

// SessionPolicyRuleModel describes a single rule within a session policy.
type SessionPolicyRuleModel struct {
	ID          types.String                  `tfsdk:"id"`
	Name        types.String                  `tfsdk:"name"`
	Description types.String                  `tfsdk:"description"`
	Priority    types.Int64                   `tfsdk:"priority"`
	Active      types.Bool                    `tfsdk:"active"`
	Actions     *SessionPolicyActionsModel    `tfsdk:"actions"`
	Conditions  []SessionPolicyConditionModel `tfsdk:"condition"`
}

// SessionPolicyActionsModel describes the actions block within a session policy rule.
type SessionPolicyActionsModel struct {
	Routing               types.String `tfsdk:"routing"`
	DisableSecurityGroups types.String `tfsdk:"disable_security_groups"`
	LocalLanAccess        types.String `tfsdk:"local_lan_access"`
}

// SessionPolicyConditionModel describes a single condition within a session policy rule.
type SessionPolicyConditionModel struct {
	Type      types.String `tfsdk:"type"`
	Operator  types.String `tfsdk:"operator"`
	TagSource types.String `tfsdk:"tag_source"`
	TagKey    types.String `tfsdk:"tag_key"`
	Values    types.List   `tfsdk:"values"`
	Metadata  types.Map    `tfsdk:"metadata"`
}

func (r *SessionPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Metadata - Setting resource metadata")
	resp.TypeName = req.ProviderTypeName + "_session_policy"
}

func (r *SessionPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Schema - Defining resource schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a SPA session policy. Session policies apply routing and security behaviour at the session level across all applications.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Session policy identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the session policy (must be unique per customer)",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the session policy. Defaults to empty string when not set.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the session policy is evaluated at runtime",
				Required:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority of the session policy (must be unique per customer). Omit to let the server auto-assign.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"generic_rules": schema.ListNestedAttribute{
				MarkdownDescription: "List of rules within the session policy (1–100 items)",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Rule identifier (assigned by the server)",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Rule name",
							Optional:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Rule description",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "Rule priority within the policy (lower = higher priority)",
							Required:            true,
						},
						"active": schema.BoolAttribute{
							MarkdownDescription: "Whether this rule is evaluated",
							Required:            true,
						},
						"actions": schema.SingleNestedAttribute{
							MarkdownDescription: "Actions to apply when this rule matches. Not all action fields are available in every tenant configuration — only include fields applicable to your deployment.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"routing": schema.StringAttribute{
									MarkdownDescription: "Routing direction. Valid values: 'default', 'external'.",
									Optional:            true,
								},
								"disable_security_groups": schema.StringAttribute{
									MarkdownDescription: "Whether to disable security groups. Valid values: 'true', 'false'.",
									Optional:            true,
								},
								"local_lan_access": schema.StringAttribute{
									MarkdownDescription: "Local LAN access. Valid values: 'enabled', 'disabled'.",
									Optional:            true,
								},
							},
						},
						"condition": schema.ListNestedAttribute{
							MarkdownDescription: "Conditions that must ALL match for the rule to fire (1–100 items)",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "Condition type (TYPE_USERGROUP, TYPE_PLATFORM, TYPE_TAG, TYPE_MACHINEGROUP, TYPE_MULTIURLDOMAIN)",
										Required:            true,
									},
									"operator": schema.StringAttribute{
										MarkdownDescription: "Condition operator (OPERATOR_EQ, OPERATOR_IN, OPERATOR_CONTAINS, OPERATOR_LTE, OPERATOR_GTE, OPERATOR_NOT, OPERATOR_RANGE)",
										Required:            true,
									},
									"tag_source": schema.StringAttribute{
										MarkdownDescription: "Tag source (NLS, CAS, EPA, ITM, ThirdPartyDevicePosture). Not used for TYPE_MULTIURLDOMAIN.",
										Optional:            true,
										Computed:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"tag_key": schema.StringAttribute{
										MarkdownDescription: "Tag key identifier. Not used for TYPE_MULTIURLDOMAIN.",
										Optional:            true,
										Computed:            true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"values": schema.ListAttribute{
										MarkdownDescription: "List of values for the condition (1–100 items)",
										Required:            true,
										ElementType:         types.StringType,
									},
									"metadata": schema.MapAttribute{
										MarkdownDescription: "Optional metadata attached to the condition (e.g., display labels for user/group entries)",
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
										PlanModifiers: []planmodifier.Map{
											mapplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *SessionPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Configure - Configuring resource client")
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

func (r *SessionPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Create - Creating session policy")
	var data SessionPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := sessionPolicyFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createdPolicy, err := r.client.CreateSessionPolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create session policy, got error: %s", err))
		return
	}

	data.ID = types.StringValue(createdPolicy.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back to populate computed values (priority, rule IDs, metadata)
	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State}
	r.Read(ctx, readReq, readResp)
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

func (r *SessionPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Read - Reading session policy")
	var data SessionPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetSessionPolicy(ctx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read session policy, got error: %s", err))
		return
	}

	sessionPolicyToModel(ctx, policy, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SessionPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Update - Updating session policy")
	var data SessionPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state to get the server-assigned priority
	var state SessionPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Carry forward state-assigned priority if plan left it unknown/null
	if data.Priority.IsUnknown() || data.Priority.IsNull() {
		data.Priority = state.Priority
	}

	policy := sessionPolicyFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Priority is required on PUT — always set it explicitly from state/plan value
	p := int(data.Priority.ValueInt64())
	policy.Priority = &p

	err := r.client.UpdateSessionPolicy(ctx, data.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update session policy, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State}
	r.Read(ctx, readReq, readResp)
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

func (r *SessionPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.Delete - Deleting session policy")
	var data SessionPolicyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSessionPolicy(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete session policy, got error: %s", err))
		return
	}
}

func (r *SessionPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "spa-terraform-provider: SessionPolicyResource.ImportState - Importing session policy", map[string]any{
		"import_id": req.ID,
	})

	data := SessionPolicyResourceModel{
		ID: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State}
	r.Read(ctx, readReq, readResp)
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

// ─── helpers ────────────────────────────────────────────────────────────────

// sessionPolicyFromModel converts a Terraform model into an API SessionPolicy struct.
func sessionPolicyFromModel(ctx context.Context, data *SessionPolicyResourceModel, diags *diag.Diagnostics) *SessionPolicy {
	policy := &SessionPolicy{
		Name:   data.Name.ValueString(),
		Active: data.Active.ValueBool(),
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		policy.Description = data.Description.ValueString()
	}
	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		p := int(data.Priority.ValueInt64())
		policy.Priority = &p
	}

	rules := make([]SessionPolicyRule, 0, len(data.Rules))
	for _, ruleData := range data.Rules {
		rule := SessionPolicyRule{
			Priority: int(ruleData.Priority.ValueInt64()),
			Active:   ruleData.Active.ValueBool(),
		}
		if !ruleData.ID.IsNull() && !ruleData.ID.IsUnknown() {
			rule.ID = ruleData.ID.ValueString()
		}
		if !ruleData.Name.IsNull() && !ruleData.Name.IsUnknown() {
			rule.Name = ruleData.Name.ValueString()
		}
		if !ruleData.Description.IsNull() && !ruleData.Description.IsUnknown() {
			rule.Description = ruleData.Description.ValueString()
		}

		// Convert actions
		if ruleData.Actions != nil {
			if !ruleData.Actions.Routing.IsNull() && !ruleData.Actions.Routing.IsUnknown() {
				rule.Actions.Routing = ruleData.Actions.Routing.ValueString()
			}
			if !ruleData.Actions.DisableSecurityGroups.IsNull() && !ruleData.Actions.DisableSecurityGroups.IsUnknown() {
				rule.Actions.DisableSecurityGroups = ruleData.Actions.DisableSecurityGroups.ValueString()
			}
			if !ruleData.Actions.LocalLanAccess.IsNull() && !ruleData.Actions.LocalLanAccess.IsUnknown() {
				rule.Actions.LocalLanAccess = ruleData.Actions.LocalLanAccess.ValueString()
			}
		}

		// Convert conditions — always initialize as empty slice so it serializes as [] not null
		conditions := make([]SessionPolicyCondition, 0)
		for _, condData := range ruleData.Conditions {
			cond := SessionPolicyCondition{
				Type:     condData.Type.ValueString(),
				Operator: condData.Operator.ValueString(),
			}
			if !condData.TagSource.IsNull() && !condData.TagSource.IsUnknown() {
				cond.TagSource = condData.TagSource.ValueString()
			}
			if !condData.TagKey.IsNull() && !condData.TagKey.IsUnknown() {
				cond.TagKey = condData.TagKey.ValueString()
			}
			if !condData.Values.IsNull() && !condData.Values.IsUnknown() {
				var values []string
				diags.Append(condData.Values.ElementsAs(ctx, &values, false)...)
				cond.Values = values
			}
			if !condData.Metadata.IsNull() && !condData.Metadata.IsUnknown() {
				metaStr := make(map[string]string)
				diags.Append(condData.Metadata.ElementsAs(ctx, &metaStr, false)...)
				meta := make(map[string]interface{})
				for k, v := range metaStr {
					meta[k] = v
				}
				cond.Metadata = meta
			}
			conditions = append(conditions, cond)
		}
		rule.Conditions = conditions

		rules = append(rules, rule)
	}
	policy.GenericRules = rules
	return policy
}

// sessionPolicyToModel maps an API SessionPolicy response back into a Terraform model.
func sessionPolicyToModel(ctx context.Context, policy *SessionPolicy, data *SessionPolicyResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Active = types.BoolValue(policy.Active)
	if policy.Priority != nil {
		data.Priority = types.Int64Value(int64(*policy.Priority))
	} else {
		data.Priority = types.Int64Value(0)
	}
	data.Description = types.StringValue(policy.Description)

	rules := make([]SessionPolicyRuleModel, 0, len(policy.GenericRules))
	for _, rule := range policy.GenericRules {
		ruleModel := SessionPolicyRuleModel{
			Priority: types.Int64Value(int64(rule.Priority)),
			Active:   types.BoolValue(rule.Active),
		}
		if rule.ID != "" {
			ruleModel.ID = types.StringValue(rule.ID)
		} else {
			ruleModel.ID = types.StringNull()
		}
		if rule.Name != "" {
			ruleModel.Name = types.StringValue(rule.Name)
		} else {
			ruleModel.Name = types.StringNull()
		}
		ruleModel.Description = types.StringValue(rule.Description)

		// Map actions
		if rule.Actions.Routing != "" || rule.Actions.DisableSecurityGroups != "" || rule.Actions.LocalLanAccess != "" {
			actionsModel := &SessionPolicyActionsModel{
				Routing:               types.StringNull(),
				DisableSecurityGroups: types.StringNull(),
				LocalLanAccess:        types.StringNull(),
			}
			if rule.Actions.Routing != "" {
				actionsModel.Routing = types.StringValue(rule.Actions.Routing)
			}
			if rule.Actions.DisableSecurityGroups != "" {
				actionsModel.DisableSecurityGroups = types.StringValue(rule.Actions.DisableSecurityGroups)
			}
			if rule.Actions.LocalLanAccess != "" {
				actionsModel.LocalLanAccess = types.StringValue(rule.Actions.LocalLanAccess)
			}
			ruleModel.Actions = actionsModel
		}

		// Map conditions — only set when the API returns at least one condition;
		// leaving Conditions nil preserves null in state for rules without conditions,
		// avoiding null-vs-[] drift for the Optional condition attribute.
		if len(rule.Conditions) > 0 {
			conditions := make([]SessionPolicyConditionModel, 0, len(rule.Conditions))
			for _, cond := range rule.Conditions {
				valuesVals := make([]attr.Value, 0, len(cond.Values))
				for _, v := range cond.Values {
					valuesVals = append(valuesVals, types.StringValue(v))
				}
				valuesList, d := types.ListValue(types.StringType, valuesVals)
				diags.Append(d...)

				metaAttrMap := make(map[string]attr.Value)
				if cond.Metadata != nil {
					for k, v := range cond.Metadata {
						metaAttrMap[k] = types.StringValue(fmt.Sprintf("%v", v))
					}
				}
				var metadata types.Map
				if len(metaAttrMap) > 0 {
					var mapDiags diag.Diagnostics
					metadata, mapDiags = types.MapValue(types.StringType, metaAttrMap)
					diags.Append(mapDiags...)
				} else {
					metadata = types.MapNull(types.StringType)
				}

				tagSource := types.StringValue(cond.TagSource)
				tagKey := types.StringValue(cond.TagKey)
				condModel := SessionPolicyConditionModel{
					Type:      types.StringValue(cond.Type),
					Operator:  types.StringValue(cond.Operator),
					TagSource: tagSource,
					TagKey:    tagKey,
					Values:    valuesList,
					Metadata:  metadata,
				}
				conditions = append(conditions, condModel)
			}
			ruleModel.Conditions = conditions
		} // end if len(rule.Conditions) > 0

		rules = append(rules, ruleModel)
	}
	data.Rules = rules
}
