package secret_store

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

// The acceptance tests cover these through a plan, but need the fixtures
// "make fetch-mock-fixtures" pulls. These are pure functions, so the boundary
// rows are cheaper here. Expectations mirror midgard's validate_prefix.
func TestPrefixValidation(t *testing.T) {
	for _, tc := range []struct {
		kind     string
		validate func(any, string) ([]string, []error)
		prefix   string
		valid    bool
	}{
		// each kind's own punctuation
		{"aws_sm", awsSMPrefixValidation, "acme", true},
		{"aws_sm", awsSMPrefixValidation, "acme-prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme_prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme.prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme+prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme=prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme@prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme/prod", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme_prod", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme.prod", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme/prod/secrets", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme+prod", false},
		{"aws_ssm", awsSSMPrefixValidation, "acme@prod", false},
		{"gcp_sm", gcpSMPrefixValidation, "acme_prod", true},
		{"gcp_sm", gcpSMPrefixValidation, "acme.prod", false},
		{"gcp_sm", gcpSMPrefixValidation, "acme/prod", false},
		{"k8s_secrets", k8sPrefixValidation, "acme.prod", true},
		{"k8s_secrets", k8sPrefixValidation, "acme_prod", false},
		{"k8s_secrets", k8sPrefixValidation, "acme/prod", false},
		{"hc_vault", hcVaultPrefixValidation, "acme_prod", true},
		{"hc_vault", hcVaultPrefixValidation, "acme.prod", true},
		{"hc_vault", hcVaultPrefixValidation, "acme/prod/secrets", true},
		{"hc_vault", hcVaultPrefixValidation, "acme+prod", false},
		{"hc_vault", hcVaultPrefixValidation, "acme@prod", false},

		// adjacency: allowed wherever the charset allows both characters, since
		// the backends attach no meaning to it
		{"aws_sm", awsSMPrefixValidation, "ac__me", true},
		{"aws_sm", awsSMPrefixValidation, "ac--me", true},
		{"aws_sm", awsSMPrefixValidation, "ac_.me", true},
		{"aws_ssm", awsSSMPrefixValidation, "ac__me", true},
		{"gcp_sm", gcpSMPrefixValidation, "ac__me", true},
		{"k8s_secrets", k8sPrefixValidation, "ac.-me", false},
		{"hc_vault", hcVaultPrefixValidation, "ac__me", true},
		{"hc_vault", hcVaultPrefixValidation, "ac.-me", true},

		// "." is the one character that may not repeat, for every kind: it is an
		// empty label in a DNS subdomain and nothing else has a use for it
		{"aws_sm", awsSMPrefixValidation, "ac..me", false},
		{"aws_ssm", awsSSMPrefixValidation, "ac..me", false},
		{"aws_ssm", awsSSMPrefixValidation, "acme/../prod", false},
		{"k8s_secrets", k8sPrefixValidation, "ac..me", false},
		// gcp_sm has no "." in its charset, so it reports that instead
		{"gcp_sm", gcpSMPrefixValidation, "ac..me", false},

		// shape: punctuation never leads or trails
		{"aws_sm", awsSMPrefixValidation, "/acme", false},
		{"aws_sm", awsSMPrefixValidation, "acme/", false},
		{"aws_sm", awsSMPrefixValidation, "acme//prod", false},
		{"gcp_sm", gcpSMPrefixValidation, "-acme", false},
		{"gcp_sm", gcpSMPrefixValidation, "acme-", false},
		{"gcp_sm", gcpSMPrefixValidation, "ac--me", false}, // "-" is its separator
		{"k8s_secrets", k8sPrefixValidation, "", false},
		{"hc_vault", hcVaultPrefixValidation, "/acme", false},
		{"hc_vault", hcVaultPrefixValidation, "acme/", false},
		{"hc_vault", hcVaultPrefixValidation, "acme//prod", false},
		{"hc_vault", hcVaultPrefixValidation, "Acme", false},

		// uppercase: every layer above mufasa is lowercase-only
		{"aws_sm", awsSMPrefixValidation, "Acme", false},
		{"aws_ssm", awsSSMPrefixValidation, "ACME", false},

		// reserved case-insensitively, and only as a prefix
		{"aws_ssm", awsSSMPrefixValidation, "aws", false},
		{"aws_ssm", awsSSMPrefixValidation, "aws-corp", false},
		{"aws_ssm", awsSSMPrefixValidation, "ssmx", false},
		{"aws_ssm", awsSSMPrefixValidation, "awt", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme-aws", true},
		{"aws_sm", awsSMPrefixValidation, "aws-corp", true},
		{"gcp_sm", gcpSMPrefixValidation, "ssmx", true},
		// nothing is reserved for a KV v2 path: the silo writes under
		// <mount>/data/<prefix>/..., so "data" and "metadata" are not special
		{"hc_vault", hcVaultPrefixValidation, "aws-corp", true},
		{"hc_vault", hcVaultPrefixValidation, "data", true},
		{"hc_vault", hcVaultPrefixValidation, "metadata", true},

		// the depth budget, at and past the boundary
		{"aws_ssm", awsSSMPrefixValidation, "a/b/c/d/e/f/g/h/i", true},
		{"aws_ssm", awsSSMPrefixValidation, "a/b/c/d/e/f/g/h/i/j", false},
		{"aws_sm", awsSMPrefixValidation, "a/b/c/d/e/f/g/h/i/j", true},
		// Vault imposes no depth of its own
		{"hc_vault", hcVaultPrefixValidation, "a/b/c/d/e/f/g/h/i/j", true},
	} {
		_, errs := tc.validate(tc.prefix, "prefix")
		if got := len(errs) == 0; got != tc.valid {
			t.Errorf("%s: validate(%q) valid = %v, want %v (%v)",
				tc.kind, tc.prefix, got, tc.valid, errs)
		}
	}
}

// The API refuses a plaintext Vault address, and refusing it at plan time is
// what keeps a customer from discovering it on apply. Mirrors midgard's
// _validate_vault_address.
func TestAddressValidation(t *testing.T) {
	for _, tc := range []struct {
		address string
		valid   bool
	}{
		{"https://vault.acme.internal:8200", true},
		{"https://vault.acme.internal", true},
		// a scheme is case-insensitive, and Go and python lower it alike, so
		// both sides answer the same
		{"HTTPS://vault.acme.internal:8200", true},
		{"http://vault.acme.internal:8200", false},
		{"ftp://vault.acme.internal", false},
		// no scheme at all: the host reads as one, and there is no https
		{"vault.acme.internal:8200", false},
		{"vault.acme.internal", false},
		{"https:///nohost", false},
		{"", false},
		// url.Parse alone accepts these three and midgard refuses them, so
		// without the scan before it they plan cleanly and fail on apply
		{"https://vault.acme.internal:8200/a b", false},
		{"https://vault.acme.internal:8200/?%zz", false},
		{"https://vault acme.internal:8200", false},
		// these two url.Parse already refuses, as control characters
		{"https://vault.acme.internal:8200/a\tb", false},
		{"https://vault.acme.internal:8200/a\nb", false},
		// a well-formed escape is still a valid address
		{"https://vault.acme.internal:8200/a%2Fb", true},
		{"https://vault.acme.internal:8200/a%2fb", true},
		// a "%" at the very end has no two digits after it
		{"https://vault.acme.internal:8200/a%", false},
		{"https://vault.acme.internal:8200/a%2", false},
	} {
		_, errs := addressValidation(tc.address, "address")
		if got := len(errs) == 0; got != tc.valid {
			t.Errorf("addressValidation(%q) valid = %v, want %v (%v)",
				tc.address, got, tc.valid, errs)
		}
	}
}

// Both mounts are interpolated into a request path the access manager builds,
// so the rule has to refuse a mount that walks out of it. Mirrors midgard's
// VAULT_MOUNT, and the two sides have to agree or a store passes the plan and
// is refused on apply with a config that cannot then be edited.
func TestMountValidation(t *testing.T) {
	for _, tc := range []struct {
		mount string
		valid bool
	}{
		{"secret", true},
		{"hush-kv", true},
		{"hush_kv", true},
		{"team/hush-kv", true},
		{"kv2", true},
		{"../sys/mounts", false},
		{"hush kv", false},
		{"hush/", false},
		{"/hush", false},
		{"kv//v2", false},
		{"", false},
	} {
		_, errs := mountValidation(tc.mount, "mount")
		if got := len(errs) == 0; got != tc.valid {
			t.Errorf("mountValidation(%q) valid = %v, want %v (%v)",
				tc.mount, got, tc.valid, errs)
		}
	}
}

// midgard caps every one of these, so a longer value plans cleanly and is
// refused on apply, leaving a config that cannot then be edited. The cap is
// checked at the boundary on both sides.
func TestVaultFieldLengthCaps(t *testing.T) {
	vault := SecretStoreResourceSchema()["hc_vault"].
		Elem.(*schema.Resource).Schema
	auth := vault["auth"].Elem.(*schema.Resource).Schema

	for _, tc := range []struct {
		field    string
		validate func(any, string) ([]string, []error)
		max      int
		// a value of the given length that the field's other rules accept
		fill func(int) string
	}{
		{"address", vault["address"].ValidateFunc, vaultFieldMax, func(n int) string {
			const head = "https://"
			return head + strings.Repeat("v", n-len(head))
		}},
		{"mount", vault["mount"].ValidateFunc, vaultFieldMax, repeatA},
		{"vault_namespace", vault["vault_namespace"].ValidateFunc, vaultFieldMax, repeatA},
		{"ca_cert", vault["ca_cert"].ValidateFunc, vaultCACertMax, repeatA},
		{"auth.mount", auth["mount"].ValidateFunc, vaultFieldMax, repeatA},
		{"auth.role", auth["role"].ValidateFunc, vaultFieldMax, repeatA},
	} {
		t.Run(tc.field, func(t *testing.T) {
			if _, errs := tc.validate(tc.fill(tc.max), tc.field); len(errs) != 0 {
				t.Errorf("%d characters rejected: %v", tc.max, errs)
			}
			if _, errs := tc.validate(tc.fill(tc.max+1), tc.field); len(errs) == 0 {
				t.Errorf("%d characters accepted, want the cap to refuse it", tc.max+1)
			}
		})
	}
}

func repeatA(n int) string { return strings.Repeat("a", n) }

// The vault block is the only one with a nested block of its own, so the two
// halves of its round trip are worth checking directly: expandConfig reads it
// into the client's flattened struct, and setConfigBlocks writes what the API
// answered back into state.
func TestHcVaultConfigRoundTrip(t *testing.T) {
	block := map[string]any{
		"prefix":          "acme/prod",
		"address":         "https://vault.acme.internal:8200",
		"mount":           "hush-kv",
		"vault_namespace": "admin/acme",
		"ca_cert":         "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n",
		"auth": []any{map[string]any{
			"method": client.SecretStoreVaultAuthKubernetes,
			"mount":  "kubernetes-eu",
			"role":   "hush-am",
		}},
	}
	d := schema.TestResourceDataRaw(t, SecretStoreResourceSchema(), map[string]any{
		"name":     "vault-store",
		"hc_vault": []any{block},
	})

	config, err := expandConfig(d)
	if err != nil {
		t.Fatalf("expandConfig: %v", err)
	}
	if config.Kind != client.SecretStoreKindHCVault {
		t.Errorf("kind = %q, want %q", config.Kind, client.SecretStoreKindHCVault)
	}
	for _, tc := range []struct{ name, got, want string }{
		{"prefix", config.Prefix, "acme/prod"},
		{"address", config.Address, "https://vault.acme.internal:8200"},
		{"mount", config.Mount, "hush-kv"},
		{"vault_namespace", config.VaultNamespace, "admin/acme"},
		{"auth.role", config.Auth.Role, "hush-am"},
		{"auth.mount", config.Auth.Mount, "kubernetes-eu"},
		{"auth.method", config.Auth.Method, client.SecretStoreVaultAuthKubernetes},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
	// no other kind's field rides along, which the API's strict model refuses
	if config.Region != "" || config.ProjectID != "" || config.Namespace != "" {
		t.Errorf("another kind's fields are set: %+v", config)
	}

	// and back into state
	fresh := schema.TestResourceDataRaw(t, SecretStoreResourceSchema(), map[string]any{})
	if err := setConfigBlocks(fresh, config); err != nil {
		t.Fatalf("setConfigBlocks: %v", err)
	}
	stored := fresh.Get("hc_vault").([]any)
	if len(stored) != 1 {
		t.Fatalf("hc_vault = %v, want one block", stored)
	}
	got := stored[0].(map[string]any)
	if got["address"] != "https://vault.acme.internal:8200" {
		t.Errorf("address = %v", got["address"])
	}
	if got["ca_cert"] != block["ca_cert"] {
		t.Errorf("ca_cert = %v", got["ca_cert"])
	}
	auth := got["auth"].([]any)[0].(map[string]any)
	if auth["role"] != "hush-am" || auth["mount"] != "kubernetes-eu" {
		t.Errorf("auth = %v", auth)
	}
	// the other blocks are cleared, so state shows only the active backend
	for _, name := range []string{"aws_sm", "aws_ssm", "gcp_sm", "k8s_secrets"} {
		if v := fresh.Get(name).([]any); len(v) != 0 {
			t.Errorf("%s = %v, want empty", name, v)
		}
	}
}

// The value Terraform gives a field it cannot resolve yet. An acceptance test
// cannot express it, so the unknown rows below go through Resource().Diff.
const unknownValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

// vaultAuthDiff runs the real resource diff, CustomizeDiff included, for a new
// store whose auth block is the one given.
func vaultAuthDiff(t *testing.T, auth map[string]any) error {
	t.Helper()
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name":           "s",
		"deployment_ids": []any{"dep-xxxxxxxxxxxxxxxx"},
		"hc_vault": []any{map[string]any{
			"prefix":  "hush",
			"address": "https://vault.example.internal:8200",
			"auth":    []any{auth},
		}},
	})
	_, err := Resource().Diff(context.Background(), nil, config, nil)
	return err
}

