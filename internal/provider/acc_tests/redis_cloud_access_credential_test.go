package acc_tests

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// midgard's _NON_REDIS_CLOUD_FIELDS: a create or PATCH naming any of these for
// a redis_cloud credential is rejected. The mock does not model engine rules,
// and stores only what the provider sent, so one appearing on the object means
// the provider sent it.
var redisNonRedisCloudFields = []string{
	"password", "username", "cache_engine",
	"region", "user_group_id", "access_key_id", "secret_access_key",
	"project", "service_name", "token",
	"tenant_id", "client_id", "client_secret", "subscription_id", "resource_group", "cluster_name",
}

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		reject := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			if obj["engine"] != "redis_cloud" {
				return nil
			}
			for field := range obj {
				if slices.Contains(redisNonRedisCloudFields, field) {
					return &testutil.HookError{
						Status: 422,
						Detail: "field(s) not valid for engine 'redis_cloud': " + field,
					}
				}
			}
			return nil
		}
		ms.OnOperation("access_credentials/redis", testutil.OpCreate, reject)
		ms.OnOperation("access_credentials/redis", testutil.OpUpdate, reject)
	})
}

// Exercises the redis_cloud engine branch: the connection fields plus the
// Redis Cloud API keys and ids, through create, an update that leaves the keys
// alone, an account change with its new secret key, and the data source.
func TestAccResourceRedisAccessCredentialRedisCloud(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisCloudConfig(redisCloudBase, `
  api_key        = "account-key-1"
  api_secret_key = "user-key-1"
`) + redisCloudDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_credential.rc", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "engine", "redis_cloud"),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.rc", "host", "redis-12345.c1.us-east-1-2.ec2.redns.redis-cloud.com",
					),
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "port", "12345"),
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "tls", "true"),
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "api_key", "account-key-1"),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.rc", "redis_cloud_subscription_id", "123",
					),
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "database_id", "456"),
					resource.TestCheckResourceAttr("data.hush_redis_access_credential.rc", "api_key", "account-key-1"),
					resource.TestCheckResourceAttr(
						"data.hush_redis_access_credential.rc", "redis_cloud_subscription_id", "123",
					),
					resource.TestCheckResourceAttr("data.hush_redis_access_credential.rc", "database_id", "456"),
				),
			},
			{
				// Moving the database needs no new secret key.
				Config: redisCloudConfig(redisCloudMoved, `
  api_key        = "account-key-1"
  api_secret_key = "user-key-1"
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "port", "23456"),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.rc", "redis_cloud_subscription_id", "789",
					),
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "database_id", "1011"),
				),
			},
			{
				// Another account, with the new secret key the change requires.
				Config: redisCloudConfig(redisCloudMoved, `
  api_key        = "account-key-2"
  api_secret_key = "user-key-2"
`),
				Check: resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "api_key", "account-key-2"),
			},
		},
	})
}

// Write-only rotation of api_secret_key. Bumping api_secret_key_wo_version
// must trigger Update and converge with no perpetual diff, and is the
// write-only way to satisfy the api_key rule.
func TestAccResourceRedisAccessCredentialWOAPISecretKeyRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisCloudConfig(redisCloudBase, `
  api_key                   = "account-key-1"
  api_secret_key_wo         = "user-key-1"
  api_secret_key_wo_version = "1"
`),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.rc", "api_secret_key_wo_version", "1",
				),
			},
			{
				Config: redisCloudConfig(redisCloudBase, `
  api_key                   = "account-key-1"
  api_secret_key_wo         = "user-key-2"
  api_secret_key_wo_version = "2"
`),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.rc", "api_secret_key_wo_version", "2",
				),
			},
			{
				Config: redisCloudConfig(redisCloudBase, `
  api_key                   = "account-key-2"
  api_secret_key_wo         = "user-key-3"
  api_secret_key_wo_version = "3"
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_redis_access_credential.rc", "api_key", "account-key-2"),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.rc", "api_secret_key_wo_version", "3",
					),
				),
			},
		},
	})
}

