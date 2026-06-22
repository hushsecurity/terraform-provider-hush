package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceKafkaAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessCredentialNativeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_kafka_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "name", "test-kafka-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "description", "test kafka credential",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "engine", "native",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "bootstrap_servers", "broker1:9092,broker2:9092",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "username", "admin",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "sasl_mechanism", "SCRAM-SHA-512",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "tls", "true",
					),
				),
			},
			{
				Config: kafkaAccessCredentialNativeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_kafka_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "name", "test-kafka-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "description", "updated kafka credential",
					),
				),
			},
		},
	})
}

// Exercises the Aiven engine branch (project/service_name/token), the other
// valid engine not covered by the native happy path.
func TestAccResourceKafkaAccessCredential_Aiven(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessCredentialAivenStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_kafka_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "engine", "aiven",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "project", "my-aiven-project",
					),
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "service_name", "my-kafka-service",
					),
				),
			},
			{
				Config: kafkaAccessCredentialAivenStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_kafka_access_credential.test", "description", "updated aiven kafka credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceKafkaAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessCredentialNativeStep1() + kafkaAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_kafka_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_kafka_access_credential.test", "name", "test-kafka-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_kafka_access_credential.test", "engine", "native",
					),
					resource.TestCheckResourceAttr(
						"data.hush_kafka_access_credential.test", "bootstrap_servers", "broker1:9092,broker2:9092",
					),
				),
			},
		},
	})
}

// Write-only secret rotation for the native engine's password. Bumping
// password_wo_version must trigger Update and converge with no perpetual diff.
func TestAccResourceKafkaAccessCredential_WOPasswordRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessCredentialWOPasswordStep1(),
				Check: resource.TestCheckResourceAttr(
					"hush_kafka_access_credential.test", "password_wo_version", "1",
				),
			},
			{
				Config: kafkaAccessCredentialWOPasswordStep2(),
				Check: resource.TestCheckResourceAttr(
					"hush_kafka_access_credential.test", "password_wo_version", "2",
				),
			},
		},
	})
}

// Write-only secret rotation for the Aiven engine's token.
func TestAccResourceKafkaAccessCredential_WOTokenRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessCredentialWOTokenStep1(),
				Check: resource.TestCheckResourceAttr(
					"hush_kafka_access_credential.test", "token_wo_version", "1",
				),
			},
			{
				Config: kafkaAccessCredentialWOTokenStep2(),
				Check: resource.TestCheckResourceAttr(
					"hush_kafka_access_credential.test", "token_wo_version", "2",
				),
			},
		},
	})
}

// Negative tests: every branch of validateEngineFields (CustomizeDiff). Each
// fails at plan time, before any request reaches the mock.
func TestAccResourceKafkaAccessCredential_EngineFieldValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// native engine, password (a required field) omitted.
				Config:      kafkaAccessCredentialNativeMissingRequired(),
				ExpectError: regexp.MustCompile(`engine "native" requires:.*password`),
			},
			{
				// native engine with an aiven-only field set.
				Config:      kafkaAccessCredentialNativeWithAivenField(),
				ExpectError: regexp.MustCompile(`engine "native" does not allow:.*project`),
			},
			{
				// aiven engine, token (a required field) omitted.
				Config:      kafkaAccessCredentialAivenMissingRequired(),
				ExpectError: regexp.MustCompile(`engine "aiven" requires:.*token`),
			},
			{
				// aiven engine with a native-only field set.
				Config:      kafkaAccessCredentialAivenWithNativeField(),
				ExpectError: regexp.MustCompile(`engine "aiven" does not allow:.*bootstrap_servers`),
			},
		},
	})
}

// Negative test: deployment_ids is immutable (credutil.ForbidDeploymentIDsChange).
func TestAccResourceKafkaAccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kafka_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kafkaAccessCredentialNativeStep1(),
			},
			{
				Config:      kafkaAccessCredentialDeploymentChanged(),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func kafkaAccessCredentialNativeStep1() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name              = "test-kafka-cred"
  description       = "test kafka credential"
  deployment_ids    = ["` + mockDeploymentID + `"]
  engine            = "native"
  bootstrap_servers = "broker1:9092,broker2:9092"
  username          = "admin"
  sasl_mechanism    = "SCRAM-SHA-512"
  tls               = true
  password          = "TestPassword123!"
}
`
}

func kafkaAccessCredentialNativeStep2() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name              = "test-kafka-cred-updated"
  description       = "updated kafka credential"
  deployment_ids    = ["` + mockDeploymentID + `"]
  engine            = "native"
  bootstrap_servers = "broker1:9092,broker2:9092"
  username          = "admin"
  sasl_mechanism    = "SCRAM-SHA-512"
  tls               = true
  password          = "TestPassword123!"
}
`
}

