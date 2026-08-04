package acc_tests

import (
	"regexp"
	"strings"
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

func redisAccessCredentialWOSecretAccessKeyStep1() string {
	return `
resource "hush_redis_access_credential" "ec_wo" {
  name                         = "test-elasticache-wo-secret"
  deployment_ids               = ["` + mockDeploymentID + `"]
  host                         = "my-cluster.cache.amazonaws.com"
  engine                       = "elasticache"
  cache_engine                 = "valkey"
  region                       = "eu-north-1"
  user_group_id                = "my-user-group"
  access_key_id                = "AKIAIOSFODNN7EXAMPLE"
  secret_access_key_wo         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  secret_access_key_wo_version = "1"
}
`
}

func redisAccessCredentialWOSecretAccessKeyStep2() string {
	return `
resource "hush_redis_access_credential" "ec_wo" {
  name                         = "test-elasticache-wo-secret"
  deployment_ids               = ["` + mockDeploymentID + `"]
  host                         = "my-cluster.cache.amazonaws.com"
  engine                       = "elasticache"
  cache_engine                 = "valkey"
  region                       = "eu-north-1"
  user_group_id                = "my-user-group"
  access_key_id                = "AKIAIOSFODNN7EXAMPLE"
  secret_access_key_wo         = "rotated/K7MDENG/bPxRfiCYEXAMPLEKEY"
  secret_access_key_wo_version = "2"
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

// Write-only secret_access_key paired with a plain access_key_id; bumping the
// version must trigger Update and converge with no perpetual diff.
func TestAccResourceRedisAccessCredentialWOSecretAccessKeyRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialWOSecretAccessKeyStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec_wo", "access_key_id", "AKIAIOSFODNN7EXAMPLE",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.ec_wo", "secret_access_key_wo_version", "1",
					),
				),
			},
			{
				Config: redisAccessCredentialWOSecretAccessKeyStep2(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.ec_wo", "secret_access_key_wo_version", "2",
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

// Exercises the Azure Managed Redis engine branch: the ARM locators plus the
// optional app credentials. No connection fields are sent; Hush resolves the
// endpoint from ARM and mints Entra ID service principals.
func TestAccResourceRedisAccessCredentialAzureManagedRedis(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialAzureStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_redis_access_credential.azure", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "engine", "azure_managed_redis",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "tenant_id", redisAzureTenantID,
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "subscription_id", redisAzureSubscriptionID,
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "resource_group", "my-redis-rg",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "cluster_name", "my-redis-cluster",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "client_id", redisAzureClientID,
					),
					// The azure engine must not carry a host (Hush resolves it).
					resource.TestCheckNoResourceAttr(
						"hush_redis_access_credential.azure", "host",
					),
				),
			},
			{
				// Re-pointing the credential at another tenant, with the fresh
				// client_secret that the rebind requires.
				Config: redisAccessCredentialAzureStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "description", "updated azure managed redis credential",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "tenant_id", redisAzureOtherTenantID,
					),
				),
			},
			{
				// The remaining locators are exempt from the rebind rule: one app
				// can drive several clusters, so moving them needs no new secret.
				Config: redisAccessCredentialAzureStep3(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "subscription_id", redisAzureOtherSubscriptionID,
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "resource_group", "other-redis-rg",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "cluster_name", "other-redis-cluster",
					),
				),
			},
		},
	})
}

// Federation: omitting both client_id and client_secret must be accepted (the
// access-manager falls back to its default Azure credentials).
func TestAccResourceRedisAccessCredentialAzureDefaultCredentials(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialAzureDefaultCredentials(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure_fed", "engine", "azure_managed_redis",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure_fed", "client_id", "",
					),
				),
			},
		},
	})
}

// Dropping client_id and client_secret from a credential that had them must
// unset both (sent as explicit nulls, which the API requires) and leave the
// credential on the access-manager's default Azure credentials.
func TestAccResourceRedisAccessCredentialAzureDropAppCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialAzureStep1(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.azure", "client_id", redisAzureClientID,
				),
			},
			{
				Config: redisAccessCredentialAzureNoAppCredentials(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.azure", "client_id", "",
				),
			},
		},
	})
}

// Write-only secret rotation for the Azure engine's client_secret. Bumping
// client_secret_wo_version must trigger Update and converge with no perpetual
// diff.
func TestAccResourceRedisAccessCredentialWOClientSecretRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialWOClientSecretStep1(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.azure", "client_secret_wo_version", "1",
				),
			},
			{
				Config: redisAccessCredentialWOClientSecretStep2(),
				Check: resource.TestCheckResourceAttr(
					"hush_redis_access_credential.azure", "client_secret_wo_version", "2",
				),
			},
			{
				// A bumped client_secret_wo_version is the write-only way to
				// satisfy the rebind rule, since the secret itself is not in state.
				Config: redisAccessCredentialWOClientSecretRebind(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "client_id", redisAzureOtherClientID,
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.azure", "client_secret_wo_version", "3",
					),
				),
			},
		},
	})
}

// Negative test: the stored client_secret is issued for one app in one tenant,
// so re-pointing client_id (or tenant_id) without a fresh secret must fail at
// plan time rather than 400 mid-apply.
func TestAccResourceRedisAccessCredentialAzureRebindNeedsSecret(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialAzureStep1(),
			},
			{
				Config:      redisAccessCredentialAzureClientIDChanged(),
				ExpectError: regexp.MustCompile(`changing client_id requires a new client_secret`),
			},
			{
				Config:      redisAccessCredentialAzureTenantIDChanged(),
				ExpectError: regexp.MustCompile(`changing tenant_id requires a new client_secret`),
			},
		},
	})
}

// engine is ForceNew, so changing it plans a replacement -- a create, which the
// rebind rule must not fire on. The default-credential destination has no
// secret to offer it.
func TestAccResourceRedisAccessCredentialEngineChangeToAzure(t *testing.T) {
	var redisID string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialMigrateRedis(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.migrate", "engine", "redis",
					),
					recordID("hush_redis_access_credential.migrate", &redisID),
				),
			},
			{
				Config: redisAccessCredentialMigrateAzureDefaultCredentials(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.migrate", "engine", "azure_managed_redis",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.migrate", "tenant_id", redisAzureTenantID,
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.migrate", "client_id", "",
					),
					checkIDChanged("hush_redis_access_credential.migrate", &redisID),
				),
			},
		},
	})
}

// The same migration carrying app credentials, which satisfies the rebind rule
// only because a first-time client_secret reads as a change.
func TestAccResourceRedisAccessCredentialEngineChangeToAzureWithAppCredentials(t *testing.T) {
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
				Config: redisAccessCredentialMigrateAzureAppCredentials(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.migrate", "engine", "azure_managed_redis",
					),
					resource.TestCheckResourceAttr(
						"hush_redis_access_credential.migrate", "client_id", redisAzureClientID,
					),
					checkIDChanged("hush_redis_access_credential.migrate", &redisID),
				),
			},
		},
	})
}

// A credential on the default Azure credentials has no secret to rotate, so the
// rebind rule must point at adopting the pair rather than at a fresh secret.
func TestAccResourceRedisAccessCredentialAzureDefaultCredsRebind(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("redis_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: redisAccessCredentialAzureDefaultCredentials(),
			},
			{
				Config:      redisAccessCredentialAzureDefaultCredentialsOtherTenant(),
				ExpectError: regexp.MustCompile(`uses the access-manager's default Azure credentials; changing tenant_id`),
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
				// elasticache engine with only half of the AWS credential pair.
				Config:      redisAccessCredentialElastiCacheAccessKeyIDOnly(),
				ExpectError: regexp.MustCompile(`requires access_key_id and secret_access_key to both be set`),
			},
			{
				// elasticache engine with only the write-only half of the pair.
				Config:      redisAccessCredentialElastiCacheWOSecretOnly(),
				ExpectError: regexp.MustCompile(`requires access_key_id and secret_access_key to both be set`),
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
			{
				// azure_managed_redis engine, cluster_name (a required field) omitted.
				Config:      redisAccessCredentialAzureMissingRequired(),
				ExpectError: regexp.MustCompile(`engine "azure_managed_redis" requires:.*cluster_name`),
			},
			{
				// azure_managed_redis engine with a connection field (host) set.
				Config:      redisAccessCredentialAzureWithHost(),
				ExpectError: regexp.MustCompile(`engine "azure_managed_redis" does not allow:.*host`),
			},
			{
				// azure_managed_redis engine with only half of the app credential pair.
				Config:      redisAccessCredentialAzureClientIDOnly(),
				ExpectError: regexp.MustCompile(`requires client_id and client_secret to both be set`),
			},
			{
				// azure_managed_redis engine with the other half of the pair.
				Config:      redisAccessCredentialAzureClientSecretOnly(),
				ExpectError: regexp.MustCompile(`requires client_id and client_secret to both be set`),
			},
			{
				// redis engine with an azure-only field set.
				Config:      redisAccessCredentialRedisWithAzureField(),
				ExpectError: regexp.MustCompile(`engine "redis" does not allow:.*tenant_id`),
			},
			{
				// elasticache engine with an azure-only field set.
				Config:      redisAccessCredentialElastiCacheWithAzureField(),
				ExpectError: regexp.MustCompile(`engine "elasticache" does not allow:.*cluster_name`),
			},
			{
				// aiven engine with an azure-only field set.
				Config:      redisAccessCredentialAivenWithAzureField(),
				ExpectError: regexp.MustCompile(`engine "aiven" does not allow:.*tenant_id`),
			},
			{
				// The API stores these as UUIDs and returns them lowercased, so an
				// uppercase value would never converge; reject it up front.
				Config:      redisAccessCredentialAzureUppercaseTenantID(),
				ExpectError: regexp.MustCompile(`tenant_id must be a lowercase UUID`),
			},
			{
				Config:      redisAccessCredentialAzureUppercaseSubscriptionID(),
				ExpectError: regexp.MustCompile(`subscription_id must be a lowercase UUID`),
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

func redisAccessCredentialElastiCacheAccessKeyIDOnly() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "elasticache"
  host           = "my-cluster.cache.amazonaws.com"
  cache_engine   = "valkey"
  region         = "eu-north-1"
  user_group_id  = "my-user-group"
  access_key_id  = "AKIAIOSFODNN7EXAMPLE"
}
`
}

func redisAccessCredentialElastiCacheWOSecretOnly() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name                         = "test-redis-bad"
  deployment_ids               = ["` + mockDeploymentID + `"]
  engine                       = "elasticache"
  host                         = "my-cluster.cache.amazonaws.com"
  cache_engine                 = "valkey"
  region                       = "eu-north-1"
  user_group_id                = "my-user-group"
  secret_access_key_wo         = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  secret_access_key_wo_version = "1"
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

// The hex letters are deliberate: the UUID class is lowercase-only, so a
// digits-only fixture would pass even if the class were narrowed to [0-9].
const (
	redisAzureTenantID       = "1111aaaa-1111-1111-1111-111111111111"
	redisAzureOtherTenantID  = "3333cccc-3333-3333-3333-333333333333"
	redisAzureSubscriptionID = "2222bbbb-2222-2222-2222-222222222222"
	redisAzureClientID       = "4444dddd-4444-4444-4444-444444444444"
	redisAzureOtherClientID  = "5555eeee-5555-5555-5555-555555555555"

	redisAzureOtherSubscriptionID = "6666ffff-6666-6666-6666-666666666666"
)

func redisAccessCredentialAzureStep1() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name            = "test-redis-azure"
  description     = "test azure managed redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_id       = "` + redisAzureClientID + `"
  client_secret   = "test-client-secret-v1"
}
`
}

func redisAccessCredentialAzureStep2() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name            = "test-redis-azure"
  description     = "updated azure managed redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureOtherTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_id       = "` + redisAzureClientID + `"
  client_secret   = "test-client-secret-v2"
}
`
}

// Same app and tenant as step 2, moved to another subscription, resource group
// and cluster, with the step 2 secret left in place.
func redisAccessCredentialAzureStep3() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name            = "test-redis-azure"
  description     = "updated azure managed redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureOtherTenantID + `"
  subscription_id = "` + redisAzureOtherSubscriptionID + `"
  resource_group  = "other-redis-rg"
  cluster_name    = "other-redis-cluster"
  client_id       = "` + redisAzureClientID + `"
  client_secret   = "test-client-secret-v2"
}
`
}

// Same credential as step 1 with the app credentials removed.
func redisAccessCredentialAzureNoAppCredentials() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name            = "test-redis-azure"
  description     = "test azure managed redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
}
`
}

