package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceApigeeAccessPrivilege_appName(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("apigee_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: apigeeAccessPrivilegeAppNameStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_apigee_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "name", "test-apigee-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "developer_email", "dev@example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "project_id", "my-gcp-project",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "api_products.0", "product-a",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "app_name", "my-app",
					),
				),
			},
			{
				Config: apigeeAccessPrivilegeAppNameStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_apigee_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "name", "test-apigee-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "api_products.0", "product-a",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.test", "api_products.1", "product-b",
					),
				),
			},
		},
	})
}

func TestAccResourceApigeeAccessPrivilege_appConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("apigee_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: apigeeAccessPrivilegeAppConfigStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_apigee_access_privilege.config_test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_privilege.config_test", "app_config.0.display_name", "My New App",
					),
				),
			},
		},
	})
}

func TestAccDataSourceApigeeAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("apigee_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: apigeeAccessPrivilegeAppNameStep1() + apigeeAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_apigee_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_apigee_access_privilege.test", "name", "test-apigee-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_apigee_access_privilege.test", "developer_email", "dev@example.com",
					),
				),
			},
		},
	})
}

func apigeeAccessPrivilegeAppNameStep1() string {
	return `
resource "hush_apigee_access_privilege" "test" {
  name            = "test-apigee-priv"
  description     = "test apigee privilege"
  developer_email = "dev@example.com"
  project_id      = "my-gcp-project"
  api_products    = ["product-a"]
  app_name        = "my-app"
}
`
}

func apigeeAccessPrivilegeAppNameStep2() string {
	return `
resource "hush_apigee_access_privilege" "test" {
  name            = "test-apigee-priv-updated"
  description     = "updated apigee privilege"
  developer_email = "dev@example.com"
  project_id      = "my-gcp-project"
  api_products    = ["product-a", "product-b"]
  app_name        = "my-app"
}
`
}

func apigeeAccessPrivilegeAppConfigStep() string {
	return `
resource "hush_apigee_access_privilege" "config_test" {
  name            = "test-apigee-config"
  description     = "apigee privilege with app config"
  developer_email = "dev@example.com"
  project_id      = "my-gcp-project"
  api_products    = ["product-a"]

  app_config {
    display_name = "My New App"
  }
}
`
}

const apigeeAccessPrivilegeDataSource = `
data "hush_apigee_access_privilege" "test" {
  id = hush_apigee_access_privilege.test.id
}
`
