package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSnowflakeAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("snowflake_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: snowflakeAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_snowflake_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "name", "test-snowflake-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "description", "test snowflake privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "grants.0.resource_type", "table",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "grants.0.privileges.0", "SELECT",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "grants.0.privileges.1", "INSERT",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "grants.1.resource_type", "warehouse",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "grants.1.privileges.0", "USAGE",
					),
				),
			},
			{
				Config: snowflakeAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_snowflake_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "name", "test-snowflake-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_privilege.test", "description", "updated snowflake privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSnowflakeAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("snowflake_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: snowflakeAccessPrivilegeStep1() + snowflakeAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_snowflake_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_snowflake_access_privilege.test", "name", "test-snowflake-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_snowflake_access_privilege.test", "grants.0.resource_type", "table",
					),
				),
			},
		},
	})
}

func snowflakeAccessPrivilegeStep1() string {
	return `
resource "hush_snowflake_access_privilege" "test" {
  name        = "test-snowflake-priv"
  description = "test snowflake privilege"

  grants {
    privileges    = ["SELECT", "INSERT"]
    resource_type = "table"
  }

  grants {
    privileges    = ["USAGE"]
    resource_type = "warehouse"
  }
}
`
}

func snowflakeAccessPrivilegeStep2() string {
	return `
resource "hush_snowflake_access_privilege" "test" {
  name        = "test-snowflake-priv-updated"
  description = "updated snowflake privilege"

  grants {
    privileges    = ["SELECT", "INSERT"]
    resource_type = "table"
  }

  grants {
    privileges    = ["USAGE"]
    resource_type = "warehouse"
  }
}
`
}

const snowflakeAccessPrivilegeDataSource = `
data "hush_snowflake_access_privilege" "test" {
  id = hush_snowflake_access_privilege.test.id
}
`
