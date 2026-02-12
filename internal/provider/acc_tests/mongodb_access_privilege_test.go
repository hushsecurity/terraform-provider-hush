package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMongoDBAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "name", "test-mongo-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "description", "test mongodb privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "grants.0.resource_type", "database",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "grants.0.privileges.0", "changeStream",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "grants.0.privileges.1", "collStats",
					),
				),
			},
			{
				Config: mongodbAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "name", "test-mongo-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_privilege.test", "description", "updated mongodb privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMongoDBAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAccessPrivilegeStep1() + mongodbAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mongodb_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_access_privilege.test", "name", "test-mongo-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_access_privilege.test", "grants.0.resource_type", "database",
					),
				),
			},
		},
	})
}

func mongodbAccessPrivilegeStep1() string {
	return `
resource "hush_mongodb_access_privilege" "test" {
  name        = "test-mongo-priv"
  description = "test mongodb privilege"

  grants {
    privileges    = ["changeStream", "collStats"]
    resource_type = "database"
  }
}
`
}

func mongodbAccessPrivilegeStep2() string {
	return `
resource "hush_mongodb_access_privilege" "test" {
  name        = "test-mongo-priv-updated"
  description = "updated mongodb privilege"

  grants {
    privileges    = ["changeStream", "collStats"]
    resource_type = "database"
  }
}
`
}

const mongodbAccessPrivilegeDataSource = `
data "hush_mongodb_access_privilege" "test" {
  id = hush_mongodb_access_privilege.test.id
}
`
