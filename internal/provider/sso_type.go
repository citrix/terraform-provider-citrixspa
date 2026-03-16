package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.DynamicTypable = SSOType{}

// SSOType defines a custom type for SSO configuration that provides more control
// over the data structure and semantic equality checking.
type SSOType struct {
	basetypes.DynamicType
}

func (t SSOType) Equal(o attr.Type) bool {
	other, ok := o.(SSOType)
	if !ok {
		return false
	}
	return t.DynamicType.Equal(other.DynamicType)
}

func (t SSOType) String() string {
	return "SSOType"
}

func (t SSOType) ValueFromDynamic(ctx context.Context, in basetypes.DynamicValue) (basetypes.DynamicValuable, diag.Diagnostics) {
	// tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromDynamic called")
	value := SSOValue{
		DynamicValue: in,
	}
	return value, nil
}

func (t SSOType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromTerraform called")
	tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromTerraform in value", map[string]any{
		"in_type":  in.Type().String(),
		"in_value": in.String(),
	})

	attrValue, err := t.DynamicType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromTerraform ValueFromTerraform completed", map[string]any{
		"value": attrValue.String(),
	})

	dynamicValue, ok := attrValue.(basetypes.DynamicValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	if dynamicValue.IsUnderlyingValueNull() {
		tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromTerraform dynamic value is null")
	} else if dynamicValue.IsUnderlyingValueUnknown() {
		tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromTerraform dynamic value is unknown")
	} else {
		v := dynamicValue.UnderlyingValue()
		if v != nil {
			tflog.Debug(ctx, "spa-terraform-provider: SSOType.ValueFromTerraform dynamic value is known", map[string]any{
				"dynamic_type":  v.Type(ctx).String(),
				"dynamic_value": v.String(),
			})
		}
	}
	dynamicValuable, diags := t.ValueFromDynamic(ctx, dynamicValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting DynamicValue to DynamicValuable: %v", diags)
	}

	return dynamicValuable, nil
}

func (t SSOType) ValueType(ctx context.Context) attr.Value {
	return SSOValue{}
}

// transformTupleToList transforms tftypes.Tuple fields to tftypes.List for specific attributes
// This specifically handles the "custom_attributes" field which should be a list of Dynamic values
// instead of a Tuple
func (t SSOType) transformTupleToList(ctx context.Context, in tftypes.Value) (tftypes.Value, error) {
	tflog.Debug(ctx, "spa-terraform-provider: SSOType.transformTupleToList called")

	// If the value is null or unknown, return as-is
	if in.IsNull() || !in.IsKnown() {
		return in, nil
	}

	// Only process Object types that might contain the custom_attributes field
	if !in.Type().Is(tftypes.Object{}) {
		return in, nil
	}

	objType := in.Type().(tftypes.Object)

	// Check if this object has a custom_attributes field with Tuple type
	customAttrType, hasCustomAttr := objType.AttributeTypes["custom_attributes"]
	if !hasCustomAttr {
		// No custom_attributes field, return as-is
		return in, nil
	}

	// Check if the custom_attributes field is a Tuple type
	if !customAttrType.Is(tftypes.Tuple{}) {
		// Not a Tuple, return as-is
		return in, nil
	}

	tflog.Debug(ctx, "spa-terraform-provider: Found custom_attributes with Tuple type, converting to List")

	// Extract the current object values
	var objValue map[string]tftypes.Value
	err := in.As(&objValue)
	if err != nil {
		return tftypes.Value{}, fmt.Errorf("failed to extract object values: %w", err)
	}

	// Transform the custom_attributes field from Tuple to List
	customAttrValue := objValue["custom_attributes"]
	if !customAttrValue.IsNull() && customAttrValue.IsKnown() {
		// Extract tuple elements
		var tupleElements []tftypes.Value
		err := customAttrValue.As(&tupleElements)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("failed to extract tuple elements: %w", err)
		}

		// Create a List type with DynamicPseudoType as element type
		listType := tftypes.List{ElementType: tftypes.DynamicPseudoType}

		// Create the list value with the same elements
		listValue := tftypes.NewValue(listType, tupleElements)

		// Update the object value
		objValue["custom_attributes"] = listValue

		// Create new object type with List instead of Tuple for custom_attributes
		newAttributeTypes := make(map[string]tftypes.Type)
		for key, attrType := range objType.AttributeTypes {
			if key == "custom_attributes" {
				newAttributeTypes[key] = listType
			} else {
				newAttributeTypes[key] = attrType
			}
		}

		newObjType := tftypes.Object{AttributeTypes: newAttributeTypes}
		transformedValue := tftypes.NewValue(newObjType, objValue)

		tflog.Debug(ctx, "spa-terraform-provider: Successfully transformed custom_attributes from Tuple to List")
		return transformedValue, nil
	}

	return in, nil
}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.DynamicValuable = SSOValue{}
var _ basetypes.DynamicValuableWithSemanticEquals = SSOValue{}

