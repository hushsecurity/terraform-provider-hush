package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockElasticsearchHost = "https://mock-es.example.com:9200"
const mockElasticsearchUsername = "mock-es-user"
const mockElasticsearchPassword = "mock-es-password"

func TestAccResourceElasticsearchAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_elasticsearch_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_credential.test", "name", "test-es-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_credential.test", "description", "test elasticsearch credential",
					),
				),
			},
			{
				Config: elasticsearchAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_elasticsearch_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_credential.test", "name", "test-es-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_elasticsearch_access_credential.test", "description", "updated elasticsearch credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceElasticsearchAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessCredentialStep1 + elasticsearchAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_elasticsearch_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_elasticsearch_access_credential.test", "name", "test-es-cred",
					),
				),
			},
		},
	})
}

const elasticsearchAccessCredentialStep1 = `
resource "hush_elasticsearch_access_credential" "test" {
  name           = "test-es-cred"
  description    = "test elasticsearch credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "` + mockElasticsearchHost + `"
  port           = 9200
  username       = "` + mockElasticsearchUsername + `"
  password       = "` + mockElasticsearchPassword + `"
  tls            = false
}
`

const elasticsearchAccessCredentialStep2 = `
resource "hush_elasticsearch_access_credential" "test" {
  name           = "test-es-cred-updated"
  description    = "updated elasticsearch credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "` + mockElasticsearchHost + `"
  port           = 9200
  username       = "` + mockElasticsearchUsername + `"
  password       = "` + mockElasticsearchPassword + `"
  tls            = false
}
`

const elasticsearchAccessCredentialDataSource = `
data "hush_elasticsearch_access_credential" "test" {
  id = hush_elasticsearch_access_credential.test.id
}
`
