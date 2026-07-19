package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceKVAccessCredential(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kv_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kvAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_kv_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_kv_access_credential.test", "name", "test-kv-cred",
					),
					checkSecretStoreID("hush_kv_access_credential.test"),
					recordID("hush_kv_access_credential.test", &id),
				),
			},
			{
				// Changing items is an in-place update, not a replacement.
				Config: kvAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_kv_access_credential.test", "items.0.value", "xyz789",
					),
					checkIDUnchanged("hush_kv_access_credential.test", &id),
				),
			},
		},
	})
}

func TestAccDataSourceKVAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("kv_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: kvAccessCredentialStep1 + kvAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_kv_access_credential.test", "name", "test-kv-cred",
					),
					checkSecretStoreID("data.hush_kv_access_credential.test"),
				),
			},
		},
	})
}

const kvAccessCredentialStep1 = `
resource "hush_kv_access_credential" "test" {
  name            = "test-kv-cred"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"

  items {
    key   = "API_KEY"
    value = "abc123"
  }
}
`

const kvAccessCredentialStep2 = `
resource "hush_kv_access_credential" "test" {
  name            = "test-kv-cred"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"

  items {
    key   = "API_KEY"
    value = "xyz789"
  }
}
`

const kvAccessCredentialDataSource = `
data "hush_kv_access_credential" "test" {
  id = hush_kv_access_credential.test.id
}
`
