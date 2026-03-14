package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAzureAppAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: azureAppAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_azure_app_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_privilege.test", "name", "test-azure-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_privilege.test", "description", "test azure privilege",
					),
				),
			},
			{
				Config: azureAppAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_azure_app_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_privilege.test", "name", "test-azure-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_privilege.test", "description", "updated azure privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAzureAppAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_app_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: azureAppAccessPrivilegeStep1() + azureAppAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_azure_app_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_azure_app_access_privilege.test", "name", "test-azure-priv",
					),
				),
			},
		},
	})
}

func azureAppAccessPrivilegeStep1() string {
	return `
resource "hush_azure_app_access_privilege" "test" {
  name        = "test-azure-priv"
  description = "test azure privilege"
  roles       = ["Storage Blob Data Reader"]
}
`
}

func azureAppAccessPrivilegeStep2() string {
	return `
resource "hush_azure_app_access_privilege" "test" {
  name        = "test-azure-priv-updated"
  description = "updated azure privilege"
  roles       = ["Storage Blob Data Reader"]
}
`
}

const azureAppAccessPrivilegeDataSource = `
data "hush_azure_app_access_privilege" "test" {
  id = hush_azure_app_access_privilege.test.id
}
`
