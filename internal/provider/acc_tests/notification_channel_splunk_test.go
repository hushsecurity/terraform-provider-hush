package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The real API stores the token in its secret store and never echoes
		// it; the provider has to carry it across a refresh on its own.
		strip := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			configs, ok := obj["config"].([]any)
			if !ok {
				return nil
			}
			for _, configItem := range configs {
				if config, ok := configItem.(map[string]any); ok {
					delete(config, "token")
				}
			}
			return nil
		}
		ms.OnOperation("notification_channels", testutil.OpCreate, strip)
		ms.OnOperation("notification_channels", testutil.OpUpdate, strip)
	})
}

// A bridge-bound collector with a write-only token: the token must reach the
// API, must not come back, and must not show up as a diff afterwards.
func TestAccResourceNotificationChannelSplunk_bridgeAndToken(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: splunkBridgeStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_notification_channel.siem", "type", "splunk"),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.siem", "splunk_config.0.url", "https://splunk.internal:8088",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.siem", "splunk_config.0.onprem_deployment_id", "dep-abcdefghijk",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.siem", "splunk_config.0.index", "security",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.siem", "splunk_config.0.sourcetype", "hush:notification",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.siem", "splunk_config.0.tls_verify", "false",
					),
					// Write-only: never persisted, so it reads as empty in state.
					resource.TestCheckResourceAttr(
						"hush_notification_channel.siem", "splunk_config.0.token", "",
					),
				),
			},
			// Re-applying the same configuration must not plan a change, which
			// is what a token the API cannot return would otherwise cause.
			{
				Config:   splunkBridgeStep1,
				PlanOnly: true,
			},
			// Rotating means bumping the version.
			{
				Config: splunkBridgeStep2,
				Check: resource.TestCheckResourceAttr(
					"hush_notification_channel.siem", "splunk_config.0.token_wo_version", "2",
				),
			},
		},
	})
}

// The schema cannot express these pairings inside a list block, so they are
// checked at plan time instead.
func TestAccResourceNotificationChannelSplunk_tokenPairing(t *testing.T) {
	for name, testCase := range map[string]struct {
		config string
		expect string
	}{
		"write-only without a version": {
			config: splunkTokenMissingVersion,
			expect: "token_wo requires token_wo_version",
		},
		"both forms at once": {
			config: splunkTokenBothForms,
			expect: "mutually exclusive",
		},
		"neither form": {
			config: splunkTokenNeitherForm,
			expect: "one of token or token_wo is required",
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
	splunkBridgeStep1 = `
resource "hush_notification_channel" "siem" {
  name = "splunk-hec"
  splunk_config {
    url                  = "https://splunk.internal:8088"
    onprem_deployment_id = "dep-abcdefghijk"
    tls_verify           = false
    index                = "security"
    token_wo             = "hec-token"
    token_wo_version     = "1"
  }
}
`
	splunkBridgeStep2 = `
resource "hush_notification_channel" "siem" {
  name = "splunk-hec"
  splunk_config {
    url                  = "https://splunk.internal:8088"
    onprem_deployment_id = "dep-abcdefghijk"
    tls_verify           = false
    index                = "security"
    token_wo             = "rotated-token"
    token_wo_version     = "2"
  }
}
`
	splunkTokenMissingVersion = `
resource "hush_notification_channel" "unpaired" {
  name = "unpaired"
  splunk_config {
    url      = "https://http-inputs-acme.splunkcloud.com"
    token_wo = "hec-token"
  }
}
`
	splunkTokenBothForms = `
resource "hush_notification_channel" "both" {
  name = "both"
  splunk_config {
    url              = "https://http-inputs-acme.splunkcloud.com"
    token            = "hec-token"
    token_wo         = "hec-token"
    token_wo_version = "1"
  }
}
`
	splunkTokenNeitherForm = `
resource "hush_notification_channel" "neither" {
  name = "neither"
  splunk_config {
    url = "https://http-inputs-acme.splunkcloud.com"
  }
}
`
)
