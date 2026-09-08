package notification_channel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

// webhookData builds ResourceData holding one webhook_config, the way the sdk
// hands it to the marshalling code.
func webhookData(t *testing.T, config map[string]any) *schema.ResourceData {
	t.Helper()
	base := map[string]any{
		"url":            "https://siem.example.com/collector",
		"method":         "POST",
		"payload_format": "text",
		"tls_verify":     true,
	}
	for k, v := range config {
		base[k] = v
	}
	return schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":           "test",
		"enabled":        true,
		"webhook_config": []any{base},
	})
}

func webhookPayload(t *testing.T, config map[string]any) map[string]any {
	t.Helper()
	channelType, configs, err := getNotificationChannelTypeAndConfig(webhookData(t, config))
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	if channelType != client.NotificationChannelTypeWebhook {
		t.Fatalf("got type %q, want webhook", channelType)
	}
	return configs[0]
}

func TestWebhookConfigSendsTheNewFields(t *testing.T) {
	payload := webhookPayload(t, map[string]any{
		"onprem_deployment_id": "dep-abcdefghijk",
		"payload_format":       "json",
		"tls_verify":           false,
	})

	if payload["onprem_deployment_id"] != "dep-abcdefghijk" {
		t.Errorf("binding not sent: %v", payload["onprem_deployment_id"])
	}
	if payload["payload_format"] != "json" {
		t.Errorf("format not sent: %v", payload["payload_format"])
	}
	if payload["tls_verify"] != false {
		t.Errorf("tls_verify not sent: %v", payload["tls_verify"])
	}
}

// The api inherits a binding by matching on the url, so a config that changes
// the url has nothing to inherit and must repeat the deployment.
func TestTheBindingIsSentEveryTime(t *testing.T) {
	payload := webhookPayload(t, map[string]any{"onprem_deployment_id": "dep-abcdefghijk"})

	if _, ok := payload["onprem_deployment_id"]; !ok {
		t.Error("a bridge-bound endpoint must always repeat its deployment")
	}
}

// An absent key means "inherit what is stored" to the api, so a removed
// binding has to be sent as an explicit null or it can never be removed.
func TestADirectEndpointSendsAnExplicitNullBinding(t *testing.T) {
	payload := webhookPayload(t, nil)

	binding, ok := payload["onprem_deployment_id"]
	if !ok {
		t.Fatal("the binding must be sent even when unset")
	}
	if binding != nil {
		t.Errorf("got %v, want an explicit null", binding)
	}
}

func TestAnUnknownAuthTypeIsRejected(t *testing.T) {
	_, _, err := getNotificationChannelTypeAndConfig(webhookData(t, map[string]any{
		"auth": []any{map[string]any{"type": "mtls", "credential": "x"}},
	}))

	if err == nil {
		t.Fatal("an unhandled type must fail rather than authenticate with nothing")
	}
}

func TestBearerAuthIsSentAsAToken(t *testing.T) {
	payload := webhookPayload(t, map[string]any{
		"auth": []any{map[string]any{"type": "bearer", "credential": "tok-123"}},
	})

	auth, _ := payload["auth"].(map[string]any)
	if auth["type"] != "bearer" || auth["token"] != "tok-123" {
		t.Errorf("got %v, want the credential under token", auth)
	}
}

func TestBasicAuthCarriesItsUsername(t *testing.T) {
	payload := webhookPayload(t, map[string]any{
		"auth": []any{map[string]any{
			"type": "basic", "username": "hush", "credential": "pw",
		}},
	})

	auth, _ := payload["auth"].(map[string]any)
	if auth["username"] != "hush" || auth["password"] != "pw" {
		t.Errorf("got %v, want the credential under password", auth)
	}
}

func TestHeaderAuthCarriesItsName(t *testing.T) {
	payload := webhookPayload(t, map[string]any{
		"auth": []any{map[string]any{
			"type": "header", "name": "X-Splunk-Token", "credential": "abc",
		}},
	})

	auth, _ := payload["auth"].(map[string]any)
	if auth["name"] != "X-Splunk-Token" || auth["value"] != "abc" {
		t.Errorf("got %v, want the credential under value", auth)
	}
}

func TestAuthWithoutACredentialIsRejected(t *testing.T) {
	_, _, err := getNotificationChannelTypeAndConfig(webhookData(t, map[string]any{
		"auth": []any{map[string]any{"type": "bearer"}},
	}))

	if err == nil {
		t.Fatal("an auth block with neither credential nor credential_wo must fail")
	}
}

func TestBasicAuthWithoutAUsernameIsRejected(t *testing.T) {
	_, _, err := getNotificationChannelTypeAndConfig(webhookData(t, map[string]any{
		"auth": []any{map[string]any{"type": "basic", "credential": "pw"}},
	}))

	if err == nil {
		t.Fatal("basic auth without a username must fail")
	}
}

