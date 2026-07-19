package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourcePlaintextAccessCredential(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("plaintext_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: plaintextAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_plaintext_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_plaintext_access_credential.test", "name", "test-plaintext-cred",
					),
					checkSecretStoreID("hush_plaintext_access_credential.test"),
					recordID("hush_plaintext_access_credential.test", &id),
				),
			},
			{
				// Rotating the secret is an in-place update, not a replacement.
				Config: plaintextAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_plaintext_access_credential.test", "secret", "rotated-secret-value",
					),
					checkIDUnchanged("hush_plaintext_access_credential.test", &id),
				),
			},
		},
	})
}

func TestAccDataSourcePlaintextAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("plaintext_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: plaintextAccessCredentialStep1 + plaintextAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_plaintext_access_credential.test", "name", "test-plaintext-cred",
					),
					checkSecretStoreID("data.hush_plaintext_access_credential.test"),
				),
			},
		},
	})
}

const plaintextAccessCredentialStep1 = `
resource "hush_plaintext_access_credential" "test" {
  name            = "test-plaintext-cred"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  secret          = "s3cr3t-value"
}
`

const plaintextAccessCredentialStep2 = `
resource "hush_plaintext_access_credential" "test" {
  name            = "test-plaintext-cred"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  secret          = "rotated-secret-value"
}
`

const plaintextAccessCredentialDataSource = `
data "hush_plaintext_access_credential" "test" {
  id = hush_plaintext_access_credential.test.id
}
`
