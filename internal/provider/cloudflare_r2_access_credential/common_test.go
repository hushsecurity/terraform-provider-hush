package cloudflare_r2_access_credential

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestValidateAPIToken(t *testing.T) {
	cases := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"plain", "AbC-123_xyz.token", false},
		{"trailing newline", "AbC-123\n", true},
		{"inner space", "AbC 123", true},
		{"tab", "AbC\t123", true},
		{"control character", "AbC\x00123", true},
		{"non-breaking space", "AbC\u00a0123", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := validateAPIToken(tc.token, "api_token")
			if (len(errs) > 0) != tc.wantErr {
				t.Fatalf("errs = %v, wantErr = %v", errs, tc.wantErr)
			}
		})
	}
}

// The SDK runs ValidateFunc on a write-only value too, so api_token_wo cannot
// slip a pasted newline past the check.
func TestAPITokenWOIsValidated(t *testing.T) {
	cfg := terraform.NewResourceConfigRaw(map[string]any{
		"name":                 "n",
		"deployment_ids":       []any{"dep-1"},
		"account_id":           "0123456789abcdef0123456789abcdef",
		"api_token_wo":         "AbC-123\n",
		"api_token_wo_version": "1",
	})
	diags := Resource().Validate(cfg)
	if !diags.HasError() {
		t.Fatal("api_token_wo with a trailing newline was accepted")
	}
	if !strings.Contains(diags[0].Summary, "whitespace or control characters") {
		t.Fatalf("unexpected error: %v", diags[0].Summary)
	}
}

type fakeDiff struct {
	id      string
	changed map[string]bool
	unknown map[string]bool
}

func (f fakeDiff) Id() string                  { return f.id }
func (f fakeDiff) HasChange(k string) bool     { return f.changed[k] }
func (f fakeDiff) NewValueKnown(k string) bool { return !f.unknown[k] }

func changed(keys ...string) map[string]bool {
	m := map[string]bool{}
	for _, k := range keys {
		m[k] = true
	}
	return m
}

func TestValidateAccountChange(t *testing.T) {
	cases := []struct {
		name    string
		diff    fakeDiff
		wantErr bool
	}{
		{"create", fakeDiff{changed: changed("account_id")}, false},
		{"account unchanged", fakeDiff{id: "acr-1", changed: changed("name")}, false},
		{"account and api_token", fakeDiff{id: "acr-1", changed: changed("account_id", "api_token")}, false},
		{"account and wo version", fakeDiff{id: "acr-1", changed: changed("account_id", "api_token_wo_version")}, false},
		{"account alone", fakeDiff{id: "acr-1", changed: changed("account_id")}, true},
		{
			"unknown account is left to the API",
			fakeDiff{id: "acr-1", changed: changed("account_id"), unknown: changed("account_id")},
			false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAccountChange(tc.diff)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