func TestHeaderAuthWithoutANameIsRejected(t *testing.T) {
	_, _, err := getNotificationChannelTypeAndConfig(webhookData(t, map[string]any{
		"auth": []any{map[string]any{"type": "header", "credential": "abc"}},
	}))

	if err == nil {
		t.Fatal("header auth without a name must fail")
	}
}

// Same for the credential: omitting it would leave a stored one live, so
// removing the block has to clear it.
func TestNoAuthBlockClearsTheCredential(t *testing.T) {
	payload := webhookPayload(t, nil)

	auth, ok := payload["auth"]
	if !ok {
		t.Fatal("auth must be sent even when there is no block")
	}
	if auth != nil {
		t.Errorf("got %v, want an explicit null", auth)
	}
}

// The api returns the shape of a credential but never the credential, so a
// read must not blank what the configuration holds.
func TestReadKeepsTheConfiguredCredential(t *testing.T) {
	d := webhookData(t, map[string]any{
		"auth": []any{map[string]any{"type": "bearer", "credential": "tok-123"}},
	})
	channel := &client.NotificationChannel{
		Type: client.NotificationChannelTypeWebhook,
		Config: []map[string]any{{
			"url":      "https://siem.example.com/collector",
			"method":   "POST",
			"verified": true,
			"auth":     map[string]any{"type": "bearer"},
		}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	configs := d.Get("webhook_config").([]any)
	auth := configs[0].(map[string]any)["auth"].([]any)[0].(map[string]any)
	if auth["credential"] != "tok-123" {
		t.Errorf("the configured credential was lost on read: %v", auth["credential"])
	}
}

// The data source shares the read path but has no credential attributes to
// set, so carrying them unconditionally made every webhook channel with a
// credential fail to read.
func TestTheDataSourceReadsAChannelThatHasACredential(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelDataSourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type: client.NotificationChannelTypeWebhook,
		Config: []map[string]any{{
			"url":      "https://siem.example.com/collector",
			"method":   "POST",
			"verified": true,
			"auth":     map[string]any{"type": "bearer"},
		}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("the data source could not read the channel: %v", err)
	}

	configs := d.Get("webhook_config").([]any)
	auth := configs[0].(map[string]any)["auth"].([]any)[0].(map[string]any)
	if auth["type"] != "bearer" {
		t.Errorf("the credential shape was lost: %v", auth)
	}
}

// The api may answer in a different order than it was sent, and a credential
// stored against the wrong endpoint shows up as a diff that never settles.
func TestCredentialsFollowTheirUrlNotTheirPosition(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{
		"name":    "test",
		"enabled": true,
		"webhook_config": []any{
			map[string]any{
				"url":  "https://first.example.com/x",
				"auth": []any{map[string]any{"type": "bearer", "credential": "first-token"}},
			},
			map[string]any{
				"url":  "https://second.example.com/x",
				"auth": []any{map[string]any{"type": "bearer", "credential": "second-token"}},
			},
		},
	})
	// Answered in the opposite order.
	channel := &client.NotificationChannel{
		Type: client.NotificationChannelTypeWebhook,
		Config: []map[string]any{
			{"url": "https://second.example.com/x", "auth": map[string]any{"type": "bearer"}},
			{"url": "https://first.example.com/x", "auth": map[string]any{"type": "bearer"}},
		},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	configs := d.Get("webhook_config").([]any)
	for _, configInterface := range configs {
		config := configInterface.(map[string]any)
		auth := config["auth"].([]any)[0].(map[string]any)
		want := "first-token"
		if config["url"] == "https://second.example.com/x" {
			want = "second-token"
		}
		if auth["credential"] != want {
			t.Errorf("%v got credential %q, want %q", config["url"], auth["credential"], want)
		}
	}
}

// A channel stored before these fields existed carries neither, and a zero
// read back plans as drift on every run.
func TestAbsentFieldsReadBackAtTheirDefaults(t *testing.T) {
	d := schema.TestResourceDataRaw(t, NotificationChannelResourceSchema(), map[string]any{})
	channel := &client.NotificationChannel{
		Type:   client.NotificationChannelTypeWebhook,
		Config: []map[string]any{{"url": "https://siem.example.com/x", "method": "POST"}},
	}

	if err := setNotificationChannelConfigFields(d, channel); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	config := d.Get("webhook_config").([]any)[0].(map[string]any)
	if config["payload_format"] != "text" {
		t.Errorf("payload_format read back as %q, want text", config["payload_format"])
	}
	if config["tls_verify"] != true {
		t.Errorf("tls_verify read back as %v, want true", config["tls_verify"])
	}
}
