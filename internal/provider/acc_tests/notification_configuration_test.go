package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNotificationConfiguration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_configuration", "v1/notification_configurations"),
		Steps: []resource.TestStep{
			{
				Config: notificationChannelDependency + notificationConfigurationStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_configuration.config", "id", regexp.MustCompile("^ncf-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "name", "test-config",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "description", "test notification configuration",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "enabled", "true",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"hush_notification_configuration.config", "notification_channels.*",
						"hush_notification_channel.email", "id",
					),
				),
			},
			{
				Config: notificationChannelDependency + notificationConfigurationStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_notification_configuration.config", "id", regexp.MustCompile("^ncf-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "name", "test-config-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "description", "updated notification configuration",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_configuration.config", "enabled", "false",
					),
				),
			},
		},
	})
}

func TestAccDataSourceNotificationConfiguration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_configuration", "v1/notification_configurations"),
		Steps: []resource.TestStep{
			{
				Config: notificationChannelDependency + notificationConfigurationStep1 + notificationConfigurationDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_notification_configuration.config", "id", regexp.MustCompile("^ncf-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "name", "test-config",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "description", "test notification configuration",
					),
					resource.TestCheckResourceAttr(
						"data.hush_notification_configuration.config", "enabled", "true",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"data.hush_notification_configuration.config", "notification_channels.*",
						"hush_notification_channel.email", "id",
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
    recipients = ["admin@example.com"]
  }
}

resource "hush_notification_channel" "webhook" {
  name        = "webhook-channel-for-config"
  description = "webhook channel for notification configuration tests"
  webhook_config {
    url    = "https://api.example.com/notifications"
    method = "POST"
    headers = {
      "Content-Type" = "application/json"
    }
  }
}
`

	notificationConfigurationStep1 = `
resource "hush_notification_configuration" "config" {
  name                  = "test-config"
  description           = "test notification configuration"
  enabled               = true
  notification_channels = [hush_notification_channel.email.id]
}
`

	notificationConfigurationStep2 = `
resource "hush_notification_configuration" "config" {
  name                  = "test-config-updated"
  description           = "updated notification configuration"
  enabled               = false
  notification_channels = [
    hush_notification_channel.email.id,
    hush_notification_channel.webhook.id
  ]
}
`

	notificationConfigurationDataSource = `
data "hush_notification_configuration" "config" {
  id = hush_notification_configuration.config.id
}
`
)
