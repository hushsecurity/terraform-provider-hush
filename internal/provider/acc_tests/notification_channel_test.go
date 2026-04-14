package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The real API infers the channel type from config contents and returns
		// it in the GET response. The mock must do the same since the provider's
		// Read function switches on 'type' to map config fields to the schema.
		ms.OnOperation("notification_channels", testutil.OpCreate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			if configs, ok := obj["config"].([]any); ok && len(configs) > 0 {
				if first, ok := configs[0].(map[string]any); ok {
					if _, has := first["address"]; has {
						obj["type"] = "email"
					} else if _, has := first["url"]; has {
						obj["type"] = "webhook"
					} else if _, has := first["integration_id"]; has {
						obj["type"] = "slack"
					}
				}
			}
			return nil
		})
	})
}

func TestAccResourceNotificationChannelEmail(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: emailNotificationChannelStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_channel.email", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "name", "email-channel",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "description", "email channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.0.address", "user1@example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.1.address", "user2@example.com",
					),
				),
			},
			{
				Config: emailNotificationChannelStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_channel.email", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "name", "email-channel-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "description", "updated email channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.0.address", "user3@example.com",
					),
				),
			},
		},
	})
}

func TestAccResourceNotificationChannelWebhook(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: webhookNotificationChannelStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_channel.webhook", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "name", "webhook-channel",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "description", "webhook channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.url", "https://example.com/webhook",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.method", "POST",
					),
				),
			},
			{
				Config: webhookNotificationChannelStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_channel.webhook", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "name", "webhook-channel-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "description", "updated webhook channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.url", "https://api.example.com/notifications",
					),
				),
			},
		},
	})
}

func TestAccResourceNotificationChannelSlack(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: slackNotificationChannelStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_channel.slack", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "name", "slack-channel",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "description", "slack channel description",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "slack_config.0.integration_id", "int-euIk8SVlvEGqOzNM5D",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.slack", "slack_config.0.channel", "test-channel",
					),
				),
			},
		},
	})
}

func TestAccDataSourceNotificationChannel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: emailNotificationChannelStep1 + notificationChannelDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_notification_channel.channel", "id", regexp.MustCompile("^nch-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "name", "email-channel",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "description", "email channel description",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "email_config.0.address", "user1@example.com",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "email_config.1.address", "user2@example.com",
					),
				),
			},
		},
	})
}

const (
	emailNotificationChannelStep1 = `
resource "hush_notification_channel" "email" {
  name        = "email-channel"
  description = "email channel description"
  email_config {
    address = "user1@example.com"
  }
  email_config {
    address = "user2@example.com"
  }
}
`

	emailNotificationChannelStep2 = `
resource "hush_notification_channel" "email" {
  name        = "email-channel-updated"
  description = "updated email channel description"
  email_config {
    address = "user3@example.com"
  }
}
`

	webhookNotificationChannelStep1 = `
resource "hush_notification_channel" "webhook" {
  name        = "webhook-channel"
  description = "webhook channel description"
  webhook_config {
    url    = "https://example.com/webhook"
    method = "POST"
  }
}
`

	webhookNotificationChannelStep2 = `
resource "hush_notification_channel" "webhook" {
  name        = "webhook-channel-updated"
  description = "updated webhook channel description"
  webhook_config {
    url    = "https://api.example.com/notifications"
    method = "POST"
  }
}
`

	slackNotificationChannelStep1 = `
resource "hush_notification_channel" "slack" {
  name        = "slack-channel"
  description = "slack channel description"
  slack_config {
    integration_id = "int-euIk8SVlvEGqOzNM5D"
    channel        = "test-channel"
  }
}
`

	notificationChannelDataSource = `
data "hush_notification_channel" "channel" {
  id = hush_notification_channel.email.id
}
`
)
