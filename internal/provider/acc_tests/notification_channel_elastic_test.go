package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The real API stores the key in its secret store and never echoes it.
		strip := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			configs, ok := obj["config"].([]any)
			if !ok {
				return nil
			}
			for _, configItem := range configs {
				if config, ok := configItem.(map[string]any); ok {
					delete(config, "api_key")
				}
			}
			return nil
		}
		ms.OnOperation("notification_channels", testutil.OpCreate, strip)
		ms.OnOperation("notification_channels", testutil.OpUpdate, strip)
	})
}

// A bridge-bound cluster with a write-only key: the key must reach the API,
// must not come back, and must not show up as a diff afterwards.
func TestAccResourceNotificationChannelElastic_bridgeAndKey(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: elasticBridgeStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_channel.es", "type", "elastic"),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.es", "elastic_config.0.url", "https://elastic.internal:9200",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.es", "elastic_config.0.onprem_deployment_id", "dep-abcdefghijk",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.es", "elastic_config.0.index", "hush-audit",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.es", "elastic_config.0.tls_verify", "false",
					),
					// Write-only: never persisted, so it reads as empty in state.
					resource.TestCheckResourceAttr(
						"hush_notification_channel.es", "elastic_config.0.api_key", "",
					),
				),
			},
			{
				Config:   elasticBridgeStep1,
				PlanOnly: true,
			},
			{
				Config: elasticBridgeStep2,
				Check: resource.TestCheckResourceAttr(
					"hush_notification_channel.es", "elastic_config.0.api_key_wo_version", "2",
				),
			},
		},
	})
}

func TestAccResourceNotificationChannelElastic_keyPairing(t *testing.T) {
	for name, testCase := range map[string]struct {
		config string
		expect string
	}{
		"write-only without a version": {
			config: elasticKeyMissingVersion,
			expect: "api_key_wo requires api_key_wo_version",
		},
		"both forms at once": {
			config: elasticKeyBothForms,
			expect: "mutually exclusive",
		},
		"neither form": {
			config: elasticKeyNeitherForm,
			expect: "one of api_key or api_key_wo is required",
		},
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
	elasticBridgeStep1 = `
resource "hush_notification_channel" "es" {
  name = "elastic"
  elastic_config {
    url                  = "https://elastic.internal:9200"
    onprem_deployment_id = "dep-abcdefghijk"
    tls_verify           = false
    index                = "hush-audit"
    api_key_wo           = "es-key"
    api_key_wo_version   = "1"
  }
}
`
	elasticBridgeStep2 = `
resource "hush_notification_channel" "es" {
  name = "elastic"
  elastic_config {
    url                  = "https://elastic.internal:9200"
    onprem_deployment_id = "dep-abcdefghijk"
    tls_verify           = false
    index                = "hush-audit"
    api_key_wo           = "rotated-key"
    api_key_wo_version   = "2"
  }
}
`
	elasticKeyMissingVersion = `
resource "hush_notification_channel" "unpaired" {
  name = "unpaired"
  elastic_config {
    url        = "https://my-deployment.es.eu-west-1.aws.found.io"
    api_key_wo = "es-key"
  }
}
`
	elasticKeyBothForms = `
resource "hush_notification_channel" "both" {
  name = "both"
  elastic_config {
    url                = "https://my-deployment.es.eu-west-1.aws.found.io"
    api_key            = "es-key"
    api_key_wo         = "es-key"
    api_key_wo_version = "1"
  }
}
`
	elasticKeyNeitherForm = `
resource "hush_notification_channel" "neither" {
  name = "neither"
  elastic_config {
    url = "https://my-deployment.es.eu-west-1.aws.found.io"
  }
}
`
)