// The API secret key is issued under one account, so changing api_key without
// a new secret key must fail at plan time, plain or write-only.
func TestAccResourceRedisAccessCredentialRedisCloudAPIKeyNeedsSecret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisCloudConfig(redisCloudBase, `
  api_key                   = "account-key-1"
  api_secret_key_wo         = "user-key-1"
  api_secret_key_wo_version = "1"
`),
			},
			{
				Config: redisCloudConfig(redisCloudBase, `
  api_key                   = "account-key-2"
  api_secret_key_wo         = "user-key-1"
  api_secret_key_wo_version = "1"
`),
				ExpectError: regexp.MustCompile(`changing api_key requires a new api_secret_key`),
			},
		},
	})
}

// engine is ForceNew, so changing it plans a create, which the api_key rule
// must not fire on although api_key goes from unset to set.
func TestAccResourceRedisAccessCredentialEngineChangeToRedisCloud(t *testing.T) {
	var redisID string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialMigrateRedis(),
				Check:  recordID("hush_redis_access_credential.migrate", &redisID),
			},
			{
				Config: `
resource "hush_redis_access_credential" "migrate" {
  name                        = "test-redis-migrate"
  deployment_ids              = ["` + mockDeploymentID + `"]
  engine                      = "redis_cloud"
  host                        = "test-redis.example.com"
  port                        = 12345
  api_key                     = "account-key-1"
  api_secret_key              = "user-key-1"
  redis_cloud_subscription_id = 123
  database_id                 = 456
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_redis_access_credential.migrate", "engine", "redis_cloud"),
					checkIDChanged("hush_redis_access_credential.migrate", &redisID),
				),
			},
		},
	})
}

// Negative tests for every redis_cloud branch of validateEngineFields, and for
// the redis_cloud fields on each other engine. Each fails at plan time, before
// any request reaches the mock.
func TestAccResourceRedisAccessCredentialRedisCloudFieldValidation(t *testing.T) {
	keys := `
  api_key        = "account-key-1"
  api_secret_key = "user-key-1"
`
	steps := []resource.TestStep{
		{
			Config: `
resource "hush_redis_access_credential" "bad" {
  name                        = "test-redis-bad"
  deployment_ids              = ["` + mockDeploymentID + `"]
  engine                      = "redis_cloud"
  host                        = "redis.example.com"
  api_key                     = "account-key-1"
  redis_cloud_subscription_id = 123
}
`,
			ExpectError: regexp.MustCompile(`engine "redis_cloud" requires: port, api_secret_key, database_id`),
		},
		{
			Config: `
resource "hush_redis_access_credential" "bad" {
  name                        = "test-redis-bad"
  deployment_ids              = ["` + mockDeploymentID + `"]
  engine                      = "redis_cloud"
  api_key                     = "account-key-1"
  api_secret_key              = "user-key-1"
  redis_cloud_subscription_id = 123
  database_id                 = 456
}
`,
			ExpectError: regexp.MustCompile(`engine "redis_cloud" requires: host, port`),
		},
		{
			Config: redisCloudConfig(
				strings.Replace(redisCloudBase, "  port                        = 12345\n", "", 1), keys),
			ExpectError: regexp.MustCompile(`engine "redis_cloud" requires: port\s`),
		},
		{
			Config:      redisCloudConfig(strings.Replace(redisCloudBase, "= 123\n", "= 0\n", 1), keys),
			ExpectError: regexp.MustCompile(`expected redis_cloud_subscription_id to be in the range \(1 - 2147483647\)`),
		},
		{
			Config:      redisCloudConfig(strings.Replace(redisCloudBase, "= 456\n", "= 2147483648\n", 1), keys),
			ExpectError: regexp.MustCompile(`expected database_id to be in the range \(1 - 2147483647\)`),
		},
	}
	for _, tc := range []struct{ field, value string }{
		{"password", `"pw"`},
		{"username", `"u"`},
		{"cache_engine", `"valkey"`},
		{"region", `"eu-north-1"`},
		{"user_group_id", `"ug1"`},
		{"access_key_id", `"AKIAIOSFODNN7EXAMPLE"`},
		{"secret_access_key", `"secret"`},
		{"project", `"my-aiven-project"`},
		{"service_name", `"my-valkey-service"`},
		{"token", `"test-aiven-token"`},
		{"tenant_id", `"` + redisAzureTenantID + `"`},
		{"client_id", `"` + redisAzureClientID + `"`},
		{"client_secret", `"test-client-secret-v1"`},
		{"subscription_id", `"` + redisAzureSubscriptionID + `"`},
		{"resource_group", `"my-redis-rg"`},
		{"cluster_name", `"my-redis-cluster"`},
	} {
		steps = append(steps, resource.TestStep{
			Config:      redisCloudConfig(redisCloudBase, keys+"  "+tc.field+" = "+tc.value+"\n"),
			ExpectError: regexp.MustCompile(`engine "redis_cloud" does not allow: ` + tc.field),
		})
	}
	for _, tc := range []struct{ engine, base string }{
		{"redis", redisCloudOnRedis},
		{"elasticache", redisCloudOnElastiCache},
		{"aiven", redisCloudOnAiven},
		{"azure_managed_redis", redisCloudOnAzure},
	} {
		for _, f := range []struct{ field, value string }{
			{"api_key", `"account-key-1"`},
			{"api_secret_key", `"user-key-1"`},
			{"redis_cloud_subscription_id", "123"},
			{"database_id", "456"},
		} {
			steps = append(steps, resource.TestStep{
				Config: `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
` + tc.base + "  " + f.field + " = " + f.value + "\n}\n",
				ExpectError: regexp.MustCompile(`engine "` + tc.engine + `" does not allow: ` + f.field),
			})
		}
	}
	// The write-only secret key counts as the secret key on another engine too.
	steps = append(steps, resource.TestStep{
		Config: `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
` + redisCloudOnRedis + `  api_secret_key_wo         = "user-key-1"
  api_secret_key_wo_version = "1"
}
`,
		ExpectError: regexp.MustCompile(`engine "redis" does not allow: api_secret_key`),
	})
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps:             steps,
	})
}

const (
	redisCloudBase = `
  host                        = "redis-12345.c1.us-east-1-2.ec2.redns.redis-cloud.com"
  port                        = 12345
  tls                         = true
  redis_cloud_subscription_id = 123
  database_id                 = 456
