package notification_channel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func datadogData(t *testing.T, config map[string]any) *schema.ResourceData {
	t.Helper()
	base := map[string]any{"api_key": "dd-key", "site": "datadoghq.eu"}
	for k, v := range config {
		base[k] = v
	}
	return schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":           "test",
		"enabled":        true,
		"datadog_config": []any{base},
	})
}

func TestDatadogConfigSendsTheKeyAndSite(t *testing.T) {
	channelType, configs, err := getNotificationChannelTypeAndConfig(datadogData(t, nil))
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	if channelType != client.NotificationChannelTypeDatadog {
		t.Fatalf("got type %q, want datadog", channelType)
	}
	want := map[string]any{"type": "datadog", "api_key": "dd-key", "site": "datadoghq.eu"}
	for key, value := range want {
		if configs[0][key] != value {
			t.Errorf("%s: got %v, want %v", key, configs[0][key], value)
		}
	}
}

func TestADatadogDestinationWithoutAKeyIsRejected(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":           "test",
		"datadog_config": []any{map[string]any{"site": "datadoghq.com"}},
	})

	if _, _, err := getNotificationChannelTypeAndConfig(d); err == nil {
		t.Fatal("a Datadog destination without a key was sent")
	}
}

// The api returns only the site; the read must keep the configured key.
func TestReadKeepsTheConfiguredDatadogKey(t *testing.T) {
	d := datadogData(t, map[string]any{"api_key_wo_version": "2"})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeDatadog,
		Config: []map[string]any{{"site": "datadoghq.eu"}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("datadog_config").([]any)[0].(map[string]any)
	if config["api_key"] != "dd-key" {
		t.Errorf("the configured key was lost on read: %v", config["api_key"])
	}
	if config["api_key_wo_version"] != "2" {
		t.Errorf("the key version was lost on read: %v", config["api_key_wo_version"])
	}
	if config["site"] != "datadoghq.eu" {
		t.Errorf("site read back as %v", config["site"])
	}
}

func TestTheDataSourceReadsADatadogChannel(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelDataSourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeDatadog,
		Config: []map[string]any{{}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("the data source could not read the channel: %v", err)
	}

	config := d.Get("datadog_config").([]any)[0].(map[string]any)
	if config["site"] != "datadoghq.com" {
		t.Errorf("an absent site read back as %v", config["site"])
	}
}
