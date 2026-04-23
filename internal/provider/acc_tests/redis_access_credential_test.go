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
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.test", "engine", "redis",
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
  engine         = "redis"
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
  engine         = "redis"
}
`
}

const redisAccessCredentialDataSource = `
data "hush_redis_access_credential" "test" {
  id = hush_redis_access_credential.test.id
}
`

func TestAccResourceRedisAccessCredentialElastiCache(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialElastiCacheStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_credential.ec", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "engine", "elasticache",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "cache_engine", "valkey",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "region", "eu-north-1",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "user_group_id", "my-user-group",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "access_key_id", "AKIAIOSFODNN7EXAMPLE",
					),
					resource.TestCheckNoResourceAttr(
						"hush_redis_access_credential.ec", "password",
					),
				),
			},
			{
				Config: redisAccessCredentialElastiCacheStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "name", "test-elasticache-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec", "cache_engine", "redis",
					),
				),
			},
		},
	})
}

// Federation: omitting both AWS keys must be accepted (workload identity
// fallback in mufasa).
func TestAccResourceRedisAccessCredentialElastiCacheFederation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialElastiCacheFederation(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec_fed", "engine", "elasticache",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec_fed", "access_key_id", "",
					),
				),
			},
		},
	})
}

func redisAccessCredentialElastiCacheStep1() string {
	return `
resource "hush_redis_access_credential" "ec" {
  name              = "test-elasticache-cred"
  description       = "test elasticache credential"
  deployment_ids    = ["` + mockDeploymentID + `"]
  host              = "my-cluster.cache.amazonaws.com"
  port              = 6379
  tls               = true
  engine            = "elasticache"
  cache_engine      = "valkey"
  region            = "eu-north-1"
  user_group_id     = "my-user-group"
  access_key_id     = "AKIAIOSFODNN7EXAMPLE"
  secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
`
}

func redisAccessCredentialElastiCacheStep2() string {
	return `
resource "hush_redis_access_credential" "ec" {
  name              = "test-elasticache-cred-updated"
  description       = "updated elasticache credential"
  deployment_ids    = ["` + mockDeploymentID + `"]
  host              = "my-cluster.cache.amazonaws.com"
  port              = 6379
  tls               = true
  engine            = "elasticache"
  cache_engine      = "redis"
  region            = "eu-north-1"
  user_group_id     = "my-user-group"
  access_key_id     = "AKIAIOSFODNN7EXAMPLE"
  secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
`
}

func redisAccessCredentialElastiCacheFederation() string {
	return `
resource "hush_redis_access_credential" "ec_fed" {
  name           = "test-elasticache-fed"
  description    = "test elasticache credential with WIF"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "my-cluster.cache.amazonaws.com"
  port           = 6379
  tls            = true
  engine         = "elasticache"
  cache_engine   = "valkey"
  region         = "eu-north-1"
  user_group_id  = "my-user-group"
}
`
}
