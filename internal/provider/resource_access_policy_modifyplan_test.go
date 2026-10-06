package provider

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestModifyPlan_deferOnUnknownAccessRules verifies that ModifyPlan does not
// raise a value-conversion diagnostic when the access_rules collection is
// unknown (e.g. wired from another resource's computed output). access_rules
// decodes into a native Go slice that cannot hold an unknown value, so a
// whole-model Get would otherwise fail and block the plan; ModifyPlan must
// defer reconciliation to apply instead.
func TestModifyPlan_deferOnUnknownAccessRules(t *testing.T) {
	ctx := context.Background()
	r := NewAccessPolicyResource().(*AccessPolicyResource)
	sresp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, sresp)
	sch := sresp.Schema

	objType := sch.Type().TerraformType(ctx).(tftypes.Object)

	// Object with every attribute null except access_rules, whose value is set
	// by the caller (known-empty for state, unknown for plan/config).
	build := func(accessRules tftypes.Value) tftypes.Value {
		vals := map[string]tftypes.Value{}
		for name, at := range objType.AttributeTypes {
			if name == "access_rules" {
				vals[name] = accessRules
				continue
			}
			vals[name] = tftypes.NewValue(at, nil)
		}
		return tftypes.NewValue(objType, vals)
	}

	arType := objType.AttributeTypes["access_rules"]
	knownEmpty := tftypes.NewValue(arType, []tftypes.Value{})
	unknown := tftypes.NewValue(arType, tftypes.UnknownValue)

	req := fwresource.ModifyPlanRequest{
		State:  tfsdk.State{Schema: sch, Raw: build(knownEmpty)},
		Plan:   tfsdk.Plan{Schema: sch, Raw: build(unknown)},
		Config: tfsdk.Config{Schema: sch, Raw: build(unknown)},
	}
	resp := &fwresource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: sch, Raw: build(unknown)}}

	r.ModifyPlan(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan must defer (no error) on unknown access_rules, got: %v", resp.Diagnostics)
	}
}
