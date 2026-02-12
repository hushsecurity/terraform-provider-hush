package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMySQLAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mysql_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: mysqlAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mysql_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "name", "test-mysql-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "description", "test mysql privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "grants.0.resource_type", "database",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "grants.0.privileges.0", "CREATE",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "grants.0.privileges.1", "ALTER",
					),
				),
			},
			{
				Config: mysqlAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mysql_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "name", "test-mysql-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_privilege.test", "description", "updated mysql privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMySQLAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mysql_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: mysqlAccessPrivilegeStep1() + mysqlAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mysql_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mysql_access_privilege.test", "name", "test-mysql-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mysql_access_privilege.test", "grants.0.resource_type", "database",
					),
				),
			},
		},
	})
}

func mysqlAccessPrivilegeStep1() string {
	return `
resource "hush_mysql_access_privilege" "test" {
  name        = "test-mysql-priv"
  description = "test mysql privilege"

  grants {
    privileges    = ["CREATE", "ALTER"]
    resource_type = "database"
  }
}
`
}

func mysqlAccessPrivilegeStep2() string {
	return `
resource "hush_mysql_access_privilege" "test" {
  name        = "test-mysql-priv-updated"
  description = "updated mysql privilege"

  grants {
    privileges    = ["CREATE", "ALTER"]
    resource_type = "database"
  }
}
`
}

const mysqlAccessPrivilegeDataSource = `
data "hush_mysql_access_privilege" "test" {
  id = hush_mysql_access_privilege.test.id
}
`
