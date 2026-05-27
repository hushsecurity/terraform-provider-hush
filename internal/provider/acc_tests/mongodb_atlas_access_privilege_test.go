package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMongoDBAtlasAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "name", "test-atlas-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "description", "test atlas privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "grants.0.resource_type", "collection",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "grants.0.privileges.0", "FIND",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "grants.0.privileges.1", "INSERT",
					),
				),
			},
			{
				Config: mongodbAtlasAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "name", "test-atlas-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_privilege.test", "description", "updated atlas privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMongoDBAtlasAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessPrivilegeStep1() + mongodbAtlasAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "name", "test-atlas-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "grants.0.resource_type", "collection",
					),
				),
			},
		},
	})
}

func mongodbAtlasAccessPrivilegeStep1() string {
	return `
resource "hush_mongodb_atlas_access_privilege" "test" {
  name        = "test-atlas-priv"
  description = "test atlas privilege"

  grants {
    privileges     = ["FIND", "INSERT"]
    resource_type  = "collection"
    resource_names = ["users"]
  }
}
`
}

func mongodbAtlasAccessPrivilegeStep2() string {
	return `
resource "hush_mongodb_atlas_access_privilege" "test" {
  name        = "test-atlas-priv-updated"
  description = "updated atlas privilege"

  grants {
    privileges     = ["FIND", "INSERT"]
    resource_type  = "collection"
    resource_names = ["users"]
  }
}
`
}

const mongodbAtlasAccessPrivilegeDataSource = `
data "hush_mongodb_atlas_access_privilege" "test" {
  id = hush_mongodb_atlas_access_privilege.test.id
}
`
