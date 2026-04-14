package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDatadogAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("datadog_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: datadogAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_datadog_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "name", "test-datadog-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "description", "test datadog privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "key_type", "application_key",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "scopes.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "scopes.0", "dashboards_read",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "scopes.1", "monitors_read",
					),
				),
			},
			{
				Config: datadogAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_datadog_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "name", "test-datadog-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "description", "updated datadog privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.test", "scopes.#", "4",
					),
				),
			},
		},
	})
}

func TestAccResourceDatadogAccessPrivilege_noScopes(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("datadog_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: datadogAccessPrivilegeNoScopesStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_datadog_access_privilege.unrestricted", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.unrestricted", "name", "test-datadog-unrestricted",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.unrestricted", "key_type", "api_key",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_privilege.unrestricted", "scopes.#", "0",
					),
				),
			},
		},
	})
}

func TestAccDataSourceDatadogAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("datadog_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: datadogAccessPrivilegeStep1() + datadogAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_datadog_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_datadog_access_privilege.test", "name", "test-datadog-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_datadog_access_privilege.test", "scopes.#", "2",
					),
				),
			},
		},
	})
}

func datadogAccessPrivilegeStep1() string {
	return `
resource "hush_datadog_access_privilege" "test" {
  name        = "test-datadog-priv"
  description = "test datadog privilege"
  key_type    = "application_key"
  scopes      = ["dashboards_read", "monitors_read"]
}
`
}

func datadogAccessPrivilegeStep2() string {
	return `
resource "hush_datadog_access_privilege" "test" {
  name        = "test-datadog-priv-updated"
  description = "updated datadog privilege"
  key_type    = "application_key"
  scopes      = ["dashboards_read", "dashboards_write", "monitors_read", "monitors_write"]
}
`
}

func datadogAccessPrivilegeNoScopesStep() string {
	return `
resource "hush_datadog_access_privilege" "unrestricted" {
  name        = "test-datadog-unrestricted"
  description = "unrestricted datadog access"
  key_type    = "api_key"
}
`
}

const datadogAccessPrivilegeDataSource = `
data "hush_datadog_access_privilege" "test" {
  id = hush_datadog_access_privilege.test.id
}
`
