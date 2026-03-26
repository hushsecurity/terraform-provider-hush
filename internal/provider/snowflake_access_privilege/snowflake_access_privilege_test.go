package snowflake_access_privilege

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandGrants(t *testing.T) {
	input := []any{
		map[string]any{
			"privileges":     []any{"USAGE"},
			"resource_type":  "database",
			"resource_names": []any{},
		},
		map[string]any{
			"privileges":     []any{"SELECT", "INSERT"},
			"resource_type":  "table",
			"resource_names": []any{"users", "orders"},
		},
		map[string]any{
			"privileges":     []any{"USAGE"},
			"resource_type":  "warehouse",
			"resource_names": []any{},
		},
	}

	result := expandGrants(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 grants, got %d", len(result))
	}

	// First grant: database USAGE
	if result[0].ResourceType != "database" {
		t.Errorf("expected resource_type 'database', got '%s'", result[0].ResourceType)
	}
	if len(result[0].Privileges) != 1 || result[0].Privileges[0] != "USAGE" {
		t.Errorf("expected privileges [USAGE], got %v", result[0].Privileges)
	}
	if result[0].ResourceNames != nil {
		t.Errorf("expected nil resource_names, got %v", result[0].ResourceNames)
	}

	// Second grant: table SELECT, INSERT on specific tables
	if result[1].ResourceType != "table" {
		t.Errorf("expected resource_type 'table', got '%s'", result[1].ResourceType)
	}
	if len(result[1].Privileges) != 2 {
		t.Errorf("expected 2 privileges, got %d", len(result[1].Privileges))
	}
	if len(result[1].ResourceNames) != 2 || result[1].ResourceNames[0] != "users" {
		t.Errorf("expected resource_names [users, orders], got %v", result[1].ResourceNames)
	}
}

func TestFlattenGrants(t *testing.T) {
	grants := []client.SnowflakeGrant{
		{
			Privileges:   []string{"USAGE"},
			ResourceType: "database",
		},
		{
			Privileges:    []string{"SELECT"},
			ResourceType:  "table",
			ResourceNames: []string{"users"},
		},
	}

	result := flattenGrants(grants)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	m0 := result[0].(map[string]any)
	if m0["resource_type"] != "database" {
		t.Errorf("expected resource_type 'database', got '%s'", m0["resource_type"])
	}
	privs0 := m0["privileges"].([]string)
	if len(privs0) != 1 || privs0[0] != "USAGE" {
		t.Errorf("expected privileges [USAGE], got %v", privs0)
	}
	names0 := m0["resource_names"].([]string)
	if len(names0) != 0 {
		t.Errorf("expected empty resource_names, got %v", names0)
	}

	m1 := result[1].(map[string]any)
	names1 := m1["resource_names"].([]string)
	if len(names1) != 1 || names1[0] != "users" {
		t.Errorf("expected resource_names [users], got %v", names1)
	}
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	input := []any{
		map[string]any{
			"privileges":     []any{"USAGE"},
			"resource_type":  "database",
			"resource_names": []any{},
		},
		map[string]any{
			"privileges":     []any{"USAGE"},
			"resource_type":  "schema",
			"resource_names": []any{},
		},
		map[string]any{
			"privileges":     []any{"SELECT", "INSERT", "UPDATE"},
			"resource_type":  "table",
			"resource_names": []any{"users", "orders"},
		},
		map[string]any{
			"privileges":     []any{"USAGE"},
			"resource_type":  "warehouse",
			"resource_names": []any{},
		},
	}

	grants := expandGrants(input)
	result := flattenGrants(grants)

	if len(result) != len(input) {
		t.Fatalf("expected %d grants, got %d", len(input), len(result))
	}

	// Verify table grant round-tripped correctly
	m := result[2].(map[string]any)
	if m["resource_type"] != "table" {
		t.Errorf("expected resource_type 'table', got '%s'", m["resource_type"])
	}
	privs := m["privileges"].([]string)
	if len(privs) != 3 {
		t.Errorf("expected 3 privileges, got %d", len(privs))
	}
	names := m["resource_names"].([]string)
	if len(names) != 2 {
		t.Errorf("expected 2 resource_names, got %d", len(names))
	}
}
