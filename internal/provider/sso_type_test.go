package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSSOType_SemanticEquals(t *testing.T) {
	ctx := context.Background()

	// Create object types for both with the same content
	elements1 := make(map[string]attr.Value)
	elementTypes1 := make(map[string]attr.Type)
	elements1["provider"] = types.StringValue("saml")
	elementTypes1["provider"] = types.StringType

	elements2 := make(map[string]attr.Value)
	elementTypes2 := make(map[string]attr.Type)
	elements2["provider"] = types.StringValue("saml")
	elementTypes2["provider"] = types.StringType

	// Convert to ObjectValue and then to DynamicValue
	obj1, _ := types.ObjectValue(elementTypes1, elements1)
	obj2, _ := types.ObjectValue(elementTypes2, elements2)

	ssoValue1 := SSOTypeValue(types.DynamicValue(obj1))
	ssoValue2 := SSOTypeValue(types.DynamicValue(obj2))

	// Test semantic equality
	equal, diags := ssoValue1.DynamicSemanticEquals(ctx, ssoValue2)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}

	if !equal {
		t.Error("Expected semantically equal SSO values to be equal")
	}
}

func TestSSOType_CustomType(t *testing.T) {
	ctx := context.Background()
	ssoType := SSOType{}

	// Test String() method
	if ssoType.String() != "SSOType" {
		t.Errorf("Expected 'SSOType', got '%s'", ssoType.String())
	}

	// Test transformTupleToList method
	// Create a mock object with custom_attributes as Tuple
	attributeTypes := map[string]tftypes.Type{
		"provider":          tftypes.String,
		"custom_attributes": tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.String}},
	}
	objType := tftypes.Object{AttributeTypes: attributeTypes}
	
	objValue := map[string]tftypes.Value{
		"provider":          tftypes.NewValue(tftypes.String, "saml"),
		"custom_attributes": tftypes.NewValue(tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.String}}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "attr1"),
			tftypes.NewValue(tftypes.String, "attr2"),
		}),
	}
	
	inputValue := tftypes.NewValue(objType, objValue)
	
	// Transform the value
	transformedValue, err := ssoType.transformTupleToList(ctx, inputValue)
	if err != nil {
		t.Fatalf("transformTupleToList failed: %v", err)
	}
	
	// Verify that custom_attributes is now a List type
	transformedObjType := transformedValue.Type().(tftypes.Object)
	customAttrType := transformedObjType.AttributeTypes["custom_attributes"]
	
	if !customAttrType.Is(tftypes.List{}) {
		t.Errorf("Expected custom_attributes to be transformed to List type, got %T", customAttrType)
	}

	// Test ValueType() method
	valueType := ssoType.ValueType(ctx)
	if _, ok := valueType.(SSOValue); !ok {
		t.Errorf("Expected SSOValue, got %T", valueType)
	}

	// Test Equal() method
	anotherSSOType := SSOType{}
	if !ssoType.Equal(anotherSSOType) {
		t.Error("Expected equal SSO types to be equal")
	}

	// Test Equal() with different type
	if ssoType.Equal(types.StringType) {
		t.Error("Expected SSO type to not equal string type")
	}
}

func TestSSOValue_Type(t *testing.T) {
	ctx := context.Background()
	ssoValue := SSOTypeNull()

	// Test Type() method
	valueType := ssoValue.Type(ctx)
	if _, ok := valueType.(SSOType); !ok {
		t.Errorf("Expected SSOType, got %T", valueType)
	}

	// Test Equal() method
	anotherSSOValue := SSOTypeNull()
	if !ssoValue.Equal(anotherSSOValue) {
		t.Error("Expected equal SSO values to be equal")
	}

	// Test Equal() with different type
	if ssoValue.Equal(types.StringNull()) {
		t.Error("Expected SSO value to not equal string value")
	}
}
