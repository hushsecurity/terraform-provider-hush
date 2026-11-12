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
			{
				// built-in roles may replace the grants; read back through
				// the data source, since resource state keeps the planned value
				Config: mongodbAtlasAccessPrivilegeBuiltinRoles() + mongodbAtlasAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "builtin_roles.#", "1",
					),
					resource.TestCheckTypeSetElemAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "builtin_roles.*", "clusterMonitor",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "grants.#", "0",
					),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasAccessPrivilege_BuiltinRoles(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				// created with built-in roles only; read back through the data
				// source, since resource state keeps the planned value
				Config: mongodbAtlasAccessPrivilegeBuiltinRoles() + mongodbAtlasAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "builtin_roles.#", "1",
					),
					resource.TestCheckTypeSetElemAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "builtin_roles.*", "clusterMonitor",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "grants.#", "0",
					),
				),
			},
			{
				ResourceName:      "hush_mongodb_atlas_access_privilege.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// back to grants only, which clears the built-in roles
				Config: mongodbAtlasAccessPrivilegeStep2() + mongodbAtlasAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "builtin_roles.#", "0",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_privilege.test", "grants.#", "1",
					),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasAccessPrivilegeValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_mongodb_atlas_access_privilege" "test" {
  name = "test-atlas-priv-empty"
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`name at least one grant or built-in role`),
			},
			{
				Config: `
resource "hush_mongodb_atlas_access_privilege" "test" {
  name          = "test-atlas-priv-bad-role"
  builtin_roles = ["atlasAdmin"]
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected builtin_roles\.\d+ to be one of \["clusterMonitor"\]`),
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

func mongodbAtlasAccessPrivilegeBuiltinRoles() string {
	return `
resource "hush_mongodb_atlas_access_privilege" "test" {
  name          = "test-atlas-priv-updated"
  description   = "updated atlas privilege"
  builtin_roles = ["clusterMonitor"]
}
`
}

const mongodbAtlasAccessPrivilegeDataSource = `
data "hush_mongodb_atlas_access_privilege" "test" {
  id = hush_mongodb_atlas_access_privilege.test.id
}
`