func redisAccessCredentialAzureDefaultCredentials() string {
	return `
resource "hush_redis_access_credential" "azure_fed" {
  name            = "test-redis-azure-fed"
  description     = "test azure managed redis credential with default credentials"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
}
`
}

// Same credential as redisAccessCredentialAzureDefaultCredentials, moved to
// another tenant, still with no app credentials.
func redisAccessCredentialAzureDefaultCredentialsOtherTenant() string {
	return `
resource "hush_redis_access_credential" "azure_fed" {
  name            = "test-redis-azure-fed"
  description     = "test azure managed redis credential with default credentials"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureOtherTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
}
`
}

func redisAccessCredentialWOClientSecretStep1() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name                     = "test-redis-wo-client-secret"
  deployment_ids           = ["` + mockDeploymentID + `"]
  engine                   = "azure_managed_redis"
  tenant_id                = "` + redisAzureTenantID + `"
  subscription_id          = "` + redisAzureSubscriptionID + `"
  resource_group           = "my-redis-rg"
  cluster_name             = "my-redis-cluster"
  client_id                = "` + redisAzureClientID + `"
  client_secret_wo         = "client-secret-v1"
  client_secret_wo_version = "1"
}
`
}

func redisAccessCredentialWOClientSecretStep2() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name                     = "test-redis-wo-client-secret"
  deployment_ids           = ["` + mockDeploymentID + `"]
  engine                   = "azure_managed_redis"
  tenant_id                = "` + redisAzureTenantID + `"
  subscription_id          = "` + redisAzureSubscriptionID + `"
  resource_group           = "my-redis-rg"
  cluster_name             = "my-redis-cluster"
  client_id                = "` + redisAzureClientID + `"
  client_secret_wo         = "client-secret-v2"
  client_secret_wo_version = "2"
}
`
}

