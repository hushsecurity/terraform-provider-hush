package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// The github_app delivery block, with git_credential_helper left to its
// default, set true, and set false, and switched in place.
func TestAccResourceAccessPolicy_withGithubAppDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyGithubAppDelivery(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "github_app_delivery_config.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "github_app_delivery_config.0.git_credential_helper", "true",
					),
				),
			},
			{
				Config: accessPolicyGithubAppDelivery("git_credential_helper = false"),
				Check: resource.TestCheckResourceAttr(
					"hush_access_policy.test", "github_app_delivery_config.0.git_credential_helper", "false",
				),
			},
			{
				Config: accessPolicyGithubAppDelivery("git_credential_helper = true"),
				Check: resource.TestCheckResourceAttr(
					"hush_access_policy.test", "github_app_delivery_config.0.git_credential_helper", "true",
				),
			},
		},
	})
}

func TestAccDataSourceAccessPolicy_withGithubAppDelivery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyGithubAppDelivery("git_credential_helper = false") + `
data "hush_access_policy" "test" {
  id = hush_access_policy.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "github_app_delivery_config.#", "1",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "github_app_delivery_config.0.git_credential_helper", "false",
					),
				),
			},
		},
	})
}

func TestAccResourceAccessPolicy_githubAppDeliveryIsExclusive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessCredentialStep1() + `
resource "hush_access_policy" "test" {
  name                 = "test-policy-github-app-two-deliveries"
  enabled              = true
  access_credential_id = hush_github_app_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  github_app_delivery_config {}

  env_delivery_config {
    name = "GITHUB_TOKEN"
    key  = "token"
  }
}
`,
				ExpectError: regexp.MustCompile(`only one of`),
			},
		},
	})
}

func accessPolicyGithubAppDelivery(block string) string {
	return githubAppAccessCredentialStep1() +
		githubAppAccessPrivilege(`["repo-b"]`, `{ contents = "read" }`) + `
resource "hush_access_policy" "test" {
  name                 = "test-policy-github-app"
  description          = "test GitHub App delivery policy"
  enabled              = true
  access_credential_id = hush_github_app_access_credential.test.id
  access_privilege_ids = [hush_github_app_access_privilege.test.id]
  deployment_ids       = ["` + mockDeploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  github_app_delivery_config {
    ` + block + `
  }
}
`
}
