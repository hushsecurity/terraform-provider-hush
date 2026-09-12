package notification_channel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const cluster = "https://my-deployment.es.eu-west-1.aws.found.io"

func elasticData(t *testing.T, config map[string]any) *schema.ResourceData {
	t.Helper()
	base := map[string]any{
		"url":        cluster,
		"api_key":    "es-key",
		"index":      "hush-notifications",
		"tls_verify": true,
	}
	for k, v := range config {
		base[k] = v
	}
	return schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":           "test",
		"enabled":        true,
		"elastic_config": []any{base},
	})
}

func elasticPayload(t *testing.T, config map[string]any) map[string]any {
	t.Helper()
	channelType, configs, err := getNotificationChannelTypeAndConfig(elasticData(t, config))
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	if channelType != client.NotificationChannelTypeElastic {
		t.Fatalf("got type %q, want elastic", channelType)
	}
	return configs[0]
}

func TestElasticConfigSendsTheDestination(t *testing.T) {
	payload := elasticPayload(t, map[string]any{
		"index":                "hush-audit",
		"onprem_deployment_id": "dep-abcdefghijk",
		"tls_verify":           false,
	})

	want := map[string]any{
		"type":                 "elastic",
		"url":                  cluster,
		"api_key":              "es-key",
		"index":                "hush-audit",
		"onprem_deployment_id": "dep-abcdefghijk",
		"tls_verify":           false,
	}
	for key, value := range want {
		if payload[key] != value {
			t.Errorf("%s: got %v, want %v", key, payload[key], value)
		}
	}
}

func TestADirectClusterSendsAnExplicitNullBinding(t *testing.T) {
	payload := elasticPayload(t, nil)

	binding, ok := payload["onprem_deployment_id"]
	if !ok {
		t.Fatal("the binding must be sent even when unset")
	}
	if binding != nil {
		t.Errorf("got %v, want an explicit null", binding)
	}
}

func TestAnElasticDestinationWithoutAKeyIsRejected(t *testing.T) {
	_, _, err := getNotificationChannelTypeAndConfig(elasticData(t, map[string]any{"api_key": ""}))

	if err == nil {
		t.Fatal("a destination with neither api_key nor api_key_wo must fail")
	}
}

// The api stores the bare cluster whatever was pasted, and answers with it.
func TestAPastedBulkUrlIsTheSameCluster(t *testing.T) {
	for _, spelling := range []string{
		"",
		"/",
		"/hush-notifications/_bulk",
		"/hush-notifications/_bulk/",
	} {
		if !suppressClusterPath("url", cluster, cluster+spelling, nil) {
			t.Errorf("%q planned as a change against the stored cluster", spelling)
		}
	}
	if suppressClusterPath("url", cluster, "https://other.example.com", nil) {
		t.Error("another cluster must still be a change")
	}
	if elasticClusterURL("https://es.internal:9200/prefix/idx/_bulk") != "https://es.internal:9200/prefix" {
		t.Error("a path prefix in front of the index must survive")
	}
}

// The api never returns the key; the read must still find the configured one.
func TestReadKeepsTheConfiguredKey(t *testing.T) {
	d := elasticData(t, map[string]any{"api_key_wo_version": "3"})
	channel := &client.NotificationChannel{
		Type: client.NotificationChannelTypeElastic,
		Config: []map[string]any{{
			"url":      cluster,
			"index":    "hush-notifications",
			"verified": true,
		}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("elastic_config").([]any)[0].(map[string]any)
	if config["api_key"] != "es-key" {
		t.Errorf("the configured key was lost on read: %v", config["api_key"])
	}
	if config["api_key_wo_version"] != "3" {
		t.Errorf("the key version was lost on read: %v", config["api_key_wo_version"])
	}
}

// The data source shares the read path but has no key attributes to set.
func TestTheDataSourceReadsAnElasticChannel(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelDataSourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeElastic,
		Config: []map[string]any{{"url": cluster, "index": "hush-audit", "verified": true}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("the data source could not read the channel: %v", err)
	}

	config := d.Get("elastic_config").([]any)[0].(map[string]any)
	if config["index"] != "hush-audit" {
		t.Errorf("index read back as %v", config["index"])
	}
}

func TestAbsentElasticFieldsReadBackAtTheirDefaults(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeElastic,
		Config: []map[string]any{{"url": cluster}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("elastic_config").([]any)[0].(map[string]any)
	if config["index"] != "hush-notifications" {
		t.Errorf("index read back as %q", config["index"])
	}
	if config["tls_verify"] != true {
		t.Errorf("tls_verify read back as %v", config["tls_verify"])
	}
}
