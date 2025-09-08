package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNotificationChannelEmail(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
						"hush_notification_channel.email", "email_config.0.recipients.0", "user1@example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.email", "email_config.0.recipients.1", "user2@example.com",
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
						"hush_notification_channel.email", "email_config.0.recipients.0", "user3@example.com",
					),
				),
			},
		},
	})
}

func TestAccResourceNotificationChannelWebhook(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
					resource.TestCheckResourceAttr(
						"hush_notification_channel.webhook", "webhook_config.0.headers.Content-Type", "application/json",
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
		PreCheck:          func() { testAccPreCheck(t) },
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
						"hush_notification_channel.slack", "slack_config.0.channel_name", "test-channel",
					),
				),
			},
		},
	})
}

func TestAccDataSourceNotificationChannel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
						"data.hush_notification_channel.channel", "email_config.0.recipients.0", "user1@example.com",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_channel.channel", "email_config.0.recipients.1", "user2@example.com",
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
    recipients = ["user1@example.com", "user2@example.com"]
  }
}
`

	emailNotificationChannelStep2 = `
resource "hush_notification_channel" "email" {
  name        = "email-channel-updated"
  description = "updated email channel description"
  email_config {
    recipients = ["user3@example.com"]
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
    headers = {
      "Content-Type" = "application/json"
    }
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
    headers = {
      "Content-Type" = "application/json"
      "Authorization" = "Bearer token"
    }
  }
}
`

	slackNotificationChannelStep1 = `
resource "hush_notification_channel" "slack" {
  name        = "slack-channel"
  description = "slack channel description"
  slack_config {
    integration_id = "int-euIk8SVlvEGqOzNM5D"
    channel_name   = "test-channel"
  }
}
`

	notificationChannelDataSource = `
data "hush_notification_channel" "channel" {
  id = hush_notification_channel.email.id
}
`
)