func kafkaAccessCredentialAivenStep1() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name           = "test-kafka-aiven"
  description     = "test aiven kafka credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-kafka-service"
  token          = "test-aiven-token"
}
`
}

func kafkaAccessCredentialAivenStep2() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name           = "test-kafka-aiven"
  description     = "updated aiven kafka credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-kafka-service"
  token          = "test-aiven-token"
}
`
}

func kafkaAccessCredentialWOPasswordStep1() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name                = "test-kafka-wo-password"
  deployment_ids      = ["` + mockDeploymentID + `"]
  engine              = "native"
  bootstrap_servers   = "broker1:9092"
  username            = "admin"
  sasl_mechanism      = "PLAIN"
  password_wo         = "secret-v1"
  password_wo_version = "1"
}
`
}

func kafkaAccessCredentialWOPasswordStep2() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name                = "test-kafka-wo-password"
  deployment_ids      = ["` + mockDeploymentID + `"]
  engine              = "native"
  bootstrap_servers   = "broker1:9092"
  username            = "admin"
  sasl_mechanism      = "PLAIN"
  password_wo         = "secret-v2"
  password_wo_version = "2"
}
`
}

func kafkaAccessCredentialWOTokenStep1() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name             = "test-kafka-wo-token"
  deployment_ids   = ["` + mockDeploymentID + `"]
  engine           = "aiven"
  project          = "my-aiven-project"
  service_name     = "my-kafka-service"
  token_wo         = "token-v1"
  token_wo_version = "1"
}
`
}

func kafkaAccessCredentialWOTokenStep2() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name             = "test-kafka-wo-token"
  deployment_ids   = ["` + mockDeploymentID + `"]
  engine           = "aiven"
  project          = "my-aiven-project"
  service_name     = "my-kafka-service"
  token_wo         = "token-v2"
  token_wo_version = "2"
}
`
}

func kafkaAccessCredentialNativeMissingRequired() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name              = "test-kafka-bad"
  deployment_ids    = ["` + mockDeploymentID + `"]
  engine            = "native"
  bootstrap_servers = "broker1:9092"
  username          = "admin"
  sasl_mechanism    = "PLAIN"
}
`
}

func kafkaAccessCredentialNativeWithAivenField() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name              = "test-kafka-bad"
  deployment_ids    = ["` + mockDeploymentID + `"]
  engine            = "native"
  bootstrap_servers = "broker1:9092"
  username          = "admin"
  sasl_mechanism    = "PLAIN"
  password          = "TestPassword123!"
  project           = "should-not-be-here"
}
`
}

func kafkaAccessCredentialAivenMissingRequired() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name           = "test-kafka-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-kafka-service"
}
`
}

func kafkaAccessCredentialAivenWithNativeField() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name              = "test-kafka-bad"
  deployment_ids    = ["` + mockDeploymentID + `"]
  engine            = "aiven"
  project           = "my-aiven-project"
  service_name      = "my-kafka-service"
  token             = "test-aiven-token"
  bootstrap_servers = "should-not-be-here:9092"
}
`
}

// Identical to the native step 1 except deployment_ids, to isolate the
// immutability check.
func kafkaAccessCredentialDeploymentChanged() string {
	return `
resource "hush_kafka_access_credential" "test" {
  name              = "test-kafka-cred"
  description       = "test kafka credential"
  deployment_ids    = ["` + mockDeploymentID2 + `"]
  engine            = "native"
  bootstrap_servers = "broker1:9092,broker2:9092"
  username          = "admin"
  sasl_mechanism    = "SCRAM-SHA-512"
  tls               = true
  password          = "TestPassword123!"
}
`
}

const kafkaAccessCredentialDataSource = `
data "hush_kafka_access_credential" "test" {
  id = hush_kafka_access_credential.test.id
}
`
