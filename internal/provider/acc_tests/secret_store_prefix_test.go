package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// what a vault block needs beyond the prefix
const vaultBody = `address = "https://vault.acme.internal:8200"
    auth { role = "hush-am" }`

// vaultAuth builds a vault block body whose auth block carries exactly what
// the case is about.
func vaultAuth(auth string) string {
	return fmt.Sprintf(`address = "https://vault.acme.internal:8200"
    auth {
      %s
    }`, auth)
}

// Each config block wires its own validator, so every block needs its own
// case: a charset legal for one backend is illegal for another.
func TestAccResourceSecretStorePrefixCharset(t *testing.T) {
	tests := []struct {
		name      string
		block     string
		body      string
		prefix    string
		expectErr *regexp.Regexp
	}{
		{
			name:   "aws_sm.plus.and.at.allowed",
			block:  "aws_sm",
			body:   `region = "eu-west-1"`,
			prefix: "acme+corp@prod",
		},
		{
			name:      "aws_ssm.rejects.plus",
			block:     "aws_ssm",
			body:      `region = "eu-west-1"`,
			prefix:    "acme+corp",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		{
			name:   "aws_ssm.slash.delimits",
			block:  "aws_ssm",
			body:   `region = "eu-west-1"`,
			prefix: "acme/prod/secrets",
		},
		{
			name:      "aws_ssm.rejects.reserved.aws",
			block:     "aws_ssm",
			body:      `region = "eu-west-1"`,
			prefix:    "awsdev",
			expectErr: regexp.MustCompile(`reserves`),
		},
		{
			name:      "aws_ssm.rejects.reserved.ssm",
			block:     "aws_ssm",
			body:      `region = "eu-west-1"`,
			prefix:    "ssmprod",
			expectErr: regexp.MustCompile(`reserves`),
		},
		{
			name:   "gcp_sm.underscore.allowed",
			block:  "gcp_sm",
			body:   `project_id = "a-project"`,
			prefix: "acme_prod",
		},
		{
			name:      "gcp_sm.rejects.dot",
			block:     "gcp_sm",
			body:      `project_id = "a-project"`,
			prefix:    "acme.prod",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		{
			name:   "k8s_secrets.dot.allowed",
			block:  "k8s_secrets",
			body:   `namespace = "hush-am"`,
			prefix: "acme.prod",
		},
		{
			name:      "k8s_secrets.rejects.underscore",
			block:     "k8s_secrets",
			body:      `namespace = "hush-am"`,
			prefix:    "acme_prod",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		{
			name:   "hc_vault.slash.delimits",
			block:  "hc_vault",
			body:   vaultBody,
			prefix: "acme/prod/secrets",
		},
		{
			name:   "hc_vault.underscore.and.dot.allowed",
			block:  "hc_vault",
			body:   vaultBody,
			prefix: "acme_prod.v1",
		},
		{
			name:      "hc_vault.rejects.plus",
			block:     "hc_vault",
			body:      vaultBody,
			prefix:    "acme+corp",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		// nothing is reserved: the silo writes under <mount>/data/<prefix>/...
		{
			name:   "hc_vault.reserves.nothing",
			block:  "hc_vault",
			body:   vaultBody,
			prefix: "metadata",
		},
		// the address rule the API applies, applied here so a customer hears
		// it at plan time rather than on apply
		{
			name:      "hc_vault.rejects.plaintext.address",
			block:     "hc_vault",
			body:      "address = \"http://vault.acme.internal:8200\"\n    auth { role = \"hush-am\" }",
			prefix:    "acme",
			expectErr: regexp.MustCompile(`address must be an https URL`),
		},
		{
			name:      "hc_vault.rejects.address.without.a.scheme",
			block:     "hc_vault",
			body:      "address = \"vault.acme.internal:8200\"\n    auth { role = \"hush-am\" }",
			prefix:    "acme",
			expectErr: regexp.MustCompile(`address must be an https URL`),
		},
		{
			name:      "hc_vault.rejects.unknown.auth.method",
			block:     "hc_vault",
			body:      vaultAuth(`method = "approle"` + "\n      " + `role = "hush-am"`),
			prefix:    "acme",
			expectErr: regexp.MustCompile(`expected hc_vault\.0\.auth\.0\.method to be one of`),
		},
		// every branch of the diff that requires a field, or refuses one
		// belonging to another method
		{
			name:      "hc_vault.kubernetes.requires.a.role",
			block:     "hc_vault",
			body:      vaultAuth(""),
			prefix:    "acme",
			expectErr: regexp.MustCompile(`auth method "kubernetes" needs a role`),
		},
		{
			name:   "hc_vault.jwt.accepts.a.role",
			block:  "hc_vault",
			body:   vaultAuth(`method = "jwt"` + "\n      " + `role = "hush-am"`),
			prefix: "acme",
		},
		{
			name:      "hc_vault.jwt.requires.a.role",
			block:     "hc_vault",
			body:      vaultAuth(`method = "jwt"`),
			prefix:    "acme",
			expectErr: regexp.MustCompile(`auth method "jwt" needs a role`),
		},
		{
			name:   "hc_vault.token.names.nothing",
			block:  "hc_vault",
			body:   vaultAuth(`method = "token"`),
			prefix: "acme",
		},
		{
			name:      "hc_vault.token.refuses.a.role",
			block:     "hc_vault",
			body:      vaultAuth(`method = "token"` + "\n      " + `role = "hush-am"`),
			prefix:    "acme",
			expectErr: regexp.MustCompile(`auth method "token" does not use role`),
		},
		{
			name:      "hc_vault.token.refuses.a.mount",
			block:     "hc_vault",
			body:      vaultAuth(`method = "token"` + "\n      " + `mount = "kubernetes"`),
			prefix:    "acme",
			expectErr: regexp.MustCompile(`auth method "token" does not use mount`),
		},
		// Both mounts reach a request path the access manager builds, so a
		// mount that walks out of it is refused here as midgard refuses it.
		{
			name:  "hc_vault.kv.mount.stays.in.its.path",
			block: "hc_vault",
			body: `address = "https://vault.acme.internal:8200"
    mount = "../sys/mounts"
    auth {
      role = "hush-am"
    }`,
			prefix:    "acme",
			expectErr: regexp.MustCompile(`mount must be letters, digits`),
		},
		{
			name:      "hc_vault.auth.mount.stays.in.its.path",
			block:     "hc_vault",
			body:      vaultAuth(`role = "hush-am"` + "\n      " + `mount = "../sys/mounts"`),
			prefix:    "acme",
			expectErr: regexp.MustCompile(`mount must be letters, digits`),
		},
		{
			name:  "hc_vault.nested.mounts.are.allowed",
			block: "hc_vault",
			body: `address = "https://vault.acme.internal:8200"
    mount = "team/hush-kv"
    auth {
      role = "hush-am"
      mount = "k8s/east"
    }`,
			prefix: "acme",
		},
		{
			name:      "hc_vault.requires.the.auth.block",
			block:     "hc_vault",
			body:      `address = "https://vault.acme.internal:8200"`,
			prefix:    "acme",
			expectErr: regexp.MustCompile(`Insufficient auth blocks|"hc_vault\.0\.auth" is required`),
		},
		{
			name:      "uppercase.rejected",
			block:     "aws_sm",
			body:      `region = "eu-west-1"`,
			prefix:    "AcmeProd",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		{
			name:      "leading.punctuation.rejected",
			block:     "aws_sm",
			body:      `region = "eu-west-1"`,
			prefix:    "-acme",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		{
			name:      "repeated.separator.rejected",
			block:     "aws_sm",
			body:      `region = "eu-west-1"`,
			prefix:    "acme//prod",
			expectErr: regexp.MustCompile(`prefix must be`),
		},
		{
			name:   "repeated.punctuation.allowed",
			block:  "aws_sm",
			body:   `region = "eu-west-1"`,
			prefix: "acme__prod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// plan-only: the provider owns plan-time validation
			step := resource.TestStep{
				Config:             secretStorePrefixConfig(tt.block, tt.prefix, tt.body),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}
			if tt.expectErr != nil {
				step.ExpectError = tt.expectErr
			}
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps:             []resource.TestStep{step},
			})
		})
	}
}

// Config is immutable in the API, so a prefix change must plan a replacement;
// if the two disagree it becomes an in-place edit the API rejects.
func TestAccResourceSecretStorePrefixForcesNew(t *testing.T) {
	var storeID string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("secret_store", "v1/secret_stores"),
		Steps: []resource.TestStep{
			{
				Config: secretStorePrefixConfig("aws_sm", "acme", `region = "eu-west-1"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_secret_store.prefix", "aws_sm.0.prefix", "acme",
					),
					recordID("hush_secret_store.prefix", &storeID),
				),
			},
			{
				Config: secretStorePrefixConfig("aws_sm", "acme-prod",
					`region = "eu-west-1"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_secret_store.prefix", "aws_sm.0.prefix", "acme-prod",
					),
					checkIDChanged("hush_secret_store.prefix", &storeID),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func secretStorePrefixConfig(block, prefix, body string) string {
	return fmt.Sprintf(`
resource "hush_secret_store" "prefix" {
  name = "prefix-test-store"

  %s {
    prefix = %q
    %s
  }
}
`, block, prefix, body)
}
