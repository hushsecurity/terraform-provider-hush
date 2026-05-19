package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGitlabIntegration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_integration", "v1/integrations"),
		Steps: []resource.TestStep{
			{
				Config: gitlabIntegrationStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gitlab_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "name", "test-gitlab",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "group_id", "12345",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "visibilities.0", "private",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "status", "ok",
					),
				),
			},
			{
				Config: gitlabIntegrationStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gitlab_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "name", "test-gitlab-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "visibilities.0", "private",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_integration.test", "visibilities.1", "internal",
					),
				),
			},
			{
				ResourceName:            "hush_gitlab_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "enable_pr_scans"},
			},
		},
	})
}

func TestAccDataSourceGitlabIntegration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: gitlabIntegrationDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gitlab_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gitlab_integration.test", "name", "test-gitlab-ds",
					),
				),
			},
		},
	})
}

const gitlabIntegrationStep1 = `
resource "hush_gitlab_integration" "test" {
  name         = "test-gitlab"
  group_id     = 12345
  token        = "glpat-test-token-12345"
  visibilities = ["private"]
}
`

const gitlabIntegrationStep2 = `
resource "hush_gitlab_integration" "test" {
  name         = "test-gitlab-updated"
  group_id     = 12345
  token        = "glpat-test-token-12345"
  visibilities = ["private", "internal"]
}
`

const gitlabIntegrationDataSource = `
resource "hush_gitlab_integration" "ds_source" {
  name     = "test-gitlab-ds"
  group_id = 99999
  token    = "glpat-test-token-ds"
}

data "hush_gitlab_integration" "test" {
  id = hush_gitlab_integration.ds_source.id
}
`
