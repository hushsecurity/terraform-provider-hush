package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The real API splits a submitted credential off into its secret store
		// and never echoes it back, so the mock must not either -- that is what
		// the provider has to carry across a refresh on its own.
		strip := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			configs, ok := obj["config"].([]any)
			if !ok {
				return nil
			}
			for _, configItem := range configs {
				config, ok := configItem.(map[string]any)
				if !ok {
					continue
				}
				auth, ok := config["auth"].(map[string]any)
				if !ok {
					continue
				}
				for _, secret := range []string{"token", "password", "value"} {
					delete(auth, secret)
				}
			}
			return nil
		}
		ms.OnOperation("notification_channels", testutil.OpCreate, strip)
		ms.OnOperation("notification_channels", testutil.OpUpdate, strip)
	})
}

// A bridge-bound endpoint with a write-only credential: the credential must
// reach the API, must not come back, and must not show up as a diff afterwards.
func TestAccResourceNotificationChannelWebhook_bridgeAndAuth(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: webhookBridgeAuthStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.url",
						"http://splunk.internal:8088/services/collector",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.onprem_deployment_id",
						"dep-abcdefghijk",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.payload_format", "json",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.tls_verify", "false",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.auth.0.type", "header",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.auth.0.name", "Authorization",
					),
					// Write-only: never persisted, so it reads as empty in state.
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.auth.0.credential", "",
					),
				),
			},
			// Re-applying the same configuration must not plan a change, which
			// is what a credential the API cannot return would otherwise cause.
			{
				Config:   webhookBridgeAuthStep1,
				PlanOnly: true,
			},
			// Rotating means bumping the version, not editing anything the API
			// can see.
			{
				Config: webhookBridgeAuthStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_notification_channel.bridged", "webhook_config.0.auth.0.credential_wo_version", "2",
					),
				),
			},
		},
	})
}

// A direct endpoint keeps the defaults it always had, so an existing
// configuration is unaffected by the new fields.
func TestAccResourceNotificationChannelWebhook_defaults(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{
			{
				Config: webhookDefaultsStep,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_notification_channel.plain", "webhook_config.0.payload_format", "text",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.plain", "webhook_config.0.tls_verify", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.plain", "webhook_config.0.onprem_deployment_id", "",
					),
					resource.TestCheckResourceAttr(
						"hush_notification_channel.plain", "webhook_config.0.auth.#", "0",
					),
				),
			},
		},
	})
}

// The schema cannot express these pairings inside a list block, so they are
// checked at plan time instead.
func TestAccResourceNotificationChannelWebhook_credentialPairing(t *testing.T) {
	for name, testCase := range map[string]struct {
		config string
		expect string
	}{
		"write-only without a version": {
			config: webhookAuthMissingVersion,
			expect: "credential_wo requires credential_wo_version",
		},
		"both forms at once": {
			config: webhookAuthBothForms,
			expect: "mutually exclusive",
		},
		"neither form": {
			config: webhookAuthNeitherForm,
			expect: "one of credential or credential_wo is required",
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

// A credential that Terraform cannot resolve until apply is unknown at plan
// time, which is the documented way to feed a write-only argument. Reading it
// as "not set" aborted every such plan.
func TestAccResourceNotificationChannelWebhook_unknownCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("notification_channel", "v1/notification_channels"),
		Steps: []resource.TestStep{{
			Config: webhookUnknownCredential,
			Check: resource.TestCheckResourceAttr(
				"hush_notification_channel.referenced", "webhook_config.0.auth.0.type", "bearer",
			),
		}},
	})
}

const (
	webhookUnknownCredential = `
resource "hush_notification_channel" "src" {
  name = "credential-source"
  webhook_config {
    url = "https://example.com/source"
  }
}

resource "hush_notification_channel" "referenced" {
  name = "referenced-credential"
  webhook_config {
    url = "https://example.com/webhook"
    auth {
      type = "bearer"
      # unknown at plan time (stand-in for random_password.x.result)
      credential_wo         = hush_notification_channel.src.id
      credential_wo_version = "1"
    }
  }
}
`
	webhookAuthMissingVersion = `
resource "hush_notification_channel" "unpaired" {
  name = "unpaired"
  webhook_config {
    url = "https://example.com/webhook"
    auth {
      type          = "bearer"
      credential_wo = "tok-123"
    }
  }
}
`
	webhookAuthBothForms = `
resource "hush_notification_channel" "both" {
  name = "both"
  webhook_config {
    url = "https://example.com/webhook"
    auth {
      type                  = "bearer"
      credential            = "tok-123"
      credential_wo         = "tok-456"
      credential_wo_version = "1"
    }
  }
}
`
	webhookAuthNeitherForm = `
resource "hush_notification_channel" "neither" {
  name = "neither"
  webhook_config {
    url = "https://example.com/webhook"
    auth {
      type = "bearer"
    }
  }
}
`
	webhookBridgeAuthStep1 = `
resource "hush_notification_channel" "bridged" {
  name        = "bridged-webhook"
  description = "through an access bridge"
  webhook_config {
    url                  = "http://splunk.internal:8088/services/collector"
    onprem_deployment_id = "dep-abcdefghijk"
    payload_format       = "json"
    tls_verify           = false
    auth {
      type                  = "header"
      name                  = "Authorization"
      credential_wo         = "Splunk tok-123"
      credential_wo_version = "1"
    }
  }
}
`
	webhookBridgeAuthStep2 = `
resource "hush_notification_channel" "bridged" {
  name        = "bridged-webhook"
  description = "through an access bridge"
  webhook_config {
    url                  = "http://splunk.internal:8088/services/collector"
    onprem_deployment_id = "dep-abcdefghijk"
    payload_format       = "json"
    tls_verify           = false
    auth {
      type                  = "header"
      name                  = "Authorization"
      credential_wo         = "Splunk rotated-456"
      credential_wo_version = "2"
    }
  }
}
`
	webhookDefaultsStep = `
resource "hush_notification_channel" "plain" {
  name        = "plain-webhook"
  description = "unchanged by the new fields"
  webhook_config {
    url = "https://example.com/webhook"
  }
}
`
)