// The acceptance tests drive these through a plan, which is what a customer
// meets; these pin the messages and cover every branch in one fast table.
func TestValidateVaultAuth(t *testing.T) {
	for _, tc := range []struct {
		name    string
		auth    map[string]any
		errPart string
	}{
		{
			name: "kubernetes with a role",
			auth: map[string]any{"method": "kubernetes", "role": "hush-am"},
		},
		{
			name: "kubernetes with a mount",
			auth: map[string]any{
				"method": "kubernetes", "role": "hush-am", "mount": "k8s-eu",
			},
		},
		{
			name: "jwt with a role",
			auth: map[string]any{"method": "jwt", "role": "hush-am"},
		},
		{
			name: "token names nothing",
			auth: map[string]any{"method": "token"},
		},
		{
			name:    "kubernetes without a role",
			auth:    map[string]any{"method": "kubernetes"},
			errPart: `auth method "kubernetes" needs a role`,
		},
		{
			name:    "jwt without a role",
			auth:    map[string]any{"method": "jwt"},
			errPart: `auth method "jwt" needs a role`,
		},
		// a field of another method is refused rather than ignored: the config
		// is immutable, so it cannot be corrected after the fact
		{
			name:    "token with a role",
			auth:    map[string]any{"method": "token", "role": "hush-am"},
			errPart: `auth method "token" does not use role`,
		},
		{
			name:    "token with a mount",
			auth:    map[string]any{"method": "token", "mount": "kubernetes"},
			errPart: `auth method "token" does not use mount`,
		},
		// A role another resource supplies reads as "" here but is set by the
		// time apply runs, so judging it would fail a plan that is fine.
		// role = vault_kubernetes_auth_backend_role.hush.role_name is the
		// ordinary way to write this.
		{
			name: "an unknown role is left to apply",
			auth: map[string]any{"method": "kubernetes", "role": unknownValue},
		},
		{
			name: "an unknown method is left to apply",
			auth: map[string]any{"method": unknownValue},
		},
		{
			name: "an unknown role on an unknown method",
			auth: map[string]any{"method": unknownValue, "role": unknownValue},
		},
		// token still refuses what it does not use, once the value is known
		{
			name:    "token with an unknown role",
			auth:    map[string]any{"method": "token", "role": unknownValue},
			errPart: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := vaultAuthDiff(t, tc.auth)
			switch {
			case tc.errPart == "" && err != nil:
				t.Errorf("unexpected error: %v", err)
			case tc.errPart != "" && err == nil:
				t.Errorf("expected an error containing %q", tc.errPart)
			case tc.errPart != "" && !strings.Contains(err.Error(), tc.errPart):
				t.Errorf("error = %v, want it to contain %q", err, tc.errPart)
			}
		})
	}
}
