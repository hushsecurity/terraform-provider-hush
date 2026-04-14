package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRedisAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "name", "test-redis-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "description", "test redis credential",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "host", "test-redis.example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "port", "6379",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "tls", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "username", "testuser",
					),
				),
			},
			{
				Config: redisAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "name", "test-redis-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "description", "updated redis credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceRedisAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialStep1() + redisAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_redis_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_redis_access_credential.test", "name", "test-redis-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_redis_access_credential.test", "host", "test-redis.example.com",
					),
				),
			},
		},
	})
}

func redisAccessCredentialStep1() string {
	return `
resource "hush_redis_access_credential" "test" {
  name           = "test-redis-cred"
  description    = "test redis credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "test-redis.example.com"
  port           = 6379
  tls            = true
  username       = "testuser"
  password       = "testpassword123"
}
`
}

func redisAccessCredentialStep2() string {
	return `
resource "hush_redis_access_credential" "test" {
  name           = "test-redis-cred-updated"
  description    = "updated redis credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "test-redis.example.com"
  port           = 6379
  tls            = true
  username       = "testuser"
  password       = "testpassword123"
}
`
}

const redisAccessCredentialDataSource = `
data "hush_redis_access_credential" "test" {
  id = hush_redis_access_credential.test.id
}
`
