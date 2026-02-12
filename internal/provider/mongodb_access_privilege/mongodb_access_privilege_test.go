package mongodb_access_privilege

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandGrants(t *testing.T) {
	input := []any{
		map[string]any{
			"privileges":     []any{"find", "insert"},
			"resource_type":  "database",
			"resource_names": []any{"mydb", "testdb"},
		},
		map[string]any{
			"privileges":     []any{"find"},
			"resource_type":  "collection",
			"resource_names": []any{"users"},
		},
	}

	result := expandGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	// First grant
	if result[0].ResourceType != "database" {
		t.Errorf("expected resource_type 'database', got '%s'", result[0].ResourceType)
	}
	if len(result[0].Privileges) != 2 || result[0].Privileges[0] != "find" || result[0].Privileges[1] != "insert" {
		t.Errorf("unexpected privileges: %v", result[0].Privileges)
	}
	if len(result[0].ResourceNames) != 2 || result[0].ResourceNames[0] != "mydb" || result[0].ResourceNames[1] != "testdb" {
		t.Errorf("unexpected resource_names: %v", result[0].ResourceNames)
	}

	// Second grant
	if result[1].ResourceType != "collection" {
		t.Errorf("expected resource_type 'collection', got '%s'", result[1].ResourceType)
	}
	if len(result[1].Privileges) != 1 || result[1].Privileges[0] != "find" {
		t.Errorf("unexpected privileges: %v", result[1].Privileges)
	}
	if len(result[1].ResourceNames) != 1 || result[1].ResourceNames[0] != "users" {
		t.Errorf("unexpected resource_names: %v", result[1].ResourceNames)
	}
}

func TestFlattenGrants(t *testing.T) {
	input := []client.MongoDBGrant{
		{
			Privileges:    []string{"find", "insert"},
			ResourceType:  "database",
			ResourceNames: []string{"mydb", "testdb"},
		},
		{
			Privileges:    []string{"find"},
			ResourceType:  "collection",
			ResourceNames: []string{"users"},
		},
	}

	result := flattenGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	m0 := result[0].(map[string]any)
	if m0["resource_type"] != "database" {
		t.Errorf("expected resource_type 'database', got '%v'", m0["resource_type"])
	}
	privs0 := m0["privileges"].([]string)
	if len(privs0) != 2 || privs0[0] != "find" || privs0[1] != "insert" {
		t.Errorf("unexpected privileges: %v", privs0)
	}
	names0 := m0["resource_names"].([]string)
	if len(names0) != 2 || names0[0] != "mydb" || names0[1] != "testdb" {
		t.Errorf("unexpected resource_names: %v", names0)
	}

	m1 := result[1].(map[string]any)
	if m1["resource_type"] != "collection" {
		t.Errorf("expected resource_type 'collection', got '%v'", m1["resource_type"])
	}
}

// terraformizeGrants converts []string to []any for use with expandGrants
func terraformizeGrants(grants []client.MongoDBGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		privs := make([]any, len(g.Privileges))
		for j, p := range g.Privileges {
			privs[j] = p
		}
		resNames := make([]any, len(g.ResourceNames))
		for j, n := range g.ResourceNames {
			resNames[j] = n
		}
		result[i] = map[string]any{
			"privileges":     privs,
			"resource_type":  g.ResourceType,
			"resource_names": resNames,
		}
	}
	return result
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	original := []client.MongoDBGrant{
		{
			Privileges:    []string{"find", "insert", "update"},
			ResourceType:  "database",
			ResourceNames: []string{"mydb"},
		},
		{
			Privileges:    []string{"find"},
			ResourceType:  "collection",
			ResourceNames: []string{"users", "orders"},
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
		if g.ResourceType != original[i].ResourceType {
			t.Errorf("grant %d: expected resource_type '%s', got '%s'", i, original[i].ResourceType, g.ResourceType)
		}
		if len(g.Privileges) != len(original[i].Privileges) {
			t.Errorf("grant %d: expected %d privileges, got %d", i, len(original[i].Privileges), len(g.Privileges))
		}
		for j, p := range g.Privileges {
			if p != original[i].Privileges[j] {
				t.Errorf("grant %d privilege %d: expected '%s', got '%s'", i, j, original[i].Privileges[j], p)
			}
		}
		if len(g.ResourceNames) != len(original[i].ResourceNames) {
			t.Errorf("grant %d: expected %d resource_names, got %d", i, len(original[i].ResourceNames), len(g.ResourceNames))
		}
		for j, n := range g.ResourceNames {
			if n != original[i].ResourceNames[j] {
				t.Errorf("grant %d resource_name %d: expected '%s', got '%s'", i, j, original[i].ResourceNames[j], n)
			}
		}
	}
}
