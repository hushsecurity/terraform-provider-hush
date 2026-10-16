package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// A write-only key on the EU site: it reaches the API (stripped on the way
// back by the elastic test's hook), and a re-plan shows no diff.
func TestAccResourceNotificationChannelDatadog(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: datadogStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_channel.dd", "type", "datadog"),
					resource.TestCheckResourceAttr("hush_notification_channel.dd", "datadog_config.0.site", "datadoghq.eu"),
					resource.TestCheckResourceAttr("hush_notification_channel.dd", "datadog_config.0.api_key", ""),
				),
			},
			{
				Config:   datadogStep1,
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceNotificationChannelDatadog_needsAKey(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{{
			Config:      datadogNoKey,
			ExpectError: regexp.MustCompile("one of api_key or api_key_wo is required"),
		}},
	})
}

func TestAccResourceNotificationChannelWebhook_postOnly(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{{
			Config:      webhookGet,
			ExpectError: regexp.MustCompile(`expected webhook_config.0.method to be one of \["POST"\]`),
		}},
	})
}

const (
	datadogStep1 = `
resource "hush_notification_channel" "dd" {
  name = "datadog"
  datadog_config {
    site               = "datadoghq.eu"
    api_key_wo         = "dd-key"
    api_key_wo_version = "1"
  }
}
`
	datadogNoKey = `
resource "hush_notification_channel" "dd" {
  name = "datadog-no-key"
  datadog_config {
    site = "datadoghq.com"
  }
}
`
	webhookGet = `
resource "hush_notification_channel" "get" {
  name = "get"
  webhook_config {
    url    = "https://api.example.com/hook"
    method = "GET"
  }
}
`
)

// A rotated key and a moved site must reach the API.
func TestAccResourceNotificationChannelDatadog_update(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{Config: datadogStep1},
			{
				Config: datadogStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_channel.dd", "datadog_config.0.site", "us3.datadoghq.com"),
					resource.TestCheckResourceAttr("hush_notification_channel.dd", "datadog_config.0.api_key_wo_version", "2"),
				),
			},
			{Config: datadogStep2, PlanOnly: true},
		},
	})
}

const datadogStep2 = `
resource "hush_notification_channel" "dd" {
  name = "datadog"
  datadog_config {
    site               = "us3.datadoghq.com"
    api_key_wo         = "rotated"
    api_key_wo_version = "2"
  }
}
`
