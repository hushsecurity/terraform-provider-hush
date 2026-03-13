package redis_access_privilege

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandGrants(t *testing.T) {
	input := []any{
		map[string]any{
			"type":   "category",
			"action": "include",
			"name":   "read",
		},
		map[string]any{
			"type":   "command",
			"action": "exclude",
			"name":   "del",
		},
	}

	result := expandGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	if result[0].Type != "category" || result[0].Action != "include" || result[0].Name != "read" {
		t.Errorf("unexpected first grant: %+v", result[0])
	}
	if result[1].Type != "command" || result[1].Action != "exclude" || result[1].Name != "del" {
		t.Errorf("unexpected second grant: %+v", result[1])
	}
}

func TestFlattenGrants(t *testing.T) {
	input := []client.RedisGrant{
		{Type: "category", Action: "include", Name: "read"},
		{Type: "category", Action: "include", Name: "write"},
		{Type: "command", Action: "exclude", Name: "flushall"},
	}

	result := flattenGrants(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 grants, got %d", len(result))
	}

	m0 := result[0].(map[string]any)
	if m0["type"] != "category" || m0["action"] != "include" || m0["name"] != "read" {
		t.Errorf("unexpected first grant: %v", m0)
	}

	m2 := result[2].(map[string]any)
	if m2["type"] != "command" || m2["action"] != "exclude" || m2["name"] != "flushall" {
		t.Errorf("unexpected third grant: %v", m2)
	}
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	original := []client.RedisGrant{
		{Type: "category", Action: "include", Name: "read"},
		{Type: "category", Action: "include", Name: "write"},
		{Type: "command", Action: "exclude", Name: "flushall"},
	}

	terraformed := make([]any, len(original))
	for i, g := range original {
		terraformed[i] = map[string]any{
			"type":   g.Type,
			"action": g.Action,
			"name":   g.Name,
		}
	}

	expanded := expandGrants(terraformed)

	if len(expanded) != len(original) {
		t.Fatalf("expected %d grants, got %d", len(original), len(expanded))
	}

	for i, g := range expanded {
		if g.Type != original[i].Type {
			t.Errorf("grant %d: expected type '%s', got '%s'", i, original[i].Type, g.Type)
		}
		if g.Action != original[i].Action {
			t.Errorf("grant %d: expected action '%s', got '%s'", i, original[i].Action, g.Action)
		}
		if g.Name != original[i].Name {
			t.Errorf("grant %d: expected name '%s', got '%s'", i, original[i].Name, g.Name)
		}
	}
}
