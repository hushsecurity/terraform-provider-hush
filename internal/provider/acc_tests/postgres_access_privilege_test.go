package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourcePostgresAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_postgres_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "name", "test-pg-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "description", "test postgres privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.object_type", "TABLE",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.privileges.0", "SELECT",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.privileges.1", "INSERT",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.all_in_schema", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "grants.0.object_names.0", "public",
					),
				),
			},
			{
				Config: postgresAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_postgres_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "name", "test-pg-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_privilege.test", "description", "updated postgres privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourcePostgresAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessPrivilegeStep1() + postgresAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_postgres_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_privilege.test", "name", "test-pg-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_privilege.test", "grants.0.object_type", "TABLE",
					),
				),
			},
		},
	})
}

func postgresAccessPrivilegeStep1() string {
	return `
resource "hush_postgres_access_privilege" "test" {
  name        = "test-pg-priv"
  description = "test postgres privilege"

  grants {
    privileges    = ["SELECT", "INSERT"]
    object_type   = "TABLE"
    all_in_schema = true
    object_names  = ["public"]
  }
}
`
}

func postgresAccessPrivilegeStep2() string {
	return `
resource "hush_postgres_access_privilege" "test" {
  name        = "test-pg-priv-updated"
  description = "updated postgres privilege"

  grants {
    privileges    = ["SELECT", "INSERT"]
    object_type   = "TABLE"
    all_in_schema = true
    object_names  = ["public"]
  }
}
`
}

const postgresAccessPrivilegeDataSource = `
data "hush_postgres_access_privilege" "test" {
  id = hush_postgres_access_privilege.test.id
}
`
