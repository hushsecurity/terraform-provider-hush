package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRedisAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "name", "test-redis-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "description", "test redis privilege",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "grants.0.type", "category",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "grants.0.action", "include",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "grants.0.name", "read",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "keys.0", "*",
					),
				),
			},
			{
				Config: redisAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "name", "test-redis-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_privilege.test", "description", "updated redis privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceRedisAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessPrivilegeStep1() + redisAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_redis_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_redis_access_privilege.test", "name", "test-redis-priv",
					),
					resource.TestCheckResourceAttr(
						"data.hush_redis_access_privilege.test", "grants.0.type", "category",
					),
				),
			},
		},
	})
}

func redisAccessPrivilegeStep1() string {
	return `
resource "hush_redis_access_privilege" "test" {
  name        = "test-redis-priv"
  description = "test redis privilege"

  grants {
    type   = "category"
    action = "include"
    name   = "read"
  }

  keys = ["*"]
}
`
}

func redisAccessPrivilegeStep2() string {
	return `
resource "hush_redis_access_privilege" "test" {
  name        = "test-redis-priv-updated"
  description = "updated redis privilege"

  grants {
    type   = "category"
    action = "include"
    name   = "read"
  }

  keys = ["*"]
}
`
}

const redisAccessPrivilegeDataSource = `
data "hush_redis_access_privilege" "test" {
  id = hush_redis_access_privilege.test.id
}
`
