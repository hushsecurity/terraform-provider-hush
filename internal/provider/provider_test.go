package provider

import (
	"testing"
)

func TestProvider(t *testing.T) {
	provider := New("test")()

	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	// Test that required schema fields are present
	requiredFields := []string{"api_key_id", "api_key_secret", "realm"}
	for _, field := range requiredFields {
		if _, ok := provider.Schema[field]; !ok {
			t.Errorf("Provider schema missing required field: %s", field)
		}
	}

	// Test that resources are registered
	expectedResources := []string{
		"hush_deployment",
		"hush_notification_channel",
		"hush_notification_configuration",
		"hush_plaintext_access_credential",
		"hush_kv_access_credential",
		"hush_access_policy",
		"hush_postgres_access_credential",
		"hush_postgres_access_privilege",
		"hush_mongodb_access_credential",
		"hush_mongodb_access_privilege",
		"hush_mysql_access_credential",
		"hush_mysql_access_privilege",
		"hush_openai_access_credential",
		"hush_openai_access_privilege",
		"hush_mariadb_access_credential",
		"hush_gemini_access_credential",
		"hush_grok_access_credential",
		"hush_grok_access_privilege",
		"hush_redis_access_credential",
		"hush_redis_access_privilege",
		"hush_snowflake_access_credential",
		"hush_snowflake_access_privilege",
		"hush_bedrock_access_credential",
		"hush_apigee_access_credential",
		"hush_apigee_access_privilege",
		"hush_elasticsearch_access_credential",
		"hush_elasticsearch_access_privilege",
		"hush_rabbitmq_access_credential",
		"hush_rabbitmq_access_privilege",
		"hush_gcp_sa_access_credential",
		"hush_gcp_sa_access_privilege",
		"hush_azure_app_access_credential",
		"hush_azure_app_access_privilege",
		"hush_aws_access_key_access_credential",
		"hush_aws_access_key_access_privilege",
		"hush_twilio_access_credential",
		"hush_twilio_access_privilege",
		"hush_aws_wif_access_credential",
		"hush_gcp_wif_access_credential",
	}
	for _, resource := range expectedResources {
		if _, ok := provider.ResourcesMap[resource]; !ok {
			t.Errorf("Provider missing expected resource: %s", resource)
		}
	}

	// Test that data sources are registered
	expectedDataSources := []string{
		"hush_deployment",
		"hush_notification_channel",
		"hush_notification_configuration",
		"hush_plaintext_access_credential",
		"hush_kv_access_credential",
		"hush_access_policy",
		"hush_postgres_access_credential",
		"hush_postgres_access_privilege",
		"hush_mongodb_access_credential",
		"hush_mongodb_access_privilege",
		"hush_mysql_access_credential",
		"hush_mysql_access_privilege",
		"hush_openai_access_credential",
		"hush_openai_access_privilege",
		"hush_mariadb_access_credential",
		"hush_gemini_access_credential",
		"hush_grok_access_credential",
		"hush_grok_access_privilege",
		"hush_redis_access_credential",
		"hush_redis_access_privilege",
		"hush_snowflake_access_credential",
		"hush_snowflake_access_privilege",
		"hush_bedrock_access_credential",
		"hush_apigee_access_credential",
		"hush_apigee_access_privilege",
		"hush_elasticsearch_access_credential",
		"hush_elasticsearch_access_privilege",
		"hush_rabbitmq_access_credential",
		"hush_rabbitmq_access_privilege",
		"hush_gcp_sa_access_credential",
		"hush_gcp_sa_access_privilege",
		"hush_azure_app_access_credential",
		"hush_azure_app_access_privilege",
		"hush_aws_access_key_access_credential",
		"hush_aws_access_key_access_privilege",
		"hush_twilio_access_credential",
		"hush_twilio_access_privilege",
		"hush_aws_wif_access_credential",
		"hush_gcp_wif_access_credential",
	}
	for _, dataSource := range expectedDataSources {
		if _, ok := provider.DataSourcesMap[dataSource]; !ok {
			t.Errorf("Provider missing expected data source: %s", dataSource)
		}
	}
}

func TestProvider_ConfigureContextFunc(t *testing.T) {
	provider := New("test")()

	if provider.ConfigureContextFunc == nil {
		t.Error("Provider ConfigureContextFunc is nil")
	}
}

func TestProviderValidation(t *testing.T) {
	provider := New("test")()

	// Test that the provider can be validated
	err := provider.InternalValidate()
	if err != nil {
		t.Errorf("Provider validation failed: %v", err)
	}
}
