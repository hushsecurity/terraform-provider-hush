package openai_access_privilege

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandPermissions(t *testing.T) {
	input := []any{
		map[string]any{
			"name":  "models",
			"level": "read",
		},
		map[string]any{
			"name":  "assistants",
			"level": "write",
		},
	}

	result := expandPermissions(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(result))
	}

	if result[0].Name != "models" {
		t.Errorf("expected name 'models', got '%s'", result[0].Name)
	}
	if result[0].Level != "read" {
		t.Errorf("expected level 'read', got '%s'", result[0].Level)
	}

	if result[1].Name != "assistants" {
		t.Errorf("expected name 'assistants', got '%s'", result[1].Name)
	}
	if result[1].Level != "write" {
		t.Errorf("expected level 'write', got '%s'", result[1].Level)
	}
}

func TestFlattenPermissions(t *testing.T) {
	input := []client.OpenAIPermission{
		{
			Name:  "models",
			Level: "read",
		},
		{
			Name:  "assistants",
			Level: "write",
		},
	}

	result := flattenPermissions(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(result))
	}

	m0 := result[0].(map[string]any)
	if m0["name"] != "models" {
		t.Errorf("expected name 'models', got '%v'", m0["name"])
	}
	if m0["level"] != "read" {
		t.Errorf("expected level 'read', got '%v'", m0["level"])
	}

	m1 := result[1].(map[string]any)
	if m1["name"] != "assistants" {
		t.Errorf("expected name 'assistants', got '%v'", m1["name"])
	}
	if m1["level"] != "write" {
		t.Errorf("expected level 'write', got '%v'", m1["level"])
	}
}

// terraformizePermissions converts OpenAIPermission structs to []any for use with expandPermissions
func terraformizePermissions(perms []client.OpenAIPermission) []any {
	result := make([]any, len(perms))
	for i, p := range perms {
		result[i] = map[string]any{
			"name":  p.Name,
			"level": p.Level,
		}
	}
	return result
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	original := []client.OpenAIPermission{
		{
			Name:  "models",
			Level: "read",
		},
		{
			Name:  "assistants",
			Level: "write",
		},
		{
			Name:  "files",
			Level: "read",
		},
	}

	// Flatten to Terraform format, then simulate what Terraform would give back
	flattened := flattenPermissions(original)
	terraformed := terraformizePermissions(original)
	expanded := expandPermissions(terraformed)

	// Verify the round-trip
	if len(expanded) != len(original) {
		t.Fatalf("expected %d permissions, got %d", len(original), len(expanded))
	}

	// Also verify flatten output length
	if len(flattened) != len(original) {
		t.Fatalf("expected %d flattened permissions, got %d", len(original), len(flattened))
	}

	for i, p := range expanded {
		if p.Name != original[i].Name {
			t.Errorf("permission %d: expected name '%s', got '%s'", i, original[i].Name, p.Name)
		}
		if p.Level != original[i].Level {
			t.Errorf("permission %d: expected level '%s', got '%s'", i, original[i].Level, p.Level)
		}
	}
}
