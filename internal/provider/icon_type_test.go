package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestNormalizeIcon(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"raw":              {"AAAAbbbb==", "AAAAbbbb=="},
		"data uri prefix":  {"data:image/png;base64,AAAAbbbb==", "AAAAbbbb=="},
		"wrapped newlines": {"AAAA\nbbbb==", "AAAAbbbb=="},
		"prefix + wrapped": {"data:image/png;base64,AAAA\n  bbbb==", "AAAAbbbb=="},
		"leading spaces":   {"  AAAAbbbb==  ", "AAAAbbbb=="},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := normalizeIcon(c.in); got != c.want {
				t.Fatalf("normalizeIcon(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestIconValueStringSemanticEquals(t *testing.T) {
	ctx := context.Background()
	cases := map[string]struct {
		a, b string
		want bool
	}{
		"identical":          {"AAAAbbbb==", "AAAAbbbb==", true},
		"data uri vs raw":    {"data:image/png;base64,AAAAbbbb==", "AAAAbbbb==", true},
		"wrapped vs raw":     {"AAAA\nbbbb==", "AAAAbbbb==", true},
		"different payloads": {"AAAAbbbb==", "CCCCdddd==", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			a := NewIconValue(c.a)
			b := NewIconValue(c.b)
			got, diags := a.StringSemanticEquals(ctx, b)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got != c.want {
				t.Fatalf("StringSemanticEquals(%q,%q) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestIconValueSemanticEqualsNullUnknown(t *testing.T) {
	ctx := context.Background()
	nullV := NewIconNull()
	unknownV := IconValue{StringValue: basetypes.NewStringUnknown()}
	known := NewIconValue("AAAAbbbb==")

	if got, _ := nullV.StringSemanticEquals(ctx, known); got {
		t.Fatalf("null vs known should not be semantically equal")
	}
	if got, _ := unknownV.StringSemanticEquals(ctx, known); got {
		t.Fatalf("unknown vs known should not be semantically equal")
	}
}
