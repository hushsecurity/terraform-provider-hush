package cloudinary_access_credential

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestValidateKeyPart(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr string
	}{
		{"plain", "123456789012345", ""},
		{"secret", "AbC-123_xyz.secret", ""},
		{"256 characters", strings.Repeat("a", 256), ""},
		{"empty", "", "must be 1 to 256 characters"},
		{"257 characters", strings.Repeat("a", 257), "must be 1 to 256 characters"},
		{"trailing newline", "AbC-123\n", "whitespace or control characters"},
		{"inner space", "AbC 123", "whitespace or control characters"},
		{"tab", "AbC\t123", "whitespace or control characters"},
		{"control character", "AbC\x00123", "whitespace or control characters"},
		{"non-breaking space", "AbC 123", "whitespace or control characters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := validateKeyPart(tc.value, "api_secret")
			if tc.wantErr == "" {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			if len(errs) == 0 || !strings.Contains(errs[0].Error(), tc.wantErr) {
				t.Fatalf("errs = %v, want one containing %q", errs, tc.wantErr)
			}
		})
	}
}

func TestAccountIDPattern(t *testing.T) {
	for id, want := range map[string]bool{
		"abcde1fg-2345-6789-0abc-d1efg2hi3jkl": true,
		"acct_1":                               true,
		strings.Repeat("a", 128):               true,
		"":                                     false,
		strings.Repeat("a", 129):               false,
		"acct/1":                               false, // it goes in API paths
		"acct 1":                               false,
	} {
		if got := accountIDRe.MatchString(id); got != want {
			t.Errorf("%q: match = %v, want %v", id, got, want)
		}
	}
}

// The SDK runs ValidateFunc on a write-only value too, so neither write-only
// secret slips a pasted newline past the check.
func TestWriteOnlyKeyPartsAreValidated(t *testing.T) {
	for _, attr := range []string{"api_key_wo", "api_secret_wo"} {
		t.Run(attr, func(t *testing.T) {
			raw := map[string]any{
				"name":                  "n",
				"deployment_ids":        []any{"dep-1"},
				"account_id":            "acct-1",
				"api_key_wo":            "key",
				"api_key_wo_version":    "1",
				"api_secret_wo":         "secret",
				"api_secret_wo_version": "1",
			}
			raw[attr] = "value\n"
			diags := Resource().Validate(terraform.NewResourceConfigRaw(raw))
			if !diags.HasError() {
				t.Fatalf("%s with a trailing newline was accepted", attr)
			}
			if !strings.Contains(diags[0].Summary, "whitespace or control characters") {
				t.Fatalf("unexpected error: %v", diags[0].Summary)
			}
		})
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
		{"account, key and secret", fakeDiff{id: "acr-1", changed: changed("account_id", "api_key", "api_secret")}, false},
		{
			"account and both wo versions",
			fakeDiff{id: "acr-1", changed: changed("account_id", "api_key_wo_version", "api_secret_wo_version")},
			false,
		},
		{"account, key and secret wo version", fakeDiff{
			id: "acr-1", changed: changed("account_id", "api_key", "api_secret_wo_version")}, false},
		{"account alone", fakeDiff{id: "acr-1", changed: changed("account_id")}, true},
		{"account and key only", fakeDiff{id: "acr-1", changed: changed("account_id", "api_key")}, true},
		{"account and secret only", fakeDiff{id: "acr-1", changed: changed("account_id", "api_secret")}, true},
		{"account and key wo version only", fakeDiff{
			id: "acr-1", changed: changed("account_id", "api_key_wo_version")}, true},
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
