package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRabbitmqAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_rabbitmq_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "name", "test-rabbitmq-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "description", "test rabbitmq privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "permissions.0.vhost", "/",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "permissions.0.configure", ".*",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "permissions.0.write", ".*",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "permissions.0.read", ".*",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "tags.0", "monitoring",
					),
				),
			},
			{
				Config: rabbitmqAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_rabbitmq_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "name", "test-rabbitmq-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "description", "updated rabbitmq privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "permissions.0.vhost", "/production",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.test", "tags.0", "administrator",
					),
				),
			},
		},
	})
}

func TestAccResourceRabbitmqAccessPrivilege_multiplePermissions(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessPrivilegeMultiplePermissions(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_rabbitmq_access_privilege.multi", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.multi", "permissions.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.multi", "permissions.0.vhost", "/",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_privilege.multi", "permissions.1.vhost", "/staging",
					),
				),
			},
		},
	})
}

func TestAccDataSourceRabbitmqAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessPrivilegeStep1() + rabbitmqAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_rabbitmq_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_rabbitmq_access_privilege.test", "name", "test-rabbitmq-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_rabbitmq_access_privilege.test", "permissions.0.vhost", "/",
					),
				),
			},
		},
	})
}

func rabbitmqAccessPrivilegeStep1() string {
	return `
resource "hush_rabbitmq_access_privilege" "test" {
  name        = "test-rabbitmq-priv"
  description = "test rabbitmq privilege"

  permissions {
    vhost     = "/"
    configure = ".*"
    write     = ".*"
    read      = ".*"
  }

  tags = ["monitoring"]
}
`
}

func rabbitmqAccessPrivilegeStep2() string {
	return `
resource "hush_rabbitmq_access_privilege" "test" {
  name        = "test-rabbitmq-priv-updated"
  description = "updated rabbitmq privilege"

  permissions {
    vhost     = "/production"
    configure = ""
    write     = ".*"
    read      = ".*"
  }

  tags = ["administrator"]
}
`
}

func rabbitmqAccessPrivilegeMultiplePermissions() string {
	return `
resource "hush_rabbitmq_access_privilege" "multi" {
  name        = "test-rabbitmq-multi-perms"
  description = "multiple permissions privilege"

  permissions {
    vhost     = "/"
    configure = ".*"
    write     = ".*"
    read      = ".*"
  }

  permissions {
    vhost     = "/staging"
    configure = ""
    write     = "app\\..*"
    read      = ".*"
  }
}
`
}

const rabbitmqAccessPrivilegeDataSource = `
data "hush_rabbitmq_access_privilege" "test" {
  id = hush_rabbitmq_access_privilege.test.id
}
`
