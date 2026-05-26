package temporal_cloud_access_privilege

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandGrants(t *testing.T) {
	input := []any{
		map[string]any{
			"namespace":  "prod.acct1",
			"permission": "read",
		},
		map[string]any{
			"namespace":  "staging.acct1",
			"permission": "write",
		},
		map[string]any{
			"namespace":  "dev.acct1",
			"permission": "admin",
		},
	}

	result := expandGrants(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 grants, got %d", len(result))
	}

	cases := []struct {
		idx              int
		wantNS, wantPerm string
	}{
		{0, "prod.acct1", "read"},
		{1, "staging.acct1", "write"},
		{2, "dev.acct1", "admin"},
	}
	for _, c := range cases {
		if result[c.idx].Namespace != c.wantNS {
			t.Errorf("grant %d: namespace = %q, want %q", c.idx, result[c.idx].Namespace, c.wantNS)
		}
		if result[c.idx].Permission != c.wantPerm {
			t.Errorf("grant %d: permission = %q, want %q", c.idx, result[c.idx].Permission, c.wantPerm)
		}
	}
}

func TestFlattenGrants(t *testing.T) {
	input := []client.TemporalCloudGrant{
		{Namespace: "prod.acct1", Permission: "read"},
		{Namespace: "staging.acct1", Permission: "write"},
	}

	result := flattenGrants(input)

	if len(result) != 2 {
		t.Fatalf("expected 2 grants, got %d", len(result))
	}

	m0 := result[0].(map[string]any)
	if m0["namespace"] != "prod.acct1" {
		t.Errorf("grant 0: namespace = %v, want %q", m0["namespace"], "prod.acct1")
	}
	if m0["permission"] != "read" {
		t.Errorf("grant 0: permission = %v, want %q", m0["permission"], "read")
	}

	m1 := result[1].(map[string]any)
	if m1["namespace"] != "staging.acct1" {
		t.Errorf("grant 1: namespace = %v, want %q", m1["namespace"], "staging.acct1")
	}
	if m1["permission"] != "write" {
		t.Errorf("grant 1: permission = %v, want %q", m1["permission"], "write")
	}
}

// terraformizeGrants converts client grants back to the []any shape that
// Terraform's schema decoder produces, so we can round-trip through expand.
func terraformizeGrants(grants []client.TemporalCloudGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		result[i] = map[string]any{
			"namespace":  g.Namespace,
			"permission": g.Permission,
		}
	}
	return result
}

func TestExpandFlattenRoundTrip(t *testing.T) {
	original := []client.TemporalCloudGrant{
		{Namespace: "prod.acct1", Permission: "read"},
		{Namespace: "staging.acct1", Permission: "write"},
		{Namespace: "dev.acct1", Permission: "admin"},
	}

	flattened := flattenGrants(original)
	expanded := expandGrants(terraformizeGrants(original))

	if len(flattened) != len(original) {
		t.Fatalf("flattened len = %d, want %d", len(flattened), len(original))
	}
	if len(expanded) != len(original) {
		t.Fatalf("expanded len = %d, want %d", len(expanded), len(original))
	}

	for i, g := range expanded {
		if g.Namespace != original[i].Namespace {
			t.Errorf("grant %d: namespace = %q, want %q", i, g.Namespace, original[i].Namespace)
		}
		if g.Permission != original[i].Permission {
			t.Errorf("grant %d: permission = %q, want %q", i, g.Permission, original[i].Permission)
		}
	}
}

func TestExpandGrants_Empty(t *testing.T) {
	result := expandGrants([]any{})
	if len(result) != 0 {
		t.Fatalf("expected 0 grants, got %d", len(result))
	}
}

func TestFlattenGrants_Empty(t *testing.T) {
	result := flattenGrants([]client.TemporalCloudGrant{})
	if len(result) != 0 {
		t.Fatalf("expected 0 grants, got %d", len(result))
	}
}
