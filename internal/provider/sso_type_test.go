package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSSOFromAPI_SAML(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":                  "saml",
		"assertion_url":         "https://sp.example.com/acs",
		"audience":              "https://sp.example.com",
		"name_id_format":        "emailAddress",
		"name_id_source":        "email",
		"sp_initiated_only":     false,
		"saml_sso_login_url":    "https://generated-url.example.com",
		"saml_cert_issuer_name": "issuer",
		"customer":              "test-customer",
		"custom_attributes":     "[]",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model == nil {
		t.Fatal("Expected non-nil model")
	}

	// Check user-provided fields
	if model.Type.ValueString() != "saml" {
		t.Errorf("Expected type 'saml', got %q", model.Type.ValueString())
	}
	if model.AssertionURL.ValueString() != "https://sp.example.com/acs" {
		t.Errorf("Expected assertion_url, got %q", model.AssertionURL.ValueString())
	}
	if model.Audience.ValueString() != "https://sp.example.com" {
		t.Errorf("Expected audience, got %q", model.Audience.ValueString())
	}
	if model.NameIDFormat.ValueString() != "emailAddress" {
		t.Errorf("Expected name_id_format 'emailAddress', got %q", model.NameIDFormat.ValueString())
	}
	if model.SpInitiatedOnly.ValueBool() != false {
		t.Errorf("Expected sp_initiated_only false")
	}

	// Check server-computed fields are populated
	if model.SamlSSOLoginURL.ValueString() != "https://generated-url.example.com" {
		t.Errorf("Expected saml_sso_login_url, got %q", model.SamlSSOLoginURL.ValueString())
	}
	if model.SamlCertIssuerName.ValueString() != "issuer" {
		t.Errorf("Expected saml_cert_issuer_name 'issuer', got %q", model.SamlCertIssuerName.ValueString())
	}

	// custom_attributes parsed from JSON string "[]" should be empty list (not null)
	if model.CustomAttributes.IsNull() {
		t.Error("Expected custom_attributes to be empty list, got null")
	}
}

func TestSSOFromAPI_Nosso(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":     "nosso",
		"customer": "test-customer",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}

	if model.Type.ValueString() != "nosso" {
		t.Errorf("Expected type 'nosso', got %q", model.Type.ValueString())
	}
	// All other fields should be null
	if !model.AssertionURL.IsNull() {
		t.Error("Expected assertion_url to be null for nosso")
	}
	if !model.SamlSSOLoginURL.IsNull() {
		t.Error("Expected saml_sso_login_url to be null for nosso")
	}
}

func TestSSOFromAPI_Form(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":           "form",
		"action_url":     "https://app.example.com/login",
		"logonform_url":  "https://app.example.com/logon",
		"username_field": "user",
		"password_field": "pass",
		"attribute":      "email",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}

	if model.Type.ValueString() != "form" {
		t.Errorf("Expected type 'form', got %q", model.Type.ValueString())
	}
	if model.ActionURL.ValueString() != "https://app.example.com/login" {
		t.Errorf("Expected action_url, got %q", model.ActionURL.ValueString())
	}
	if model.UsernameField.ValueString() != "user" {
		t.Errorf("Expected username_field 'user', got %q", model.UsernameField.ValueString())
	}
}

func TestSSOFromAPI_Empty(t *testing.T) {
	ctx := context.Background()

	model, diags := ssoFromAPI(ctx, map[string]any{})
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model != nil {
		t.Error("Expected nil model for empty SSO map")
	}
}

