package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGCPSAAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gcp_sa_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: gcpSAAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gcp_sa_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_privilege.test", "name", "test-gcp-sa-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_privilege.test", "description", "test gcp service account privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_privilege.test", "roles.0", "roles/storage.objectViewer",
					),
				),
			},
			{
				Config: gcpSAAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gcp_sa_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_privilege.test", "name", "test-gcp-sa-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_privilege.test", "description", "updated gcp service account privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGCPSAAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gcp_sa_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: gcpSAAccessPrivilegeStep1() + gcpSAAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gcp_sa_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_sa_access_privilege.test", "name", "test-gcp-sa-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_sa_access_privilege.test", "roles.0", "roles/storage.objectViewer",
					),
				),
			},
		},
	})
}

func gcpSAAccessPrivilegeStep1() string {
	return `
resource "hush_gcp_sa_access_privilege" "test" {
  name        = "test-gcp-sa-priv"
  description = "test gcp service account privilege"
  roles       = ["roles/storage.objectViewer"]
}
`
}

func gcpSAAccessPrivilegeStep2() string {
	return `
resource "hush_gcp_sa_access_privilege" "test" {
  name        = "test-gcp-sa-priv-updated"
  description = "updated gcp service account privilege"
  roles       = ["roles/storage.objectViewer"]
}
`
}

const gcpSAAccessPrivilegeDataSource = `
data "hush_gcp_sa_access_privilege" "test" {
  id = hush_gcp_sa_access_privilege.test.id
}
`
