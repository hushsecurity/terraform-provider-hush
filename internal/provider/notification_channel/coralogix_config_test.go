package notification_channel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func coralogixData(t *testing.T, config map[string]any) *schema.ResourceData {
	t.Helper()
	base := map[string]any{"api_key": "cx-key", "region": "eu1"}
	for k, v := range config {
		base[k] = v
	}
	return schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":             "test",
		"enabled":          true,
		"coralogix_config": []any{base},
	})
}

func TestCoralogixConfigSendsTheKeyRegionAndApplication(t *testing.T) {
	channelType, configs, err := getNotificationChannelTypeAndConfig(coralogixData(t, nil))
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	if channelType != client.NotificationChannelTypeCoralogix {
		t.Fatalf("got type %q, want coralogix", channelType)
	}
	want := map[string]any{"type": "coralogix", "api_key": "cx-key", "region": "eu1", "application": "hush"}
	for key, value := range want {
		if configs[0][key] != value {
			t.Errorf("%s: got %v, want %v", key, configs[0][key], value)
		}
	}
}

func TestACoralogixDestinationWithoutAKeyIsRejected(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":             "test",
		"coralogix_config": []any{map[string]any{"region": "eu1"}},
	})

	if _, _, err := getNotificationChannelTypeAndConfig(d); err == nil {
		t.Fatal("a Coralogix destination without a key was sent")
	}
}

// The api returns only the region and application; the read must keep the configured key.
func TestReadKeepsTheConfiguredCoralogixKey(t *testing.T) {
	d := coralogixData(t, map[string]any{"api_key_wo_version": "2", "application": "prod"})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeCoralogix,
		Config: []map[string]any{{"region": "eu1", "application": "prod"}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("coralogix_config").([]any)[0].(map[string]any)
	if config["api_key"] != "cx-key" {
		t.Errorf("the configured key was lost on read: %v", config["api_key"])
	}
	if config["api_key_wo_version"] != "2" {
		t.Errorf("the key version was lost on read: %v", config["api_key_wo_version"])
	}
	if config["region"] != "eu1" || config["application"] != "prod" {
		t.Errorf("read back as %v / %v", config["region"], config["application"])
	}
}