func TestSSOToAPI_SAML(t *testing.T) {
	ctx := context.Background()

	model := &SSOModel{
		Type:            types.StringValue("saml"),
		AssertionURL:    types.StringValue("https://sp.example.com/acs"),
		Audience:        types.StringValue("https://sp.example.com"),
		NameIDFormat:    types.StringValue("emailAddress"),
		NameIDSource:    types.StringValue("email"),
		SpInitiatedOnly: types.BoolValue(false),
		// Computed fields — should NOT appear in output
		SamlSSOLoginURL:    types.StringValue("https://generated.example.com"),
		SamlCertIssuerName: types.StringValue("issuer"),
		// Unset fields
		SamlType:         types.StringNull(),
		RelayState:       types.StringNull(),
		SignAssertion:    types.StringNull(),
		CustomAttributes: types.ListNull(CustomAttributeObjectType),
		ActionURL:        types.StringNull(),
		LogonformURL:     types.StringNull(),
		UsernameField:    types.StringNull(),
		PasswordField:    types.StringNull(),
		Attribute:        types.StringNull(),
		UsernameFormat:   types.StringNull(),
		UserRealm:        types.StringNull(),
	}

	result, diags := ssoToAPI(ctx, model)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}

	// Check user-provided fields are present
	if result["type"] != "saml" {
		t.Errorf("Expected type 'saml', got %v", result["type"])
	}
	if result["assertion_url"] != "https://sp.example.com/acs" {
		t.Errorf("Expected assertion_url, got %v", result["assertion_url"])
	}

	// Check server-computed fields are NOT present
	if _, ok := result["saml_sso_login_url"]; ok {
		t.Error("saml_sso_login_url should not be in API output")
	}
	if _, ok := result["saml_cert_issuer_name"]; ok {
		t.Error("saml_cert_issuer_name should not be in API output")
	}

	// Check null fields are not present
	if _, ok := result["saml_type"]; ok {
		t.Error("Null saml_type should not be in API output")
	}
}

func TestSSOToAPI_Nil(t *testing.T) {
	ctx := context.Background()

	result, diags := ssoToAPI(ctx, nil)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if result != nil {
		t.Error("Expected nil result for nil model")
	}
}

func TestSSOFromAPI_MissingType(t *testing.T) {
	ctx := context.Background()

	// API returns SSO data without a "type" field — ssoFromAPI should return nil
	// rather than an object with type=null, which violates the schema.
	apiSSO := map[string]any{
		"assertion_url": "https://sp.example.com/acs",
		"audience":      "https://sp.example.com",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model != nil {
		t.Error("Expected nil model when type is missing from SSO map")
	}
}

func TestCustomAttributesFromAPI_JSONString(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":              "saml",
		"custom_attributes": `[{"format":"uri","name":"attr1","value":"val1"}]`,
	}

	model, d := ssoFromAPI(ctx, apiSSO)
	if d.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", d)
	}

	if model.CustomAttributes.IsNull() {
		t.Fatal("Expected non-null custom_attributes")
	}

	elements := model.CustomAttributes.Elements()
	if len(elements) != 1 {
		t.Fatalf("Expected 1 custom attribute, got %d", len(elements))
	}
}

// --- ssoFromAPI: additional SSO types and edge cases ---

func TestSSOFromAPI_Kerberos(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":            "kerberos",
		"user_realm":      "EXAMPLE.COM",
		"username_format": "upn",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model.Type.ValueString() != "kerberos" {
		t.Errorf("Expected type 'kerberos', got %q", model.Type.ValueString())
	}
	if model.UserRealm.ValueString() != "EXAMPLE.COM" {
		t.Errorf("Expected user_realm 'EXAMPLE.COM', got %q", model.UserRealm.ValueString())
	}
	if model.UsernameFormat.ValueString() != "upn" {
		t.Errorf("Expected username_format 'upn', got %q", model.UsernameFormat.ValueString())
	}
	// SAML fields should be null
	if !model.AssertionURL.IsNull() {
		t.Error("Expected assertion_url to be null for kerberos")
	}
}

func TestSSOFromAPI_Basic(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":            "basic",
		"username_format": "email",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model.Type.ValueString() != "basic" {
		t.Errorf("Expected type 'basic', got %q", model.Type.ValueString())
	}
	if model.UsernameFormat.ValueString() != "email" {
		t.Errorf("Expected username_format 'email', got %q", model.UsernameFormat.ValueString())
	}
}

func TestSSOFromAPI_NilMap(t *testing.T) {
	ctx := context.Background()

	model, diags := ssoFromAPI(ctx, nil)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model != nil {
		t.Error("Expected nil model for nil SSO map")
	}
}

func TestSSOFromAPI_NullFieldValues(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":           "saml",
		"assertion_url":  nil,
		"audience":       nil,
		"name_id_format": nil,
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model.Type.ValueString() != "saml" {
		t.Errorf("Expected type 'saml', got %q", model.Type.ValueString())
	}
	if !model.AssertionURL.IsNull() {
		t.Error("Expected assertion_url to be null when API value is nil")
	}
	if !model.Audience.IsNull() {
		t.Error("Expected audience to be null when API value is nil")
	}
}

