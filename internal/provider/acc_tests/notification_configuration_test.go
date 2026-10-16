package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// Pre-seed notification configurations. These resources use an adopt-via-Read+Update
		// Create pattern (not POST), so the mock must have them already in the store.
		ms.SeedObject("notification_configurations", "ncf-mock-config-1", map[string]any{
			"id":          "ncf-mock-config-1",
			"config_id":   "ncf-mock-config-1",
			"name":        "Mock Test Notification",
			"description": "Pre-seeded notification configuration for testing",
			"enabled":     true,
			"channel_ids": []any{},
			"status":      "ok",
		})
		ms.SeedObject("notification_configurations", "ncf-mock-config-2", map[string]any{
			"id":          "ncf-mock-config-2",
			"config_id":   "ncf-mock-config-2",
			"name":        "Mock Test Notification 2",
			"description": "Pre-seeded notification configuration 2 for testing",
			"enabled":     true,
			"channel_ids": []any{},
			"status":      "ok",
		})
	})
}

func TestAccResourceNotificationConfiguration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		// No CheckDestroy — notification configurations are predefined and cannot be deleted.
		// The provider's delete resets them (PATCH enabled=false, channel_ids=[]).
		Steps: []resource.TestStep{
			{
				Config: notificationChannelDependency + notificationConfigurationStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "id", "ncf-mock-config-1",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "name", "Mock Test Notification",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "channel_ids.#", "1",
					),
				),
			},
			{
				Config: notificationChannelDependency + notificationConfigurationStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "id", "ncf-mock-config-1",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "enabled", "false",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "channel_ids.#", "2",
					),
				),
			},
		},
	})
}

func TestAccDataSourceNotificationConfiguration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: notificationChannelDependency + notificationConfigurationDSStep1 + notificationConfigurationDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "id", "ncf-mock-config-2",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "name", "Mock Test Notification 2",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "channel_ids.#", "1",
					),
				),
			},
		},
	})
}

const (
	notificationChannelDependency = `
resource "hush_notification_channel" "email" {
  name        = "email-channel-for-config"
  description = "email channel for notification configuration tests"
  email_config {
    address = "admin@example.com"
  }
}

resource "hush_notification_channel" "webhook" {
  name        = "webhook-channel-for-config"
  description = "webhook channel for notification configuration tests"
  webhook_config {
    url    = "https://api.example.com/notifications"
    method = "POST"
  }
}
`

	notificationConfigurationStep1 = `
resource "hush_notification_configuration" "config" {
  config_id   = "ncf-mock-config-1"
  enabled     = true
  channel_ids = [hush_notification_channel.email.id]
}
`

	notificationConfigurationStep2 = `
resource "hush_notification_configuration" "config" {
  config_id   = "ncf-mock-config-1"
  enabled     = false
  channel_ids = [
    hush_notification_channel.email.id,
    hush_notification_channel.webhook.id
  ]
}
`

	notificationConfigurationDSStep1 = `
resource "hush_notification_configuration" "config" {
  config_id   = "ncf-mock-config-2"
  enabled     = true
  channel_ids = [hush_notification_channel.email.id]
}
`

	notificationConfigurationDataSource = `
data "hush_notification_configuration" "config" {
  id = hush_notification_configuration.config.id
}
`
)

// Without config_id the resource creates the configuration, edits it in place
// and deletes it.
func TestAccResourceNotificationConfiguration_stream(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_configuration", "v1/notification_configurations"),
		Steps: []resource.TestStep{
			{
				Config: notificationChannelDependency + streamStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "trigger", "agw_activity"),
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "name", "activity"),
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "dispatch_filters.0.name", "deployment"),
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "dispatch_filters.0.values.0", mockDeploymentID),
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "channel_ids.#", "1"),
				),
			},
			{
				Config:   notificationChannelDependency + streamStep1,
				PlanOnly: true,
			},
			{
				Config: notificationChannelDependency + streamStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "name", "activity renamed"),
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "dispatch_filters.#", "0"),
					resource.TestCheckResourceAttr("hush_notification_configuration.stream", "channel_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNotificationConfiguration_planChecks(t *testing.T) {
	for name, testCase := range map[string]struct {
		config string
		expect string
	}{
		"no name or trigger":                 {streamNoTrigger, "name and trigger are required without config_id"},
		"aggregation on a stream":            {streamWithAggregation, "aggregation is fixed for trigger agw_activity"},
		"certificate without aggregation":    {certificateNoAggregation, `aggregation \(day or week\) is required`},
		"name on a predefined configuration": {predefinedWithName, "name is the predefined configuration's own"},
		"certificate by month":               {certificateMonth, "aggregation for trigger certificate_expiring is day or week"},
		"nhi_digest is predefined only":      {digestCreated, `expected trigger to be one of`},
	} {
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config:      testCase.config,
					ExpectError: regexp.MustCompile(testCase.expect),
				}},
			})
		})
	}
}

const (
	streamStep1 = `
resource "hush_notification_configuration" "stream" {
  name        = "activity"
  trigger     = "agw_activity"
  enabled     = true
  channel_ids = [hush_notification_channel.webhook.id]
  dispatch_filters {
    name   = "deployment"
    op     = "eq"
    values = ["` + mockDeploymentID + `"]
  }
}
`
	streamStep2 = `
resource "hush_notification_configuration" "stream" {
  name        = "activity renamed"
  trigger     = "agw_activity"
  enabled     = true
  channel_ids = []
}
`
	streamNoTrigger = `
resource "hush_notification_configuration" "bad" {
  enabled = true
}
`
	streamWithAggregation = `
resource "hush_notification_configuration" "bad" {
  name        = "bad"
  trigger     = "agw_activity"
  aggregation = "day"
}
`
	certificateNoAggregation = `
resource "hush_notification_configuration" "bad" {
  name    = "bad"
  trigger = "certificate_expiring"
}
`
	predefinedWithName = `
resource "hush_notification_configuration" "bad" {
  config_id = "ncf-mock-config-1"
  name      = "renamed"
}
`
	digestCreated = `
resource "hush_notification_configuration" "bad" {
  name    = "bad"
  trigger = "nhi_digest"
}
`
	certificateMonth = `
resource "hush_notification_configuration" "bad" {
  name        = "bad"
  trigger     = "certificate_expiring"
  aggregation = "month"
}
`
)
