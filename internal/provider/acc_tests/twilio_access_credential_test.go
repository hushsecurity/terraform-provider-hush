package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTwilioAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("twilio_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: twilioAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_twilio_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_twilio_access_credential.test", "name", "test-twilio-cred",
					),
					checkSecretStoreID("hush_twilio_access_credential.test"),
				),
			},
		},
	})
}

func TestAccDataSourceTwilioAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("twilio_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: twilioAccessCredentialStep1 + twilioAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_twilio_access_credential.test", "name", "test-twilio-cred",
					),
					checkSecretStoreID("data.hush_twilio_access_credential.test"),
				),
			},
		},
	})
}

const twilioAccessCredentialStep1 = `
resource "hush_twilio_access_credential" "test" {
  name            = "test-twilio-cred"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  account_sid     = "AC00000000000000000000000000000000"
  api_key_sid     = "SK00000000000000000000000000000000"
  api_key_secret  = "mock-twilio-secret"
}
`

const twilioAccessCredentialDataSource = `
data "hush_twilio_access_credential" "test" {
  id = hush_twilio_access_credential.test.id
}
`
