package acc_tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGithubAppAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessPrivilege(`["repo-a", "repo-b"]`, `{ contents = "read", issues = "write" }`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_github_app_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "repositories.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "repositories.1", "repo-b",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "permissions.%", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "permissions.issues", "write",
					),
				),
			},
			{
				// clearing the repositories means every repository of the
				// installation
				Config: githubAppAccessPrivilege(`[]`, `{ repository_projects = "admin" }`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "repositories.#", "0",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "permissions.%", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "permissions.repository_projects", "admin",
					),
				),
			},
		},
	})
}

func TestAccResourceGithubAppAccessPrivilege_WithoutRepositories(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_github_app_access_privilege" "test" {
  name        = "test-github-app-priv-all-repos"
  permissions = { metadata = "read" }
}
`,
				Check: resource.TestCheckResourceAttr(
					"hush_github_app_access_privilege.test", "repositories.#", "0",
				),
			},
		},
	})
}

func TestAccResourceGithubAppAccessPrivilege_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessPrivilege(`["repo-a", "repo-b"]`, `{ contents = "read", issues = "write" }`),
			},
			{
				ResourceName:      "hush_github_app_access_privilege.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Values unknown at plan time must not trip the plan-time checks: the
// duplicate check skips unknown repositories and the level check skips
// unknown levels.
func TestAccResourceGithubAppAccessPrivilege_UnknownValues(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "repo" {
  input = "repo-a"
}

resource "terraform_data" "level" {
  input = "read"
}

resource "hush_github_app_access_privilege" "test" {
  name         = "test-github-app-priv-unknown"
  repositories = [terraform_data.repo.output, "repo-b"]
  permissions  = { contents = terraform_data.level.output }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "repositories.0", "repo-a",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_privilege.test", "permissions.contents", "read",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGithubAppAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessPrivilege(`["repo-a"]`, `{ contents = "read" }`) + `
data "hush_github_app_access_privilege" "test" {
  id = hush_github_app_access_privilege.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_github_app_access_privilege.test", "repositories.0", "repo-a",
					),
					resource.TestCheckResourceAttr(
						"data.hush_github_app_access_privilege.test", "permissions.contents", "read",
					),
				),
			},
		},
	})
}

func TestAccResourceGithubAppAccessPrivilege_Validation(t *testing.T) {
	tooMany := make([]string, 501)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf(`"r%d"`, i)
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      githubAppAccessPrivilege(`["owner/repo"]`, `{ contents = "read" }`),
				ExpectError: regexp.MustCompile(`use the bare repository name, without the owner`),
			},
			{
				Config:      githubAppAccessPrivilege(`["."]`, `{ contents = "read" }`),
				ExpectError: regexp.MustCompile(`repository "\.": not a repository name`),
			},
			{
				Config:      githubAppAccessPrivilege(`[".."]`, `{ contents = "read" }`),
				ExpectError: regexp.MustCompile(`repository "\.\.": not a repository name`),
			},
			{
				Config:      githubAppAccessPrivilege(`["repo-a", "repo-b", "repo-a"]`, `{ contents = "read" }`),
				ExpectError: regexp.MustCompile(`repository "repo-a": listed more than once`),
			},
			{
				Config:      githubAppAccessPrivilege(`[]`, `{}`),
				ExpectError: regexp.MustCompile(`permissions: name at least one permission`),
			},
			{
				Config:      githubAppAccessPrivilege(`["a b"]`, `{ contents = "read" }`),
				ExpectError: regexp.MustCompile(`must be 1-100 letters, digits`),
			},
			{
				Config:      githubAppAccessPrivilege("["+strings.Join(tooMany, ", ")+"]", `{ contents = "read" }`),
				ExpectError: regexp.MustCompile(`at most 500 elements|500 item maximum`),
			},
			{
				Config:      githubAppAccessPrivilege(`[]`, `{ contents = "superuser" }`),
				ExpectError: regexp.MustCompile(`level must be one of read, write, admin`),
			},
			{
				Config:      githubAppAccessPrivilege(`[]`, `{ "Contents-Read" = "read" }`),
				ExpectError: regexp.MustCompile(`not a GitHub App permission name`),
			},
			{
				Config: `
resource "hush_github_app_access_privilege" "test" {
  name = "test-github-app-priv-no-perms"
}
`,
				ExpectError: regexp.MustCompile(`"permissions" is required`),
			},
		},
	})
}

func githubAppAccessPrivilege(repositories, permissions string) string {
	return `
resource "hush_github_app_access_privilege" "test" {
  name         = "test-github-app-priv"
  description  = "test github app privilege"
  repositories = ` + repositories + `
  permissions  = ` + permissions + `
}
`
}
