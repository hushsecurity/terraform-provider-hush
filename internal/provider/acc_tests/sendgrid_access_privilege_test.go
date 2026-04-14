package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSendGridAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("sendgrid_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: sendgridAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_sendgrid_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "name", "test-sendgrid-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "description", "test sendgrid privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "scopes.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "scopes.0", "mail.send",
					),
				),
			},
			{
				Config: sendgridAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_sendgrid_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "name", "test-sendgrid-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "description", "updated sendgrid privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "scopes.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "scopes.0", "mail.send",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_privilege.test", "scopes.1", "templates.read",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSendGridAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("sendgrid_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: sendgridAccessPrivilegeStep1() + sendgridAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_sendgrid_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_sendgrid_access_privilege.test", "name", "test-sendgrid-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_sendgrid_access_privilege.test", "scopes.#", "1",
					),
				),
			},
		},
	})
}

func sendgridAccessPrivilegeStep1() string {
	return `
resource "hush_sendgrid_access_privilege" "test" {
  name        = "test-sendgrid-priv"
  description = "test sendgrid privilege"
  scopes      = ["mail.send"]
}
`
}

func sendgridAccessPrivilegeStep2() string {
	return `
resource "hush_sendgrid_access_privilege" "test" {
  name        = "test-sendgrid-priv-updated"
  description = "updated sendgrid privilege"
  scopes      = ["mail.send", "templates.read"]
}
`
}

const sendgridAccessPrivilegeDataSource = `
data "hush_sendgrid_access_privilege" "test" {
  id = hush_sendgrid_access_privilege.test.id
}
`
