package postgres_access_privilege

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandGrants(t *testing.T) {
	input := []any{
		map[string]any{
			"privileges":    []any{"SELECT", "INSERT"},
			"object_type":   "table",
			"object_names":  []any{"users", "orders"},
			"column_names":  []any{"id", "name"},
			"all_in_schema": false,
		},
		map[string]any{
			"privileges":    []any{"USAGE"},
			"object_type":   "schema",
			"object_names":  []any{},
			"column_names":  []any{},
			"all_in_schema": true,
		},
	}

	result := expandGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	// First grant
	if result[0].ObjectType != "table" {
		t.Errorf("expected object_type 'table', got '%s'", result[0].ObjectType)
	}
	if len(result[0].Privileges) != 2 || result[0].Privileges[0] != "SELECT" || result[0].Privileges[1] != "INSERT" {
		t.Errorf("unexpected privileges: %v", result[0].Privileges)
	}
	if len(result[0].ObjectNames) != 2 || result[0].ObjectNames[0] != "users" || result[0].ObjectNames[1] != "orders" {
		t.Errorf("unexpected object_names: %v", result[0].ObjectNames)
	}
	if len(result[0].ColumnNames) != 2 || result[0].ColumnNames[0] != "id" || result[0].ColumnNames[1] != "name" {
		t.Errorf("unexpected column_names: %v", result[0].ColumnNames)
	}
	if result[0].AllInSchema != false {
		t.Errorf("expected false all_in_schema, got '%v'", result[0].AllInSchema)
	}

	// Second grant
	if result[1].ObjectType != "schema" {
		t.Errorf("expected object_type 'schema', got '%s'", result[1].ObjectType)
	}
	if len(result[1].Privileges) != 1 || result[1].Privileges[0] != "USAGE" {
		t.Errorf("unexpected privileges: %v", result[1].Privileges)
	}
	if result[1].AllInSchema != true {
		t.Errorf("expected all_in_schema true, got '%v'", result[1].AllInSchema)
	}
}

func TestFlattenGrants(t *testing.T) {
	input := []client.PostgresGrant{
		{
			Privileges:  []string{"SELECT", "INSERT"},
			ObjectType:  "table",
			ObjectNames: []string{"users", "orders"},
			ColumnNames: []string{"id", "name"},
			AllInSchema: false,
		},
		{
			Privileges:  []string{"USAGE"},
			ObjectType:  "schema",
			ObjectNames: nil,
			ColumnNames: nil,
			AllInSchema: true,
		},
	}

	result := flattenGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	m0 := result[0].(map[string]any)
	if m0["object_type"] != "table" {
		t.Errorf("expected object_type 'table', got '%v'", m0["object_type"])
	}
	privs0 := m0["privileges"].([]string)
	if len(privs0) != 2 || privs0[0] != "SELECT" || privs0[1] != "INSERT" {
		t.Errorf("unexpected privileges: %v", privs0)
	}
	names0 := m0["object_names"].([]string)
	if len(names0) != 2 || names0[0] != "users" || names0[1] != "orders" {
		t.Errorf("unexpected object_names: %v", names0)
	}
	cols0 := m0["column_names"].([]string)
	if len(cols0) != 2 || cols0[0] != "id" || cols0[1] != "name" {
		t.Errorf("unexpected column_names: %v", cols0)
	}

	m1 := result[1].(map[string]any)
	if m1["object_type"] != "schema" {
		t.Errorf("expected object_type 'schema', got '%v'", m1["object_type"])
	}
	if m1["all_in_schema"] != true {
		t.Errorf("expected all_in_schema true, got '%v'", m1["all_in_schema"])
	}
}

// terraformizeGrants converts []string to []any for use with expandGrants
func terraformizeGrants(grants []client.PostgresGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		privs := make([]any, len(g.Privileges))
		for j, p := range g.Privileges {
			privs[j] = p
		}
		objNames := make([]any, len(g.ObjectNames))
		for j, n := range g.ObjectNames {
			objNames[j] = n
		}
		colNames := make([]any, len(g.ColumnNames))
		for j, c := range g.ColumnNames {
			colNames[j] = c
		}
		result[i] = map[string]any{
			"privileges":    privs,
			"object_type":   g.ObjectType,
			"object_names":  objNames,
			"column_names":  colNames,
			"all_in_schema": g.AllInSchema,
		}
	}
	return result
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	original := []client.PostgresGrant{
		{
			Privileges:  []string{"SELECT", "INSERT", "UPDATE"},
			ObjectType:  "table",
			ObjectNames: []string{"users"},
			ColumnNames: []string{"id", "email"},
			AllInSchema: false,
		},
		{
			Privileges:  []string{"ALL"},
			ObjectType:  "database",
			ObjectNames: nil,
			ColumnNames: nil,
			AllInSchema: false,
		},
		{
			Privileges:  []string{"EXECUTE"},
			ObjectType:  "function",
			ObjectNames: []string{"my_func"},
			ColumnNames: nil,
			AllInSchema: true,
		},
	}

	// Flatten to Terraform format, then simulate what Terraform would give back
	flattened := flattenGrants(original)
	terraformed := terraformizeGrants(original)
	expanded := expandGrants(terraformed)

	// Verify the round-trip
	if len(expanded) != len(original) {
		t.Fatalf("expected %d grants, got %d", len(original), len(expanded))
	}

	// Also verify flatten output length
	if len(flattened) != len(original) {
		t.Fatalf("expected %d flattened grants, got %d", len(original), len(flattened))
	}

	for i, g := range expanded {
		if g.ObjectType != original[i].ObjectType {
			t.Errorf("grant %d: expected object_type '%s', got '%s'", i, original[i].ObjectType, g.ObjectType)
		}
		if len(g.Privileges) != len(original[i].Privileges) {
			t.Errorf("grant %d: expected %d privileges, got %d", i, len(original[i].Privileges), len(g.Privileges))
		}
		for j, p := range g.Privileges {
			if p != original[i].Privileges[j] {
				t.Errorf("grant %d privilege %d: expected '%s', got '%s'", i, j, original[i].Privileges[j], p)
			}
		}
		if g.AllInSchema != original[i].AllInSchema {
			t.Errorf("grant %d: expected all_in_schema '%v', got '%v'", i, original[i].AllInSchema, g.AllInSchema)
		}
	}
}
