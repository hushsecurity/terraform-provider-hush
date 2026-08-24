package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
