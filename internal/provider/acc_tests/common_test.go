package acc_tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	p "github.com/hushsecurity/terraform-provider-hush/internal/provider"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const (
	envHushAPIKeyID     = "HUSH_API_KEY_ID"
	envHushAPIKeySecret = "HUSH_API_KEY_SECRET"
	envHushRealm        = "HUSH_REALM"
	envHushDevBaseURL   = "HUSH_DEV_BASE_URL"

	// Mock values used directly in HCL config strings (compile-time concatenation)
	mockDeploymentID  = "dep-mock-1234"
	mockDeploymentID2 = "dep-mock-5678"
)

var provider *schema.Provider
var mockServer *testutil.MockServer

// mockSetupFuncs collects resource-specific mock setup functions registered
// via init() in each test file. TestMain calls them after creating the mock server.
var mockSetupFuncs []func(ms *testutil.MockServer)

// registerMockSetup queues a function to run after the mock server is created.
// Call from init() in test files to register hooks, seeds, etc.
func registerMockSetup(fn func(ms *testutil.MockServer)) {
	mockSetupFuncs = append(mockSetupFuncs, fn)
}

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"hush": func() (*schema.Provider, error) {
		if provider == nil {
			provider = p.New("dev")()
		}
		return provider, nil
	},
}

// TestMain sets up the mock server for all acceptance tests.
func TestMain(m *testing.M) {
	setEnv("TF_ACC", "1")

	fixtures, err := testutil.LoadFixtures()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load fixtures: %v\n", err)
		os.Exit(1)
	}

	mockServer = testutil.NewMockServer(fixtures)

	// Configure provider to use mock server
	setEnv(envHushDevBaseURL, mockServer.URL())
	setEnv(envHushAPIKeyID, "mock-key-id")
	setEnv(envHushAPIKeySecret, "mock-key-secret")
	setEnv(envHushRealm, "US")

	// Apply resource-specific mock setups registered via init() in test files
	for _, fn := range mockSetupFuncs {
		fn(mockServer)
	}

	code := m.Run()

	mockServer.Close()
	os.Exit(code)
}

// setEnv is a helper that panics on error (acceptable in test setup).
func setEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		panic(fmt.Sprintf("failed to set env %s: %v", key, err))
	}
}

