package writeonly

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// fakeGetter satisfies rawConfigGetter for tests, since *schema.ResourceData's
// internal rawConfig fields are unexported and TestResourceDataRaw does not
// populate CtyValue.
type fakeGetter struct{ v cty.Value }

func (f fakeGetter) GetRawConfig() cty.Value { return f.v }

func objVal(attrs map[string]cty.Value) cty.Value {
	if attrs == nil {
		return cty.NullVal(cty.Object(map[string]cty.Type{}))
	}
	return cty.ObjectVal(attrs)
}

func TestReadRawString_AttrSet(t *testing.T) {
	g := fakeGetter{v: objVal(map[string]cty.Value{
		"api_key_wo": cty.StringVal("wo-value"),
	})}
	if got := readRawString(g, "api_key_wo"); got != "wo-value" {
		t.Fatalf("got %q, want %q", got, "wo-value")
	}
}

func TestReadRawString_AttrNull(t *testing.T) {
	g := fakeGetter{v: objVal(map[string]cty.Value{
		"api_key_wo": cty.NullVal(cty.String),
	})}
	if got := readRawString(g, "api_key_wo"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestReadRawString_AttrUnknown(t *testing.T) {
	g := fakeGetter{v: objVal(map[string]cty.Value{
		"api_key_wo": cty.UnknownVal(cty.String),
	})}
	if got := readRawString(g, "api_key_wo"); got != "" {
		t.Fatalf("got %q, want empty (unknown should not panic AsString)", got)
	}
}

func TestReadRawString_RawConfigNull(t *testing.T) {
	g := fakeGetter{v: cty.NullVal(cty.Object(map[string]cty.Type{"api_key_wo": cty.String}))}
	if got := readRawString(g, "api_key_wo"); got != "" {
		t.Fatalf("got %q, want empty when raw config is null", got)
	}
}

func TestIsSet_NonEmpty(t *testing.T) {
	g := fakeGetter{v: objVal(map[string]cty.Value{
		"api_key_wo": cty.StringVal("x"),
	})}
	if !IsSet(g, "api_key_wo") {
		t.Fatal("expected IsSet=true")
	}
}

func TestIsSet_EmptyString(t *testing.T) {
	// Mirrors GetOk semantics: empty string counts as unset.
	g := fakeGetter{v: objVal(map[string]cty.Value{
		"api_key_wo": cty.StringVal(""),
	})}
	if IsSet(g, "api_key_wo") {
		t.Fatal("expected IsSet=false for empty string")
	}
}

func TestIsSet_Null(t *testing.T) {
	g := fakeGetter{v: objVal(map[string]cty.Value{
		"api_key_wo": cty.NullVal(cty.String),
	})}
	if IsSet(g, "api_key_wo") {
		t.Fatal("expected IsSet=false for null")
	}
}

func TestGetString_PlainSet_ReturnsPlain(t *testing.T) {
	sch := map[string]*schema.Schema{
		"api_key":    {Type: schema.TypeString, Optional: true},
		"api_key_wo": {Type: schema.TypeString, Optional: true, WriteOnly: true},
	}
	d := schema.TestResourceDataRaw(t, sch, map[string]any{"api_key": "plain-value"})
	if got := GetString(d, "api_key", "api_key_wo"); got != "plain-value" {
		t.Fatalf("got %q, want %q", got, "plain-value")
	}
}

func TestGetString_NeitherSet_ReturnsEmpty(t *testing.T) {
	sch := map[string]*schema.Schema{
		"api_key":    {Type: schema.TypeString, Optional: true},
		"api_key_wo": {Type: schema.TypeString, Optional: true, WriteOnly: true},
	}
	d := schema.TestResourceDataRaw(t, sch, map[string]any{})
	if got := GetString(d, "api_key", "api_key_wo"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