// Rebinding client_id, with the fresh secret supplied write-only.
func redisAccessCredentialWOClientSecretRebind() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name                     = "test-redis-wo-client-secret"
  deployment_ids           = ["` + mockDeploymentID + `"]
  engine                   = "azure_managed_redis"
  tenant_id                = "` + redisAzureTenantID + `"
  subscription_id          = "` + redisAzureSubscriptionID + `"
  resource_group           = "my-redis-rg"
  cluster_name             = "my-redis-cluster"
  client_id                = "` + redisAzureOtherClientID + `"
  client_secret_wo         = "client-secret-v3"
  client_secret_wo_version = "3"
}
`
}

// Same as step 1 but pointing at another app, with the original secret left in
// place.
func redisAccessCredentialAzureClientIDChanged() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name            = "test-redis-azure"
  description     = "test azure managed redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_id       = "` + redisAzureOtherClientID + `"
  client_secret   = "test-client-secret-v1"
}
`
}

// Same as step 1 but pointing at another tenant, with the original secret left
// in place.
func redisAccessCredentialAzureTenantIDChanged() string {
	return `
resource "hush_redis_access_credential" "azure" {
  name            = "test-redis-azure"
  description     = "test azure managed redis credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureOtherTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_id       = "` + redisAzureClientID + `"
  client_secret   = "test-client-secret-v1"
}
`
}

