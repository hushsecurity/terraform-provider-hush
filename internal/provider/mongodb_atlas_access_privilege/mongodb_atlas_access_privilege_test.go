package mongodb_atlas_access_privilege

import (
	"reflect"
	"testing"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandGrants(t *testing.T) {
	input := []any{
		map[string]any{
			"privileges":     []any{"FIND", "INSERT"},
			"resource_type":  "collection",
			"resource_names": []any{"users", "orders"},
		},
		map[string]any{
			"privileges":     []any{"all"},
			"resource_type":  "database",
			"resource_names": []any{},
		},
	}

	result := expandGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}
	if !reflect.DeepEqual(result[0].Privileges, []string{"FIND", "INSERT"}) {
		t.Errorf("grant 0 privileges = %v", result[0].Privileges)
	}
	if result[0].ResourceType != "collection" {
		t.Errorf("grant 0 resource_type = %q", result[0].ResourceType)
	}
	if !reflect.DeepEqual(result[0].ResourceNames, []string{"users", "orders"}) {
		t.Errorf("grant 0 resource_names = %v", result[0].ResourceNames)
	}
	// Empty resource_names list should leave ResourceNames nil (omitempty).
	if result[1].ResourceNames != nil {
		t.Errorf("grant 1 resource_names = %v, want nil", result[1].ResourceNames)
	}
}

func TestFlattenGrants(t *testing.T) {
	input := []client.MongoDBAtlasGrant{
		{Privileges: []string{"FIND"}, ResourceType: "collection", ResourceNames: []string{"users"}},
		{Privileges: []string{"DROP_DATABASE"}, ResourceType: "database"},
	}

	result := flattenGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}
	m0 := result[0].(map[string]any)
	if !reflect.DeepEqual(m0["privileges"], []string{"FIND"}) {
		t.Errorf("grant 0 privileges = %v", m0["privileges"])
	}
	if m0["resource_type"] != "collection" {
		t.Errorf("grant 0 resource_type = %v", m0["resource_type"])
	}
	if !reflect.DeepEqual(m0["resource_names"], []string{"users"}) {
		t.Errorf("grant 0 resource_names = %v", m0["resource_names"])
	}
	// nil ResourceNames flattens to an empty slice (never nil) for stable state.
	m1 := result[1].(map[string]any)
	if !reflect.DeepEqual(m1["resource_names"], []string{}) {
		t.Errorf("grant 1 resource_names = %v, want []", m1["resource_names"])
	}
}

// terraformizeGrants converts client grants back to the []any shape that
// Terraform's schema decoder produces, so we can round-trip through expand.
func terraformizeGrants(grants []client.MongoDBAtlasGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		names := make([]any, len(g.ResourceNames))
		for j, n := range g.ResourceNames {
			names[j] = n
		}
		privs := make([]any, len(g.Privileges))
		for j, p := range g.Privileges {
			privs[j] = p
		}
		result[i] = map[string]any{
			"privileges":     privs,
			"resource_type":  g.ResourceType,
			"resource_names": names,
		}
	}
	return result
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	original := []client.MongoDBAtlasGrant{
		{Privileges: []string{"FIND", "INSERT"}, ResourceType: "collection", ResourceNames: []string{"users"}},
		{Privileges: []string{"all"}, ResourceType: "database"},
	}

	expanded := expandGrants(terraformizeGrants(original))

	if len(expanded) != len(original) {
		t.Fatalf("expanded len = %d, want %d", len(expanded), len(original))
	}
	for i, g := range expanded {
		if !reflect.DeepEqual(g.Privileges, original[i].Privileges) {
			t.Errorf("grant %d privileges = %v, want %v", i, g.Privileges, original[i].Privileges)
		}
		if g.ResourceType != original[i].ResourceType {
			t.Errorf("grant %d resource_type = %q, want %q", i, g.ResourceType, original[i].ResourceType)
		}
	}
}

func TestExpandGrants_Empty(t *testing.T) {
	if result := expandGrants([]any{}); len(result) != 0 {
		t.Fatalf("expected 0 grants, got %d", len(result))
	}
}

func TestFlattenGrants_Empty(t *testing.T) {
	if result := flattenGrants([]client.MongoDBAtlasGrant{}); len(result) != 0 {
		t.Fatalf("expected 0 grants, got %d", len(result))
	}
}

func TestCheckGrantsOrBuiltinRoles(t *testing.T) {
	grant := cty.ObjectVal(map[string]cty.Value{
		"resource_type": cty.StringVal("collection"),
	})
	grants := cty.ListVal([]cty.Value{grant})
	noGrants := cty.ListValEmpty(grant.Type())
	roles := cty.SetVal([]cty.Value{cty.StringVal("clusterMonitor")})
	noRoles := cty.SetValEmpty(cty.String)

	tests := []struct {
		name          string
		grants, roles cty.Value
		wantErr       bool
	}{
		{"grants only", grants, cty.NullVal(cty.Set(cty.String)), false},
		{"built-in roles only", noGrants, roles, false},
		{"neither, roles null", noGrants, cty.NullVal(cty.Set(cty.String)), true},
		{"neither, roles empty", noGrants, noRoles, true},
		{"grants unknown", cty.UnknownVal(grants.Type()), noRoles, false},
		{"roles unknown", noGrants, cty.UnknownVal(cty.Set(cty.String)), false},
		{
			"roles hold an unknown element",
			noGrants,
			cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkGrantsOrBuiltinRoles(tt.grants, tt.roles)
			if tt.wantErr {
				if err == nil || err.Error() != "name at least one grant or built-in role" {
					t.Fatalf("err = %v, want the grant-or-role error", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
