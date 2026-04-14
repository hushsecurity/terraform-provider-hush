package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockGitlabToken = "glpat-mock-token-1234567890"
const mockGitlabResourceID = "12345"

func TestAccResourceGitlabAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gitlabAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gitlab_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_credential.test", "name", "test-gitlab-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_credential.test", "description", "test gitlab credential",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_credential.test", "base_url", "https://gitlab.com",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_credential.test", "resource_type", "group",
					),
				),
			},
			{
				Config: gitlabAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gitlab_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_credential.test", "name", "test-gitlab-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_credential.test", "description", "updated gitlab credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGitlabAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gitlabAccessCredentialStep1 + gitlabAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gitlab_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gitlab_access_credential.test", "name", "test-gitlab-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gitlab_access_credential.test", "resource_type", "group",
					),
				),
			},
		},
	})
}

const gitlabAccessCredentialStep1 = `
resource "hush_gitlab_access_credential" "test" {
  name           = "test-gitlab-cred"
  description    = "test gitlab credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  token          = "` + mockGitlabToken + `"
  resource_type  = "group"
  resource_id    = "` + mockGitlabResourceID + `"
}
`

const gitlabAccessCredentialStep2 = `
resource "hush_gitlab_access_credential" "test" {
  name           = "test-gitlab-cred-updated"
  description    = "updated gitlab credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  token          = "` + mockGitlabToken + `"
  resource_type  = "group"
  resource_id    = "` + mockGitlabResourceID + `"
}
`

const gitlabAccessCredentialDataSource = `
data "hush_gitlab_access_credential" "test" {
  id = hush_gitlab_access_credential.test.id
}
`
