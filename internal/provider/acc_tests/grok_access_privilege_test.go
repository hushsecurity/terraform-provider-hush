package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGrokAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("grok_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: grokAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_grok_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "name", "test-grok-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "description", "test grok privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "endpoints.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "endpoints.0", "Chat",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "models.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "models.0", "grok-2",
					),
				),
			},
			{
				Config: grokAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_grok_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "name", "test-grok-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "description", "updated grok privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "endpoints.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "endpoints.0", "Chat",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "endpoints.1", "Embed",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "models.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "models.0", "grok-2",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_privilege.test", "models.1", "grok-3",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGrokAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("grok_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: grokAccessPrivilegeStep1() + grokAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_grok_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_grok_access_privilege.test", "name", "test-grok-priv",
					),
				),
			},
		},
	})
}

func grokAccessPrivilegeStep1() string {
	return `
resource "hush_grok_access_privilege" "test" {
  name        = "test-grok-priv"
  description = "test grok privilege"
  endpoints   = ["Chat"]
  models      = ["grok-2"]
}
`
}

func grokAccessPrivilegeStep2() string {
	return `
resource "hush_grok_access_privilege" "test" {
  name        = "test-grok-priv-updated"
  description = "updated grok privilege"
  endpoints   = ["Chat", "Embed"]
  models      = ["grok-2", "grok-3"]
}
`
}

const grokAccessPrivilegeDataSource = `
data "hush_grok_access_privilege" "test" {
  id = hush_grok_access_privilege.test.id
}
`