`
	redisCloudMoved = `
  host                        = "redis-23456.c2.us-east-1-2.ec2.redns.redis-cloud.com"
  port                        = 23456
  tls                         = true
  redis_cloud_subscription_id = 789
  database_id                 = 1011
`
	redisCloudOnRedis = `  engine   = "redis"
  host     = "redis.example.com"
  password = "testpassword123"
`
	redisCloudOnElastiCache = `  engine        = "elasticache"
  host          = "my-cluster.cache.amazonaws.com"
  cache_engine  = "valkey"
  region        = "eu-north-1"
  user_group_id = "my-user-group"
`
	redisCloudOnAiven = `  engine       = "aiven"
  project      = "my-aiven-project"
  service_name = "my-valkey-service"
  token        = "test-aiven-token"
`
	redisCloudOnAzure = `  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
`
	redisCloudDataSource = `
data "hush_redis_access_credential" "rc" {
  id = hush_redis_access_credential.rc.id
}
`
)

// redisCloudConfig is a redis_cloud credential with the given locators and
// keys.
func redisCloudConfig(locators, keys string) string {
	return `
resource "hush_redis_access_credential" "rc" {
  name           = "test-redis-cloud"
  description    = "test redis cloud credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis_cloud"
` + locators + keys + `}
`
}
