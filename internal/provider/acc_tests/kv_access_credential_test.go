package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceKVAccessCredential(t *testing.T) {
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

const kvAccessCredentialDataSource = `
data "hush_kv_access_credential" "test" {
  id = hush_kv_access_credential.test.id
}
`
