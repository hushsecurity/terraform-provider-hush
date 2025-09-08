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
