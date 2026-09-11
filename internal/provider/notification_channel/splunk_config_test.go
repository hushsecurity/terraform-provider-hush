package notification_channel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const hecHost = "https://http-inputs-acme.splunkcloud.com"

func splunkData(t *testing.T, config map[string]any) *schema.ResourceData {
	t.Helper()
	base := map[string]any{
		"url":        hecHost,
		"token":      "hec-token",
		"sourcetype": "hush:notification",
		"tls_verify": true,
	}
	for k, v := range config {
		base[k] = v
	}
	return schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":          "test",
		"enabled":       true,
		"splunk_config": []any{base},
	})
}

func splunkPayload(t *testing.T, config map[string]any) map[string]any {
	t.Helper()
	channelType, configs, err := getNotificationChannelTypeAndConfig(splunkData(t, config))
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	if channelType != client.NotificationChannelTypeSplunk {
		t.Fatalf("got type %q, want splunk", channelType)
	}
	return configs[0]
}

func TestSplunkConfigSendsTheDestination(t *testing.T) {
	payload := splunkPayload(t, map[string]any{
		"index":                "security",
		"onprem_deployment_id": "dep-abcdefghijk",
		"tls_verify":           false,
	})

	want := map[string]any{
		"type":                 "splunk",
		"url":                  hecHost,
		"token":                "hec-token",
		"index":                "security",
		"sourcetype":           "hush:notification",
		"onprem_deployment_id": "dep-abcdefghijk",
		"tls_verify":           false,
	}
	for key, value := range want {
		if payload[key] != value {
			t.Errorf("%s: got %v, want %v", key, payload[key], value)
		}
	}
}

// An absent key means "leave what is stored" to the api, so a removed index
// has to be sent as an explicit null or it can never be removed.
func TestAnUnsetIndexIsSentAsAnExplicitNull(t *testing.T) {
	payload := splunkPayload(t, nil)

	for _, key := range []string{"index", "onprem_deployment_id"} {
		value, ok := payload[key]
		if !ok {
			t.Errorf("%s must be sent even when unset", key)
		}
		if value != nil {
			t.Errorf("%s: got %v, want an explicit null", key, value)
		}
	}
}

func TestASplunkDestinationWithoutATokenIsRejected(t *testing.T) {
	_, _, err := getNotificationChannelTypeAndConfig(splunkData(t, map[string]any{"token": ""}))

	if err == nil {
		t.Fatal("a destination with neither token nor token_wo must fail")
	}
}

// The api stores the bare collector whatever spelling was given, and answers
// with it; none of these may plan as a change.
func TestEverySpellingOfTheCollectorIsTheSameUrl(t *testing.T) {
	for _, spelling := range []string{
		"",
		"/",
		"/services/collector",
		"/services/collector/",
		"/services/collector/event",
		"/services/collector/raw",
	} {
		if !suppressCollectorPath("url", hecHost+"/services/collector", hecHost+spelling, nil) {
			t.Errorf("%q planned as a change against the stored collector", spelling)
		}
	}
	if suppressCollectorPath("url", hecHost+"/services/collector", "https://other.example.com", nil) {
		t.Error("another host must still be a change")
	}
}

// The api never returns the token, and answers with the collector path the
// configuration may not have spelled out; the read must still find it.
func TestReadKeepsTheConfiguredToken(t *testing.T) {
	d := splunkData(t, map[string]any{"token_wo_version": "3"})
	channel := &client.NotificationChannel{
		Type: client.NotificationChannelTypeSplunk,
		Config: []map[string]any{{
			"url":        hecHost + "/services/collector",
			"sourcetype": "hush:notification",
			"verified":   true,
		}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("splunk_config").([]any)[0].(map[string]any)
	if config["token"] != "hec-token" {
		t.Errorf("the configured token was lost on read: %v", config["token"])
	}
	if config["token_wo_version"] != "3" {
		t.Errorf("the token version was lost on read: %v", config["token_wo_version"])
	}
	if config["verified"] != true {
		t.Errorf("verified read back as %v", config["verified"])
	}
}

// The data source shares the read path but has no token attributes to set.
func TestTheDataSourceReadsASplunkChannel(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelDataSourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type: client.NotificationChannelTypeSplunk,
		Config: []map[string]any{{
			"url":        hecHost + "/services/collector",
			"index":      "security",
			"sourcetype": "hush:notification",
			"verified":   true,
		}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("the data source could not read the channel: %v", err)
	}

	config := d.Get("splunk_config").([]any)[0].(map[string]any)
	if config["index"] != "security" {
		t.Errorf("index read back as %v", config["index"])
	}
}

// A destination the api answers without these plans as drift otherwise.
func TestAbsentSplunkFieldsReadBackAtTheirDefaults(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeSplunk,
		Config: []map[string]any{{"url": hecHost + "/services/collector"}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("splunk_config").([]any)[0].(map[string]any)
	if config["sourcetype"] != "hush:notification" {
		t.Errorf("sourcetype read back as %q", config["sourcetype"])
	}
	if config["tls_verify"] != true {
		t.Errorf("tls_verify read back as %v", config["tls_verify"])
	}
}
