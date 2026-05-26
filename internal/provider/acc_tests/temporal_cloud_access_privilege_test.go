package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTemporalCloudAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("temporal_cloud_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: temporalCloudAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "name", "test-temporal-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "description", "test temporal cloud privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "grants.0.namespace", "test-namespace.acct",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "grants.0.permission", "read",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "grants.1.namespace", "prod-namespace.acct",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "grants.1.permission", "admin",
					),
				),
			},
			{
				Config: temporalCloudAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "name", "test-temporal-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_privilege.test", "description", "updated temporal cloud privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceTemporalCloudAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("temporal_cloud_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: temporalCloudAccessPrivilegeStep1() + temporalCloudAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_temporal_cloud_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_temporal_cloud_access_privilege.test", "name", "test-temporal-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_temporal_cloud_access_privilege.test", "grants.0.namespace", "test-namespace.acct",
					),
				),
			},
		},
	})
}

func temporalCloudAccessPrivilegeStep1() string {
	return `
resource "hush_temporal_cloud_access_privilege" "test" {
  name        = "test-temporal-priv"
  description = "test temporal cloud privilege"

  grants {
    namespace  = "test-namespace.acct"
    permission = "read"
  }

  grants {
    namespace  = "prod-namespace.acct"
    permission = "admin"
  }
}
`
}

func temporalCloudAccessPrivilegeStep2() string {
	return `
resource "hush_temporal_cloud_access_privilege" "test" {
  name        = "test-temporal-priv-updated"
  description = "updated temporal cloud privilege"

  grants {
    namespace  = "test-namespace.acct"
    permission = "read"
  }

  grants {
    namespace  = "prod-namespace.acct"
    permission = "admin"
  }
}
`
}

const temporalCloudAccessPrivilegeDataSource = `
data "hush_temporal_cloud_access_privilege" "test" {
  id = hush_temporal_cloud_access_privilege.test.id
}
`