// SSOValue represents an SSO configuration value with custom semantic equality logic.
type SSOValue struct {
	basetypes.DynamicValue
}

func (v SSOValue) Equal(o attr.Value) bool {
	other, ok := o.(SSOValue)
	if !ok {
		return false
	}
	return v.DynamicValue.Equal(other.DynamicValue)
}

func (v SSOValue) Type(ctx context.Context) attr.Type {
	return SSOType{}
}

// DynamicSemanticEquals implements semantic equality for SSO values.
// This helps prevent unnecessary plan changes when the SSO data is semantically equivalent
// but might have minor structural differences (e.g., different property ordering in objects).
func (v SSOValue) DynamicSemanticEquals(ctx context.Context, newValuable basetypes.DynamicValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	tflog.Debug(ctx, "spa-terraform-provider: SSOValue.DynamicSemanticEquals called")
	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(SSOValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	// If both values are null or unknown, they are equal
	if v.IsNull() && newValue.IsNull() {
		return true, diags
	}
	if v.IsUnknown() && newValue.IsUnknown() {
		return true, diags
	}
	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return false, diags
	}

	// Get the underlying values for comparison
	priorUnderlying := v.UnderlyingValue()
	newUnderlying := newValue.UnderlyingValue()

	// Convert both to Go interface{} for deep comparison
	priorGo, err := convertToGoValue(ctx, priorUnderlying)
	if err != nil {
		diags.AddWarning(
			"Semantic Equality Check Warning",
			fmt.Sprintf("Could not convert prior SSO value for comparison: %v", err),
		)
		return false, diags
	}

	newGo, err := convertToGoValue(ctx, newUnderlying)
	if err != nil {
		diags.AddWarning(
			"Semantic Equality Check Warning",
			fmt.Sprintf("Could not convert new SSO value for comparison: %v", err),
		)
		return false, diags
	}

	// Perform deep semantic comparison
	return deepEqualSSO(ctx, priorGo, newGo), diags
}

// Helper function to convert Terraform values to Go values for comparison
func convertToGoValue(ctx context.Context, val attr.Value) (interface{}, error) {
	tflog.Debug(ctx, "spa-terraform-provider: convertToGoValue called")
	switch v := val.(type) {
	case basetypes.StringValue:
		return v.ValueString(), nil
	case basetypes.BoolValue:
		return v.ValueBool(), nil
	case basetypes.Int64Value:
		return v.ValueInt64(), nil
	case basetypes.NumberValue:
		f, _ := v.ValueBigFloat().Float64()
		return f, nil
	case basetypes.ListValue:
		elements := v.Elements()
		result := make([]interface{}, len(elements))
		for i, elem := range elements {
			converted, err := convertToGoValue(ctx, elem)
			if err != nil {
				return nil, err
			}
			result[i] = converted
		}
		return result, nil
	case basetypes.ObjectValue:
		attributes := v.Attributes()
		result := make(map[string]interface{})
		for key, attr := range attributes {
			converted, err := convertToGoValue(ctx, attr)
			if err != nil {
				return nil, err
			}
			result[key] = converted
		}
		return result, nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// deepEqualSSO performs deep semantic equality comparison for SSO values
// This function handles the complex nested structure of SSO data and considers
// them equal if they represent the same logical configuration, regardless of
// minor differences in representation.
func deepEqualSSO(ctx context.Context, a, b interface{}) bool {
	// Convert both to JSON and back to normalize the structure
	// This handles cases where the same data might be represented slightly differently
	tflog.Debug(ctx, "spa-terraform-provider: deepEqualSSO called")
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}

	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}

	// Parse back to interface{} to normalize types
	var aNormalized, bNormalized interface{}
	if err := json.Unmarshal(aJSON, &aNormalized); err != nil {
		return false
	}
	if err := json.Unmarshal(bJSON, &bNormalized); err != nil {
		return false
	}

	// Convert back to JSON strings for comparison
	// This ensures consistent ordering and type representation
	aNormalizedJSON, err := json.Marshal(aNormalized)
	if err != nil {
		return false
	}

	bNormalizedJSON, err := json.Marshal(bNormalized)
	if err != nil {
		return false
	}

	return string(aNormalizedJSON) == string(bNormalizedJSON)
}

// SSOTypeNull creates a null SSO value
func SSOTypeNull() SSOValue {
	return SSOValue{
		DynamicValue: basetypes.NewDynamicNull(),
	}
}

// SSOTypeUnknown creates an unknown SSO value
func SSOTypeUnknown() SSOValue {
	return SSOValue{
		DynamicValue: basetypes.NewDynamicUnknown(),
	}
}

// SSOTypeValue creates an SSO value from a Dynamic value
func SSOTypeValue(value basetypes.DynamicValue) SSOValue {
	return SSOValue{
		DynamicValue: value,
	}
}