func redisAccessCredentialAzureMissingRequired() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name            = "test-redis-bad"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
}
`
}

func redisAccessCredentialAzureWithHost() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name            = "test-redis-bad"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  host            = "should-not-be-here.example.com"
}
`
}

func redisAccessCredentialMigrateRedis() string {
	return `
resource "hush_redis_access_credential" "migrate" {
  name           = "test-redis-migrate"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis"
  host           = "test-redis.example.com"
  password       = "testpassword123"
}
`
}

func redisAccessCredentialMigrateAzureDefaultCredentials() string {
	return `
resource "hush_redis_access_credential" "migrate" {
  name            = "test-redis-migrate"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
}
`
}

func redisAccessCredentialMigrateAzureAppCredentials() string {
	return `
resource "hush_redis_access_credential" "migrate" {
  name            = "test-redis-migrate"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_id       = "` + redisAzureClientID + `"
  client_secret   = "test-client-secret-v1"
}
`
}

func redisAccessCredentialAzureUppercaseTenantID() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name            = "test-redis-bad"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + strings.ToUpper(redisAzureTenantID) + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
}
`
}

func redisAccessCredentialAzureUppercaseSubscriptionID() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name            = "test-redis-bad"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + strings.ToUpper(redisAzureSubscriptionID) + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
}
`
}

func redisAccessCredentialAzureClientIDOnly() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name            = "test-redis-bad"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_id       = "` + redisAzureClientID + `"
}
`
}

func redisAccessCredentialAzureClientSecretOnly() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name            = "test-redis-bad"
  deployment_ids  = ["` + mockDeploymentID + `"]
  engine          = "azure_managed_redis"
  tenant_id       = "` + redisAzureTenantID + `"
  subscription_id = "` + redisAzureSubscriptionID + `"
  resource_group  = "my-redis-rg"
  cluster_name    = "my-redis-cluster"
  client_secret   = "test-client-secret-v1"
}
`
}

func redisAccessCredentialRedisWithAzureField() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "redis"
  host           = "redis.example.com"
  password       = "testpassword123"
  tenant_id      = "` + redisAzureTenantID + `"
}
`
}

func redisAccessCredentialElastiCacheWithAzureField() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "elasticache"
  host           = "my-cluster.cache.amazonaws.com"
  cache_engine   = "valkey"
  region         = "eu-north-1"
  user_group_id  = "my-user-group"
  cluster_name   = "should-not-be-here"
}
`
}

func redisAccessCredentialAivenWithAzureField() string {
	return `
resource "hush_redis_access_credential" "bad" {
  name           = "test-redis-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-valkey-service"
  token          = "test-aiven-token"
  tenant_id      = "` + redisAzureTenantID + `"
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
