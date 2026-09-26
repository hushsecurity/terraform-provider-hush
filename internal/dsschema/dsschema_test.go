package dsschema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func TestFromResource(t *testing.T) {
	resource := map[string]*schema.Schema{
		"name":      {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.StringIsNotEmpty},
		"enabled":   {Type: schema.TypeBool, Optional: true, Default: true},
		"secret":    {Type: schema.TypeString, Optional: true, Sensitive: true},
		"secret_wo": {Type: schema.TypeString, Optional: true, WriteOnly: true},
		"tags":      {Type: schema.TypeSet, Optional: true, MinItems: 1, Elem: &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty}},
		"labels": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{
			Type: schema.TypeMap, Elem: &schema.Schema{Type: schema.TypeString},
		}},
		"auth": {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"type":   {Type: schema.TypeString, Required: true},
			"secret": {Type: schema.TypeString, Optional: true, Sensitive: true},
		}}},
	}

	got := FromResource(resource, Options{
		Drop:         []string{"secret", "secret_wo", "auth.secret"},
		Descriptions: map[string]string{"auth.type": "read"},
	})

	for _, name := range []string{"secret", "secret_wo"} {
		if _, ok := got[name]; ok {
			t.Errorf("%s was not dropped", name)
		}
	}
	nested := got["auth"].Elem.(*schema.Resource).Schema
	if _, ok := nested["secret"]; ok {
		t.Error("auth.secret was not dropped")
	}
	for path, s := range map[string]*schema.Schema{
		"name": got["name"], "enabled": got["enabled"], "tags": got["tags"], "auth": got["auth"], "auth.type": nested["type"],
	} {
		if !s.Computed || s.Optional || s.Required || s.ForceNew || s.Default != nil ||
			s.ValidateFunc != nil || s.MinItems != 0 || s.MaxItems != 0 {
			t.Errorf("%s is not purely computed: %+v", path, s)
		}
	}
	if elem := got["tags"].Elem.(*schema.Schema); elem.ValidateFunc != nil {
		t.Error("a list element keeps its validation")
	}
	if inner := got["labels"].Elem.(*schema.Schema).Elem; inner == nil {
		t.Error("a collection of collections lost its inner element")
	}
	if nested["type"].Description != "read" {
		t.Errorf("auth.type description %q was not replaced", nested["type"].Description)
	}
	// The source is left as it was.
	if !resource["name"].Required {
		t.Error("the resource schema was modified")
	}

	// The whole derived schema must be one the SDK accepts for a data source.
	Lookup(got, "name", "", "tags")
	if err := (&schema.Resource{Schema: got}).InternalValidate(nil, false); err != nil {
		t.Fatal(err)
	}
}

// A path that names nothing is a mistake, and a secret left in by one is the
// worst way to find it.
func TestFromResourcePanicsOnAPathThatNamesNothing(t *testing.T) {
	for name, opts := range map[string]Options{
		"drop":        {Drop: []string{"auth.secrett"}},
		"description": {Descriptions: map[string]string{"nope": "x"}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected a panic")
				}
			}()
			FromResource(map[string]*schema.Schema{
				"auth": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"secret": {Type: schema.TypeString, Optional: true},
				}}},
			}, opts)
		})
	}
}
