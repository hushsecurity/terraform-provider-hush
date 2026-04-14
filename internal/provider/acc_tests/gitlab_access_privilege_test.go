package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGitlabAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: gitlabAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gitlab_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "name", "test-gitlab-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "description", "test gitlab privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "access_level", "Developer",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "scopes.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "scopes.0", "read_api",
					),
				),
			},
			{
				Config: gitlabAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gitlab_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "name", "test-gitlab-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "description", "updated gitlab privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "access_level", "Maintainer",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "scopes.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "scopes.0", "read_api",
					),
					resource.TestCheckResourceAttr(
						"hush_gitlab_access_privilege.test", "scopes.1", "read_repository",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGitlabAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gitlab_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: gitlabAccessPrivilegeStep1() + gitlabAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gitlab_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gitlab_access_privilege.test", "name", "test-gitlab-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gitlab_access_privilege.test", "access_level", "Developer",
					),
				),
			},
		},
	})
}

func gitlabAccessPrivilegeStep1() string {
	return `
resource "hush_gitlab_access_privilege" "test" {
  name         = "test-gitlab-priv"
  description  = "test gitlab privilege"
  scopes       = ["read_api"]
  access_level = "Developer"
}
`
}

func gitlabAccessPrivilegeStep2() string {
	return `
resource "hush_gitlab_access_privilege" "test" {
  name         = "test-gitlab-priv-updated"
  description  = "updated gitlab privilege"
  scopes       = ["read_api", "read_repository"]
  access_level = "Maintainer"
}
`
}

const gitlabAccessPrivilegeDataSource = `
data "hush_gitlab_access_privilege" "test" {
  id = hush_gitlab_access_privilege.test.id
}
`
