package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// IconType is a custom string type for the application `icon` attribute. Two
// values compare equal when their normalized base64 payloads match, so a
// hand-written icon carrying a `data:<mime>;base64,` prefix or wrapped
// (newline/space-separated) base64 does not produce a spurious diff or an
// "inconsistent result after apply" against the prefix-stripped, whitespace-free
// value the SPA backend stores and returns.
type IconType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = IconType{}

func (t IconType) Equal(o attr.Type) bool {
	other, ok := o.(IconType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t IconType) String() string {
	return "IconType"
}

func (t IconType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return IconValue{StringValue: in}, nil
}

func (t IconType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}
	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}
	return stringValuable, nil
}

func (t IconType) ValueType(_ context.Context) attr.Value {
	return IconValue{}
}

// IconValue is the value type produced by IconType.
type IconValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuableWithSemanticEquals = IconValue{}

func (v IconValue) Type(_ context.Context) attr.Type {
	return IconType{}
}

func (v IconValue) Equal(o attr.Value) bool {
	other, ok := o.(IconValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

var iconWhitespace = regexp.MustCompile(`\s+`)

// normalizeIcon strips an optional `data:<mime>;base64,` prefix and removes all
// whitespace, matching how the SPA backend stores an icon.
func normalizeIcon(s string) string {
	if i := strings.Index(s, "base64,"); i >= 0 {
		s = s[i+len("base64,"):]
	}
	return iconWhitespace.ReplaceAllString(s, "")
}

func (v IconValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newValue, ok := newValuable.(IconValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("expected value type %T but got %T", v, newValuable),
		)
		return false, diags
	}
	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return v.StringValue.Equal(newValue.StringValue), diags
	}
	return normalizeIcon(v.ValueString()) == normalizeIcon(newValue.ValueString()), diags
}

// NewIconValue builds a known IconValue from a raw string.
func NewIconValue(s string) IconValue {
	return IconValue{StringValue: basetypes.NewStringValue(s)}
}

// NewIconNull builds a null IconValue.
func NewIconNull() IconValue {
	return IconValue{StringValue: basetypes.NewStringNull()}
}
