package deployment

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func resourceDataFor(t *testing.T, raw map[string]any) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, DeploymentResourceSchema(), raw)
}

// A deployment spanning two clusters is expressed as two blocks, so order and
// per-entry fields have to survive expansion intact.
func TestExpandOidcProviders(t *testing.T) {
	d := resourceDataFor(t, map[string]any{
		"oidc_provider": []any{
			map[string]any{
				"issuer":           "https://blue.example.com",
				"audience":         "hush",
				"allowed_subjects": []any{"system:serviceaccount:hush-security:*"},
			},
			map[string]any{
				"issuer":   "https://green.example.com",
				"audience": "hush",
			},
		},
	})

	got := expandOidcProviders(d)
	if len(got) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(got))
	}
	if got[0].Issuer != "https://blue.example.com" {
		t.Fatalf("unexpected first issuer %q", got[0].Issuer)
	}
	if got[1].Issuer != "https://green.example.com" {
		t.Fatalf("unexpected second issuer %q", got[1].Issuer)
	}
	if len(got[0].AllowedSubjects) != 1 {
		t.Fatalf("expected the first entry to keep its subjects, got %v",
			got[0].AllowedSubjects)
	}
	if got[1].AllowedSubjects != nil {
		t.Fatalf("expected no subjects on the second entry, got %v",
			got[1].AllowedSubjects)
	}
}

// No block must expand to nil rather than an empty list, so an update that did
// not touch the field sends nothing and one that emptied it sends null.
func TestExpandOidcProvidersAbsent(t *testing.T) {
	d := resourceDataFor(t, map[string]any{})
	if got := expandOidcProviders(d); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// The list is preferred on read.
func TestFlattenOidcProvidersFromList(t *testing.T) {
	got := flattenOidcProviders(&client.Deployment{
		OidcProviders: []client.OidcConfig{
			{Issuer: "https://blue.example.com", Audience: "hush"},
			{Issuer: "https://green.example.com", Audience: "hush"},
		},
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0]["issuer"] != "https://blue.example.com" {
		t.Fatalf("unexpected first issuer %v", got[0]["issuer"])
	}
}

// A deployment this provider has not written since the list was introduced
// still answers through the singular field, and has to read back as one block
// so a configuration declaring one produces no diff.
func TestFlattenOidcProvidersFallsBackToSingular(t *testing.T) {
	got := flattenOidcProviders(&client.Deployment{
		OidcProvider: &client.OidcConfig{
			Issuer:          "https://legacy.example.com",
			Audience:        "hush",
			AllowedSubjects: []string{"system:serviceaccount:hush-security:*"},
		},
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0]["issuer"] != "https://legacy.example.com" {
		t.Fatalf("unexpected issuer %v", got[0]["issuer"])
	}
}

// Neither field set must flatten to an empty list rather than to null, or a
// deployment with no OIDC would show a permanent diff.
func TestFlattenOidcProvidersEmpty(t *testing.T) {
	got := flattenOidcProviders(&client.Deployment{})
	if got == nil {
		t.Fatal("expected an empty list, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected no entries, got %d", len(got))
	}
}

// The block is the only OIDC surface the provider offers, and it is capped at
// what the API accepts for the field it is stored in.
func TestOidcProviderBlockCap(t *testing.T) {
	s := DeploymentResourceSchema()

	block, ok := s["oidc_provider"]
	if !ok {
		t.Fatal("oidc_provider missing from the resource schema")
	}
	if block.MaxItems != maxOidcProviders {
		t.Fatalf("expected the block capped at %d, got %d",
			maxOidcProviders, block.MaxItems)
	}
	if _, exists := s["oidc_providers"]; exists {
		t.Fatal("oidc_providers must not be a separate field")
	}
	if block.Deprecated != "" {
		t.Fatalf("oidc_provider is the surviving field, not deprecated: %q",
			block.Deprecated)
	}
}