func TestSSOFromAPI_PartialSAML(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":          "saml",
		"assertion_url": "https://sp.example.com/acs",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model.AssertionURL.ValueString() != "https://sp.example.com/acs" {
		t.Errorf("Expected assertion_url, got %q", model.AssertionURL.ValueString())
	}
	// All other SAML fields should be null
	if !model.Audience.IsNull() {
		t.Error("Expected audience null for partial SAML")
	}
	if !model.NameIDFormat.IsNull() {
		t.Error("Expected name_id_format null for partial SAML")
	}
	if !model.SignAssertion.IsNull() {
		t.Error("Expected sign_assertion null for partial SAML")
	}
	if !model.SpInitiatedOnly.IsNull() {
		t.Error("Expected sp_initiated_only null for partial SAML")
	}
}

// --- ssoToAPI: additional SSO types ---

func TestSSOToAPI_Form(t *testing.T) {
	ctx := context.Background()

	model := &SSOModel{
		Type:           types.StringValue("form"),
		ActionURL:      types.StringValue("https://app.example.com/login"),
		LogonformURL:   types.StringValue("https://app.example.com/logon"),
		UsernameField:  types.StringValue("user"),
		PasswordField:  types.StringValue("pass"),
		Attribute:      types.StringValue("email"),
		UsernameFormat: types.StringValue("upn"),
		// Null SAML fields
		SamlType:           types.StringNull(),
		SpInitiatedOnly:    types.BoolNull(),
		AssertionURL:       types.StringNull(),
		Audience:           types.StringNull(),
		RelayState:         types.StringNull(),
		SignAssertion:      types.StringNull(),
		NameIDSource:       types.StringNull(),
		NameIDFormat:       types.StringNull(),
		CustomAttributes:   types.ListNull(CustomAttributeObjectType),
		SamlSSOLoginURL:    types.StringNull(),
		SamlCertIssuerName: types.StringNull(),
		UserRealm:          types.StringNull(),
	}

	result, diags := ssoToAPI(ctx, model)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if result["type"] != "form" {
		t.Errorf("Expected type 'form', got %v", result["type"])
	}
	if result["action_url"] != "https://app.example.com/login" {
		t.Errorf("Expected action_url, got %v", result["action_url"])
	}
	if result["username_field"] != "user" {
		t.Errorf("Expected username_field, got %v", result["username_field"])
	}
	if result["password_field"] != "pass" {
		t.Errorf("Expected password_field, got %v", result["password_field"])
	}
	// Null SAML fields should not appear
	if _, ok := result["assertion_url"]; ok {
		t.Error("Null assertion_url should not be in API output for form SSO")
	}
}

func TestSSOToAPI_Kerberos(t *testing.T) {
	ctx := context.Background()

	model := &SSOModel{
		Type:               types.StringValue("kerberos"),
		UserRealm:          types.StringValue("EXAMPLE.COM"),
		UsernameFormat:     types.StringValue("upn"),
		SamlType:           types.StringNull(),
		SpInitiatedOnly:    types.BoolNull(),
		AssertionURL:       types.StringNull(),
		Audience:           types.StringNull(),
		RelayState:         types.StringNull(),
		SignAssertion:      types.StringNull(),
		NameIDSource:       types.StringNull(),
		NameIDFormat:       types.StringNull(),
		CustomAttributes:   types.ListNull(CustomAttributeObjectType),
		SamlSSOLoginURL:    types.StringNull(),
		SamlCertIssuerName: types.StringNull(),
		ActionURL:          types.StringNull(),
		LogonformURL:       types.StringNull(),
		UsernameField:      types.StringNull(),
		PasswordField:      types.StringNull(),
		Attribute:          types.StringNull(),
	}

	result, diags := ssoToAPI(ctx, model)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if result["type"] != "kerberos" {
		t.Errorf("Expected type 'kerberos', got %v", result["type"])
	}
	if result["user_realm"] != "EXAMPLE.COM" {
		t.Errorf("Expected user_realm, got %v", result["user_realm"])
	}
	if _, ok := result["action_url"]; ok {
		t.Error("Form fields should not appear in kerberos output")
	}
}

