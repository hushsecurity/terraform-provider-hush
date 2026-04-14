package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSalesforceAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("salesforce_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: salesforceAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_salesforce_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "name", "test-salesforce-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "description", "test salesforce privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "run_as_user", "admin@test.salesforce.com",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "scopes.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "scopes.0", "Api",
					),
				),
			},
			{
				Config: salesforceAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_salesforce_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "name", "test-salesforce-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "description", "updated salesforce privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "run_as_user", "admin@test.salesforce.com",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "scopes.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "scopes.0", "Api",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_privilege.test", "scopes.1", "RefreshToken",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSalesforceAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("salesforce_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: salesforceAccessPrivilegeStep1() + salesforceAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_salesforce_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_salesforce_access_privilege.test", "name", "test-salesforce-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_salesforce_access_privilege.test", "run_as_user", "admin@test.salesforce.com",
					),
				),
			},
		},
	})
}

func salesforceAccessPrivilegeStep1() string {
	return `
resource "hush_salesforce_access_privilege" "test" {
  name        = "test-salesforce-priv"
  description = "test salesforce privilege"
  run_as_user = "admin@test.salesforce.com"
  scopes      = ["Api"]
}
`
}

func salesforceAccessPrivilegeStep2() string {
	return `
resource "hush_salesforce_access_privilege" "test" {
  name        = "test-salesforce-priv-updated"
  description = "updated salesforce privilege"
  run_as_user = "admin@test.salesforce.com"
  scopes      = ["Api", "RefreshToken"]
}
`
}

const salesforceAccessPrivilegeDataSource = `
data "hush_salesforce_access_privilege" "test" {
  id = hush_salesforce_access_privilege.test.id
}
`
