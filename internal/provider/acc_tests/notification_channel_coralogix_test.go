package acc_tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// A write-only key: it reaches the API, doesn't come back, and a re-plan shows
// no diff; a rotated key and a new application are sent on update.
func TestAccResourceNotificationChannelCoralogix(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: coralogixStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_channel.cx", "type", "coralogix"),
					resource.TestCheckResourceAttr("hush_notification_channel.cx", "coralogix_config.0.region", "eu1"),
					resource.TestCheckResourceAttr("hush_notification_channel.cx", "coralogix_config.0.application", "hush"),
					resource.TestCheckResourceAttr("hush_notification_channel.cx", "coralogix_config.0.api_key", ""),
				),
			},
			{Config: coralogixStep1, PlanOnly: true},
			{
				Config: coralogixStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_channel.cx", "coralogix_config.0.application", "prod"),
					resource.TestCheckResourceAttr("hush_notification_channel.cx", "coralogix_config.0.api_key_wo_version", "2"),
				),
			},
			{Config: coralogixStep2, PlanOnly: true},
		},
	})
}

const (
	coralogixStep1 = `
resource "hush_notification_channel" "cx" {
  name = "coralogix"
  coralogix_config {
    region             = "eu1"
    api_key_wo         = "cx-key"
    api_key_wo_version = "1"
  }
}
`
	coralogixStep2 = `
resource "hush_notification_channel" "cx" {
  name = "coralogix"
  coralogix_config {
    region             = "eu1"
    application        = "prod"
    api_key_wo         = "rotated"
    api_key_wo_version = "2"
  }
}
`
)
