package acc_tests

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// deployment_ids is capped at a single deployment (schema MaxItems: 1) on every
// access-credential kind. Each config below is an otherwise-valid single-deployment
// config (reused from each kind's acceptance test where one exists); twoDeployments
// adds a second deployment so the only validation error is the cap. A regression
// that drops MaxItems from any kind makes that kind's subtest fail.
func TestAccResourceAccessCredentials_RejectMultipleDeployments(t *testing.T) {
	twoDeployments := func(cfg string) string {
		return strings.Replace(
			cfg,
			`["`+mockDeploymentID+`"]`,
			`["`+mockDeploymentID+`", "`+mockDeploymentID2+`"]`,
			1,
		)
	}

	configs := map[string]string{
		"hush_apigee_access_credential":         apigeeAccessCredentialStep1(),
		"hush_aws_wif_access_credential":        awsWifAccessCredentialStep1(),
		"hush_azure_wif_access_credential":      azureWifAccessCredentialStep1(),
		"hush_bedrock_access_credential":        bedrockAccessCredentialStep1(),
		"hush_datadog_access_credential":        datadogAccessCredentialStep1(),
		"hush_gcp_sa_access_credential":         gcpSAAccessCredentialStep1(),
		"hush_gcp_wif_access_credential":        gcpWifAccessCredentialStep1(),
		"hush_gemini_access_credential":         geminiAccessCredentialStep1(),
		"hush_kafka_access_credential":          kafkaAccessCredentialNativeStep1(),
		"hush_mariadb_access_credential":        mariadbAccessCredentialStep1(),
		"hush_mongodb_access_credential":        mongodbAccessCredentialStep1(),
		"hush_mongodb_atlas_access_credential":  mongodbAtlasAccessCredentialStep1(),
		"hush_mysql_access_credential":          mysqlAccessCredentialStep1(),
		"hush_postgres_access_credential":       postgresAccessCredentialStep1(),
		"hush_redis_access_credential":          redisAccessCredentialStep1(),
		"hush_salesforce_access_credential":     salesforceAccessCredentialStep1(),
		"hush_sendgrid_access_credential":       sendgridAccessCredentialStep1(),
		"hush_snowflake_access_credential":      snowflakeAccessCredentialStep1(),
		"hush_temporal_cloud_access_credential": temporalCloudAccessCredentialStep1(),
		"hush_aws_access_key_access_credential": `
resource "hush_aws_access_key_access_credential" "test" {
  name                = "test-aws-cred"
  description         = "test aws credential"
  deployment_ids      = ["` + mockDeploymentID + `"]
  access_key_id_value = "` + mockAWSAccessKeyID + `"
  secret_access_key   = "` + mockAWSSecretAccessKey + `"
}
`,
		"hush_azure_app_access_credential": `
resource "hush_azure_app_access_credential" "test" {
  name           = "test-azure-cred"
  description    = "test azure credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  tenant_id      = "` + mockAzureTenantID + `"
  client_id      = "` + mockAzureClientID + `"
  client_secret  = "` + mockAzureClientSecret + `"
}
`,
		"hush_elasticsearch_access_credential": `
resource "hush_elasticsearch_access_credential" "test" {
  name           = "test-es-cred"
  description    = "test elasticsearch credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "` + mockElasticsearchHost + `"
  port           = 9200
  username       = "` + mockElasticsearchUsername + `"
  password       = "` + mockElasticsearchPassword + `"
  tls            = false
}
`,
		"hush_gitlab_access_credential": `
resource "hush_gitlab_access_credential" "test" {
  name           = "test-gitlab-cred"
  description    = "test gitlab credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  token          = "` + mockGitlabToken + `"
  resource_type  = "group"
  resource_id    = "` + mockGitlabResourceID + `"
}
`,
		"hush_grok_access_credential": `
resource "hush_grok_access_credential" "test" {
  name           = "test-grok-cred"
  description    = "test grok credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "` + mockGrokAPIKey + `"
  team_id        = "` + mockGrokTeamID + `"
}
`,
		"hush_openai_access_credential": `
resource "hush_openai_access_credential" "test" {
  name           = "test-openai-cred"
  description    = "test openai credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "` + mockOpenAIAPIKey + `"
  project_id     = "` + mockOpenAIProjectID + `"
}
`,
		"hush_rabbitmq_access_credential": `
resource "hush_rabbitmq_access_credential" "test" {
  name            = "test-rabbitmq-cred"
  description     = "test rabbitmq credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  host            = "` + mockRabbitmqHost + `"
  port            = 5672
  management_port = 15672
  username        = "` + mockRabbitmqUsername + `"
  password        = "` + mockRabbitmqPassword + `"
  vhost           = "/"
  tls             = false
}
`,
		"hush_kv_access_credential": `
resource "hush_kv_access_credential" "test" {
  name           = "test-kv-cred"
  deployment_ids = ["` + mockDeploymentID + `"]
  items {
    key   = "k"
    value = "v"
  }
}
`,
		"hush_plaintext_access_credential": `
resource "hush_plaintext_access_credential" "test" {
  name           = "test-plaintext-cred"
  deployment_ids = ["` + mockDeploymentID + `"]
  secret         = "s3cret"
}
`,
		"hush_twilio_access_credential": `
resource "hush_twilio_access_credential" "test" {
  name           = "test-twilio-cred"
  deployment_ids = ["` + mockDeploymentID + `"]
  account_sid    = "AC00000000000000000000000000000000"
  api_key_sid    = "SK00000000000000000000000000000000"
  api_key_secret = "test-twilio-api-key-secret"
}
`,
	}

	for resourceType, cfg := range configs {
		t.Run(resourceType, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{
					{
						Config:      twoDeployments(cfg),
						ExpectError: regexp.MustCompile(`supports 1 item maximum`),
					},
				},
			})
		})
	}
}
