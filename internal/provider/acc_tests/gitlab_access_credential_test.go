package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestGitlabToken = "HUSH_TEST_GITLAB_TOKEN"
const envHushTestGitlabResourceID = "HUSH_TEST_GITLAB_RESOURCE_ID"

func testAccGitlabAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestGitlabToken) == "" {
		t.Fatalf("%s env var must be set", envHushTestGitlabToken)
	}
	if os.Getenv(envHushTestGitlabResourceID) == "" {
		t.Fatalf("%s env var must be set", envHushTestGitlabResourceID)
	}
}

func TestAccResourceGitlabAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccGitlabAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gitlabAccessCredentialStep1(),
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
				Config: gitlabAccessCredentialStep2(),
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
		PreCheck:          func() { testAccGitlabAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gitlabAccessCredentialStep1() + gitlabAccessCredentialDataSource,
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

func gitlabAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	token := os.Getenv(envHushTestGitlabToken)
	resourceID := os.Getenv(envHushTestGitlabResourceID)
	return `
resource "hush_gitlab_access_credential" "test" {
  name           = "test-gitlab-cred"
  description    = "test gitlab credential"
  deployment_ids = ["` + deploymentID + `"]
  token          = "` + token + `"
  resource_type  = "group"
  resource_id    = "` + resourceID + `"
}
`
}

func gitlabAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	token := os.Getenv(envHushTestGitlabToken)
	resourceID := os.Getenv(envHushTestGitlabResourceID)
	return `
resource "hush_gitlab_access_credential" "test" {
  name           = "test-gitlab-cred-updated"
  description    = "updated gitlab credential"
  deployment_ids = ["` + deploymentID + `"]
  token          = "` + token + `"
  resource_type  = "group"
  resource_id    = "` + resourceID + `"
}
`
}

const gitlabAccessCredentialDataSource = `
data "hush_gitlab_access_credential" "test" {
  id = hush_gitlab_access_credential.test.id
}
`
