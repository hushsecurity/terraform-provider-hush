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
					checkSecretStoreID("hush_redis_access_credential.test"),
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
  name            = "test-redis-cred"
  description     = "test redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  host            = "test-redis.example.com"
  port            = 6379
  tls             = true
  username        = "testuser"
  password        = "testpassword123"
  engine          = "redis"
}
`
}

func redisAccessCredentialStep2() string {
	return `
resource "hush_redis_access_credential" "test" {
  name            = "test-redis-cred-updated"
  description     = "updated redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  host            = "test-redis.example.com"
  port            = 6379
  tls             = true
  username        = "testuser"
  password        = "testpassword123"
  engine          = "redis"
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

// Exercises the Aiven engine branch (project/service_name/token). No host/port
// are sent; Hush resolves the endpoint from the Aiven API.
func TestAccResourceRedisAccessCredentialAiven(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialAivenStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_credential.aiven", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.aiven", "engine", "aiven",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.aiven", "project", "my-aiven-project",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.aiven", "service_name", "my-valkey-service",
					),
					// The aiven engine must not carry a host (Hush resolves it).
					resource.TestCheckNoResourceAttr(
						"hush_redis_access_credential.aiven", "host",
					),
				),
			},
			{
				Config: redisAccessCredentialAivenStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.aiven", "description", "updated aiven redis credential",
					),
				),
			},
		},
	})
}

// Write-only secret rotation for the Aiven engine's token. Bumping
// token_wo_version must trigger Update and converge with no perpetual diff.
func TestAccResourceRedisAccessCredentialWOTokenRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialWOTokenStep1(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.aiven", "token_wo_version", "1",
				),
			},
			{
				Config: redisAccessCredentialWOTokenStep2(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.aiven", "token_wo_version", "2",
				),
			},
		},
	})
}

// Negative tests: every branch of validateEngineFields (CustomizeDiff), both the
// missing-required and forbidden-field paths. Each fails at plan time, before
// any request reaches the mock.
func TestAccResourceRedisAccessCredentialEngineFieldValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// redis engine, password (a required field) omitted.
				Config:      redisAccessCredentialRedisMissingRequired(),
				ExpectError: regexp.MustCompile(`engine "redis" requires:.*password`),
			},
			{
				// redis engine with an aiven-only field set.
				Config:      redisAccessCredentialRedisWithAivenField(),
				ExpectError: regexp.MustCompile(`engine "redis" does not allow:.*project`),
			},
			{
				// elasticache engine, user_group_id (a required field) omitted.
				Config:      redisAccessCredentialElastiCacheMissingRequired(),
				ExpectError: regexp.MustCompile(`engine "elasticache" requires:.*user_group_id`),
			},
			{
				// elasticache engine with a redis-only field (password) set.
				Config:      redisAccessCredentialElastiCacheWithPassword(),
				ExpectError: regexp.MustCompile(`engine "elasticache" does not allow:.*password`),
			},
			{
				// aiven engine, token (a required field) omitted.
				Config:      redisAccessCredentialAivenMissingRequired(),
				ExpectError: regexp.MustCompile(`engine "aiven" requires:.*token`),
			},
			{
				// aiven engine with a connection field (host) set.
				Config:      redisAccessCredentialAivenWithHost(),
				ExpectError: regexp.MustCompile(`engine "aiven" does not allow:.*host`),
			},
		},
	})
}

func redisAccessCredentialAivenStep1() string {
	return `
resource "hush_redis_access_credential" "aiven" {
  name           = "test-redis-aiven"
  description    = "test aiven redis credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-valkey-service"
  token          = "test-aiven-token"
}
`
}

func redisAccessCredentialAivenStep2() string {
	return `
resource "hush_redis_access_credential" "aiven" {
  name           = "test-redis-aiven"
  description    = "updated aiven redis credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-valkey-service"
  token          = "test-aiven-token"
}
`
}

func redisAccessCredentialWOTokenStep1() string {
	return `
resource "hush_redis_access_credential" "aiven" {
  name             = "test-redis-wo-token"
  deployment_ids   = ["` + mockDeploymentID + `"]
  engine           = "aiven"
  project          = "my-aiven-project"
  service_name     = "my-valkey-service"
  token_wo         = "token-v1"
  token_wo_version = "1"
}
`
}

func redisAccessCredentialWOTokenStep2() string {
	return `
resource "hush_redis_access_credential" "aiven" {
  name             = "test-redis-wo-token"
  deployment_ids   = ["` + mockDeploymentID + `"]
  engine           = "aiven"
  project          = "my-aiven-project"
  service_name     = "my-valkey-service"
  token_wo         = "token-v2"
  token_wo_version = "2"
}
`
}

func redisAccessCredentialRedisMissingRequired() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis"
  host           = "redis.example.com"
}
`
}

func redisAccessCredentialRedisWithAivenField() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis"
  host           = "redis.example.com"
  password       = "testpassword123"
  project        = "should-not-be-here"
}
`
}

func redisAccessCredentialElastiCacheMissingRequired() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "elasticache"
  host           = "my-cluster.cache.amazonaws.com"
  cache_engine   = "valkey"
  region         = "eu-north-1"
}
`
}

func redisAccessCredentialElastiCacheWithPassword() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "elasticache"
  host           = "my-cluster.cache.amazonaws.com"
  cache_engine   = "valkey"
  region         = "eu-north-1"
  user_group_id  = "my-user-group"
  password       = "should-not-be-here"
}
`
}

func redisAccessCredentialAivenMissingRequired() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-valkey-service"
}
`
}

func redisAccessCredentialAivenWithHost() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-valkey-service"
  token          = "test-aiven-token"
  host           = "should-not-be-here.example.com"
}
`
}

// A required field sourced from another resource's computed attribute is unknown
// at plan time. validateEngineFields must not reject it as missing.
func TestAccResourceRedisAccessCredentialComputedRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialComputedRequired(),
				Check: resource.TestMatchResourceAttr(
					"hush_redis_access_credential.consumer", "id", regexp.MustCompile(`^acr-.+$`),
				),
			},
		},
	})
}

func redisAccessCredentialComputedRequired() string {
	return `
resource "hush_redis_access_credential" "src" {
  name           = "test-redis-src"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis"
  host           = "redis.example.com"
  password       = "testpassword123"
}

resource "hush_redis_access_credential" "consumer" {
  name           = "test-redis-consumer"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis"
  host           = "redis.example.com"
  # unknown at plan time (stand-in for random_password.x.result)
  password       = hush_redis_access_credential.src.id
}
`
}
