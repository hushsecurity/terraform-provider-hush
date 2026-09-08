package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAuth0AccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("auth0_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: auth0AccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_auth0_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_privilege.test", "name", "test-auth0-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_privilege.test", "application_id", "abc123ApplicationId",
					),
				),
			},
			{
				// Pointing the privilege at a different application is what
				// makes the next version rotate somewhere else.
				Config: auth0AccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_auth0_access_privilege.test", "name", "test-auth0-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_privilege.test", "application_id", "def456ApplicationId",
					),
				),
			},
		},
	})
}

func TestAccResourceAuth0AccessPrivilege_RequiresApplicationID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      auth0AccessPrivilegeNoApplication(),
				ExpectError: regexp.MustCompile(`application_id`),
				PlanOnly:    true,
			},
			{
				Config:      auth0AccessPrivilegeEmptyApplication(),
				ExpectError: regexp.MustCompile(`application_id`),
				PlanOnly:    true,
			},
		},
	})
}

func auth0AccessPrivilegeStep1() string {
	return `
resource "hush_auth0_access_privilege" "test" {
  name           = "test-auth0-priv"
  description    = "test auth0 privilege"
  application_id = "abc123ApplicationId"
}
`
}

func auth0AccessPrivilegeStep2() string {
	return `
resource "hush_auth0_access_privilege" "test" {
  name           = "test-auth0-priv-updated"
  description    = "updated auth0 privilege"
  application_id = "def456ApplicationId"
}
`
}

func auth0AccessPrivilegeNoApplication() string {
	return `
resource "hush_auth0_access_privilege" "test" {
  name = "test-auth0-priv-empty"
}
`
}

func auth0AccessPrivilegeEmptyApplication() string {
	return `
resource "hush_auth0_access_privilege" "test" {
  name           = "test-auth0-priv-blank"
  application_id = ""
}
`
}
