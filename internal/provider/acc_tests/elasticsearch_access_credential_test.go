package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestElasticsearchHost = "HUSH_TEST_ELASTICSEARCH_HOST"
const envHushTestElasticsearchUsername = "HUSH_TEST_ELASTICSEARCH_USERNAME"
const envHushTestElasticsearchPassword = "HUSH_TEST_ELASTICSEARCH_PASSWORD"

func testAccElasticsearchAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestElasticsearchHost) == "" {
		t.Fatalf("%s env var must be set", envHushTestElasticsearchHost)
	}
	if os.Getenv(envHushTestElasticsearchUsername) == "" {
		t.Fatalf("%s env var must be set", envHushTestElasticsearchUsername)
	}
	if os.Getenv(envHushTestElasticsearchPassword) == "" {
		t.Fatalf("%s env var must be set", envHushTestElasticsearchPassword)
	}
}

func TestAccResourceElasticsearchAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccElasticsearchAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessCredentialStep1(),
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
				Config: elasticsearchAccessCredentialStep2(),
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
		PreCheck:          func() { testAccElasticsearchAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("elasticsearch_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: elasticsearchAccessCredentialStep1() + elasticsearchAccessCredentialDataSource,
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

func elasticsearchAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	host := os.Getenv(envHushTestElasticsearchHost)
	username := os.Getenv(envHushTestElasticsearchUsername)
	password := os.Getenv(envHushTestElasticsearchPassword)
	return `
resource "hush_elasticsearch_access_credential" "test" {
  name           = "test-es-cred"
  description    = "test elasticsearch credential"
  deployment_ids = ["` + deploymentID + `"]
  host           = "` + host + `"
  port           = 9200
  username       = "` + username + `"
  password       = "` + password + `"
  tls            = false
}
`
}

func elasticsearchAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	host := os.Getenv(envHushTestElasticsearchHost)
	username := os.Getenv(envHushTestElasticsearchUsername)
	password := os.Getenv(envHushTestElasticsearchPassword)
	return `
resource "hush_elasticsearch_access_credential" "test" {
  name           = "test-es-cred-updated"
  description    = "updated elasticsearch credential"
  deployment_ids = ["` + deploymentID + `"]
  host           = "` + host + `"
  port           = 9200
  username       = "` + username + `"
  password       = "` + password + `"
  tls            = false
}
`
}

const elasticsearchAccessCredentialDataSource = `
data "hush_elasticsearch_access_credential" "test" {
  id = hush_elasticsearch_access_credential.test.id
}
`
