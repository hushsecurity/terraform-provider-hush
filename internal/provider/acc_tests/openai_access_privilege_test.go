package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceOpenAIAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_openai_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.test", "name", "test-openai-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.test", "description", "test openai privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.test", "permission_type", "Member",
					),
				),
			},
			{
				Config: openaiAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_openai_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.test", "name", "test-openai-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.test", "description", "updated openai privilege",
					),
				),
			},
		},
	})
}

func TestAccResourceOpenAIAccessPrivilege_restricted(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessPrivilegeRestrictedStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_openai_access_privilege.restricted", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.restricted", "permission_type", "Restricted",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.restricted", "permissions.0.name", "Models",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_privilege.restricted", "permissions.0.level", "read",
					),
				),
			},
		},
	})
}

func TestAccDataSourceOpenAIAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessPrivilegeStep1() + openaiAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_openai_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_openai_access_privilege.test", "name", "test-openai-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_openai_access_privilege.test", "permission_type", "Member",
					),
				),
			},
		},
	})
}

func openaiAccessPrivilegeStep1() string {
	return `
resource "hush_openai_access_privilege" "test" {
  name            = "test-openai-priv"
  description     = "test openai privilege"
  permission_type = "Member"
}
`
}

func openaiAccessPrivilegeStep2() string {
	return `
resource "hush_openai_access_privilege" "test" {
  name            = "test-openai-priv-updated"
  description     = "updated openai privilege"
  permission_type = "Member"
}
`
}

func openaiAccessPrivilegeRestrictedStep() string {
	return `
resource "hush_openai_access_privilege" "restricted" {
  name            = "test-openai-restricted"
  description     = "restricted openai privilege"
  permission_type = "Restricted"

  permissions {
    name  = "Models"
    level = "read"
  }
}
`
}

const openaiAccessPrivilegeDataSource = `
data "hush_openai_access_privilege" "test" {
  id = hush_openai_access_privilege.test.id
}
`