func validateResourceDestroyed(resource, resourcePath string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		c := provider.Meta().(*client.Client)
		resourceType := fmt.Sprintf("hush_%s", resource)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			resourceId := rs.Primary.ID

			var err error
			switch resource {
			case "notification_channel":
				_, err = client.GetNotificationChannel(context.Background(), c, resourceId)
			case "notification_configuration":
				_, err = client.GetNotificationConfiguration(context.Background(), c, resourceId)
			case "deployment":
				_, err = client.GetDeployment(context.Background(), c, resourceId)
			case "access_policy":
				_, err = client.GetAccessPolicy(context.Background(), c, resourceId)
			case "postgres_access_credential":
				_, err = client.GetPostgresAccessCredential(context.Background(), c, resourceId)
			case "postgres_access_privilege":
				_, err = client.GetPostgresAccessPrivilege(context.Background(), c, resourceId)
			case "mongodb_access_credential":
				_, err = client.GetMongoDBAccessCredential(context.Background(), c, resourceId)
			case "mongodb_access_privilege":
				_, err = client.GetMongoDBAccessPrivilege(context.Background(), c, resourceId)
			case "mysql_access_credential":
				_, err = client.GetMySQLAccessCredential(context.Background(), c, resourceId)
			case "mysql_access_privilege":
				_, err = client.GetMySQLAccessPrivilege(context.Background(), c, resourceId)
			case "openai_access_credential":
				_, err = client.GetOpenAIAccessCredential(context.Background(), c, resourceId)
			case "openai_access_privilege":
				_, err = client.GetOpenAIAccessPrivilege(context.Background(), c, resourceId)
			case "mariadb_access_credential":
				_, err = client.GetMariaDBAccessCredential(context.Background(), c, resourceId)
			case "gemini_access_credential":
				_, err = client.GetGeminiAccessCredential(context.Background(), c, resourceId)
			case "grok_access_credential":
				_, err = client.GetGrokAccessCredential(context.Background(), c, resourceId)
			case "grok_access_privilege":
				_, err = client.GetGrokAccessPrivilege(context.Background(), c, resourceId)
			case "redis_access_credential":
				_, err = client.GetRedisAccessCredential(context.Background(), c, resourceId)
			case "redis_access_privilege":
				_, err = client.GetRedisAccessPrivilege(context.Background(), c, resourceId)
			case "bedrock_access_credential":
				_, err = client.GetBedrockAccessCredential(context.Background(), c, resourceId)
			case "apigee_access_credential":
				_, err = client.GetApigeeAccessCredential(context.Background(), c, resourceId)
			case "apigee_access_privilege":
				_, err = client.GetApigeeAccessPrivilege(context.Background(), c, resourceId)
			case "elasticsearch_access_credential":
				_, err = client.GetElasticsearchAccessCredential(context.Background(), c, resourceId)
			case "elasticsearch_access_privilege":
				_, err = client.GetElasticsearchAccessPrivilege(context.Background(), c, resourceId)
			case "rabbitmq_access_credential":
				_, err = client.GetRabbitmqAccessCredential(context.Background(), c, resourceId)
			case "rabbitmq_access_privilege":
				_, err = client.GetRabbitmqAccessPrivilege(context.Background(), c, resourceId)
			case "gcp_sa_access_credential":
				_, err = client.GetGCPSAAccessCredential(context.Background(), c, resourceId)
			case "gcp_sa_access_privilege":
				_, err = client.GetGCPSAAccessPrivilege(context.Background(), c, resourceId)
			case "azure_app_access_credential":
				_, err = client.GetAzureAppAccessCredential(context.Background(), c, resourceId)
			case "azure_app_access_privilege":
				_, err = client.GetAzureAppAccessPrivilege(context.Background(), c, resourceId)
			case "aws_access_key_access_credential":
				_, err = client.GetAWSAccessKeyAccessCredential(context.Background(), c, resourceId)
			case "aws_access_key_access_privilege":
				_, err = client.GetAWSAccessKeyAccessPrivilege(context.Background(), c, resourceId)
			case "snowflake_access_credential":
				_, err = client.GetSnowflakeAccessCredential(context.Background(), c, resourceId)
			case "snowflake_access_privilege":
				_, err = client.GetSnowflakeAccessPrivilege(context.Background(), c, resourceId)
			case "aws_wif_access_credential":
				_, err = client.GetAwsWifAccessCredential(context.Background(), c, resourceId)
			case "gcp_wif_access_credential":
				_, err = client.GetGcpWifAccessCredential(context.Background(), c, resourceId)
			case "gitlab_access_credential":
				_, err = client.GetGitlabAccessCredential(context.Background(), c, resourceId)
			case "gitlab_access_privilege":
				_, err = client.GetGitlabAccessPrivilege(context.Background(), c, resourceId)
			case "datadog_access_credential":
				_, err = client.GetDatadogAccessCredential(context.Background(), c, resourceId)
			case "datadog_access_privilege":
				_, err = client.GetDatadogAccessPrivilege(context.Background(), c, resourceId)
			case "salesforce_access_credential":
				_, err = client.GetSalesforceAccessCredential(context.Background(), c, resourceId)
			case "salesforce_access_privilege":
				_, err = client.GetSalesforceAccessPrivilege(context.Background(), c, resourceId)
			case "sendgrid_access_credential":
				_, err = client.GetSendGridAccessCredential(context.Background(), c, resourceId)
			case "sendgrid_access_privilege":
				_, err = client.GetSendGridAccessPrivilege(context.Background(), c, resourceId)
			default:
				return fmt.Errorf("unknown resource type: %s", resource)
			}

			if err == nil {
				return fmt.Errorf("%s %s still exists", resource, resourceId)
			}
			apiError, ok := err.(*client.APIError)
			if ok && apiError.IsNotFound() {
				return nil
			}
			return fmt.Errorf("failed to verify %s %s was destroyed: %s", resource, resourceId, err)
		}
		return nil
	}
}