func TestSSOToAPI_Nosso(t *testing.T) {
	ctx := context.Background()

	model := &SSOModel{
		Type:               types.StringValue("nosso"),
		SamlType:           types.StringNull(),
		SpInitiatedOnly:    types.BoolNull(),
		AssertionURL:       types.StringNull(),
		Audience:           types.StringNull(),
		RelayState:         types.StringNull(),
		SignAssertion:      types.StringNull(),
		NameIDSource:       types.StringNull(),
		NameIDFormat:       types.StringNull(),
		CustomAttributes:   types.ListNull(CustomAttributeObjectType),
		SamlSSOLoginURL:    types.StringNull(),
		SamlCertIssuerName: types.StringNull(),
		ActionURL:          types.StringNull(),
		LogonformURL:       types.StringNull(),
		UsernameField:      types.StringNull(),
		PasswordField:      types.StringNull(),
		Attribute:          types.StringNull(),
		UsernameFormat:     types.StringNull(),
		UserRealm:          types.StringNull(),
	}

	result, diags := ssoToAPI(ctx, model)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if result["type"] != "nosso" {
		t.Errorf("Expected type 'nosso', got %v", result["type"])
	}
	// Only "type" should be present
	if len(result) != 1 {
		t.Errorf("Expected only 'type' key for nosso, got %d keys: %v", len(result), result)
	}
}

func TestSSOToAPI_WithCustomAttributes(t *testing.T) {
	ctx := context.Background()

	// Build custom attributes list
	ca1, _ := types.ObjectValue(customAttributeAttrTypes, map[string]attr.Value{
		"format":      types.StringValue("uri"),
		"name":        types.StringValue("attr1"),
		"value":       types.StringValue("val1"),
		"prefix_expr": types.BoolValue(false),
	})
	ca2, _ := types.ObjectValue(customAttributeAttrTypes, map[string]attr.Value{
		"format":      types.StringValue("basic"),
		"name":        types.StringValue("attr2"),
		"value":       types.StringValue("val2"),
		"prefix_expr": types.BoolNull(),
	})
	caList, _ := types.ListValue(CustomAttributeObjectType, []attr.Value{ca1, ca2})

	model := &SSOModel{
		Type:               types.StringValue("saml"),
		AssertionURL:       types.StringValue("https://sp.example.com/acs"),
		CustomAttributes:   caList,
		SamlType:           types.StringNull(),
		SpInitiatedOnly:    types.BoolNull(),
		Audience:           types.StringNull(),
		RelayState:         types.StringNull(),
		SignAssertion:      types.StringNull(),
		NameIDSource:       types.StringNull(),
		NameIDFormat:       types.StringNull(),
		SamlSSOLoginURL:    types.StringNull(),
		SamlCertIssuerName: types.StringNull(),
		ActionURL:          types.StringNull(),
		LogonformURL:       types.StringNull(),
		UsernameField:      types.StringNull(),
		PasswordField:      types.StringNull(),
		Attribute:          types.StringNull(),
		UsernameFormat:     types.StringNull(),
		UserRealm:          types.StringNull(),
	}

	result, diags := ssoToAPI(ctx, model)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}

	caResult, ok := result["custom_attributes"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected custom_attributes to be []map[string]any, got %T", result["custom_attributes"])
	}
	if len(caResult) != 2 {
		t.Fatalf("Expected 2 custom attributes, got %d", len(caResult))
	}
	if caResult[0]["name"] != "attr1" {
		t.Errorf("Expected first attr name 'attr1', got %v", caResult[0]["name"])
	}
	if caResult[1]["name"] != "attr2" {
		t.Errorf("Expected second attr name 'attr2', got %v", caResult[1]["name"])
	}
	// prefix_expr null should not appear in second entry
	if _, ok := caResult[1]["prefix_expr"]; ok {
		t.Error("Null prefix_expr should not be in API output")
	}
}

// --- ssoModelToObject / ssoObjectToModel ---

func TestSSOModelToObject_Nil(t *testing.T) {
	ctx := context.Background()

	obj, diags := ssoModelToObject(ctx, nil)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if !obj.IsNull() {
		t.Error("Expected null object for nil model")
	}
}

func TestSSOModelToObject_Roundtrip(t *testing.T) {
	ctx := context.Background()

	original := &SSOModel{
		Type:               types.StringValue("saml"),
		SamlType:           types.StringValue("SP"),
		SpInitiatedOnly:    types.BoolValue(true),
		AssertionURL:       types.StringValue("https://sp.example.com/acs"),
		Audience:           types.StringValue("https://sp.example.com"),
		RelayState:         types.StringNull(),
		SignAssertion:      types.StringValue("ASSERTION"),
		NameIDSource:       types.StringValue("email"),
		NameIDFormat:       types.StringValue("emailAddress"),
		CustomAttributes:   types.ListNull(CustomAttributeObjectType),
		SamlSSOLoginURL:    types.StringValue("https://login.example.com"),
		SamlCertIssuerName: types.StringValue("issuer"),
		ActionURL:          types.StringNull(),
		LogonformURL:       types.StringNull(),
		UsernameField:      types.StringNull(),
		PasswordField:      types.StringNull(),
		Attribute:          types.StringNull(),
		UsernameFormat:     types.StringNull(),
		UserRealm:          types.StringNull(),
	}

	obj, diags := ssoModelToObject(ctx, original)
	if diags.HasError() {
		t.Fatalf("ssoModelToObject diagnostics: %v", diags)
	}
	if obj.IsNull() {
		t.Fatal("Expected non-null object")
	}

	roundtripped, diags := ssoObjectToModel(ctx, obj)
	if diags.HasError() {
		t.Fatalf("ssoObjectToModel diagnostics: %v", diags)
	}
	if roundtripped == nil {
		t.Fatal("Expected non-nil model after roundtrip")
	}

	// Verify key fields survived roundtrip
	if roundtripped.Type.ValueString() != "saml" {
		t.Errorf("Roundtrip type: got %q", roundtripped.Type.ValueString())
	}
	if roundtripped.AssertionURL.ValueString() != "https://sp.example.com/acs" {
		t.Errorf("Roundtrip assertion_url: got %q", roundtripped.AssertionURL.ValueString())
	}
	if roundtripped.SpInitiatedOnly.ValueBool() != true {
		t.Error("Roundtrip sp_initiated_only: expected true")
	}
	if roundtripped.SignAssertion.ValueString() != "ASSERTION" {
		t.Errorf("Roundtrip sign_assertion: got %q", roundtripped.SignAssertion.ValueString())
	}
	if !roundtripped.RelayState.IsNull() {
		t.Error("Roundtrip relay_state: expected null")
	}
	if !roundtripped.ActionURL.IsNull() {
		t.Error("Roundtrip action_url: expected null")
	}
}

func TestSSOObjectToModel_NullObject(t *testing.T) {
	ctx := context.Background()

	model, diags := ssoObjectToModel(ctx, types.ObjectNull(ssoAttrTypes))
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model != nil {
		t.Error("Expected nil model for null object")
	}
}

func TestSSOObjectToModel_UnknownObject(t *testing.T) {
	ctx := context.Background()

	model, diags := ssoObjectToModel(ctx, types.ObjectUnknown(ssoAttrTypes))
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model != nil {
		t.Error("Expected nil model for unknown object")
	}
}

// --- customAttributesFromAPI: edge cases ---

func TestCustomAttributesFromAPI_EmptyNativeArray(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":              "saml",
		"custom_attributes": []any{},
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if model.CustomAttributes.IsNull() {
		t.Error("Expected empty list, not null, for empty native array")
	}
	if len(model.CustomAttributes.Elements()) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(model.CustomAttributes.Elements()))
	}
}

func TestCustomAttributesFromAPI_MultipleItems(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type": "saml",
		"custom_attributes": []any{
			map[string]any{"format": "uri", "name": "a1", "value": "v1"},
			map[string]any{"format": "basic", "name": "a2", "value": "v2", "prefix_expr": true},
			map[string]any{"name": "a3", "value": "v3"},
		},
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	elements := model.CustomAttributes.Elements()
	if len(elements) != 3 {
		t.Fatalf("Expected 3 custom attributes, got %d", len(elements))
	}
}

func TestCustomAttributesFromAPI_MalformedJSON(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":              "saml",
		"custom_attributes": "[{invalid json",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if !model.CustomAttributes.IsNull() {
		t.Error("Expected null custom_attributes for malformed JSON")
	}
}

func TestCustomAttributesFromAPI_NonArrayString(t *testing.T) {
	ctx := context.Background()

	apiSSO := map[string]any{
		"type":              "saml",
		"custom_attributes": "some plain string",
	}

	model, diags := ssoFromAPI(ctx, apiSSO)
	if diags.HasError() {
		t.Fatalf("Unexpected diagnostics: %v", diags)
	}
	if !model.CustomAttributes.IsNull() {
		t.Error("Expected null custom_attributes for non-array string")
	}
}
