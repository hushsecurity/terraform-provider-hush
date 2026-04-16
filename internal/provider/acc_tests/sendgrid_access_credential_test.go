package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestSendGridAPIKey = "HUSH_TEST_SENDGRID_API_KEY"

func testAccSendGridAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestSendGridAPIKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestSendGridAPIKey)
	}
}

func TestAccResourceSendGridAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccSendGridAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("sendgrid_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: sendgridAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_sendgrid_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_credential.test", "name", "test-sendgrid-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_credential.test", "description", "test sendgrid credential",
					),
				),
			},
			{
				Config: sendgridAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_sendgrid_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_credential.test", "name", "test-sendgrid-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_sendgrid_access_credential.test", "description", "updated sendgrid credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSendGridAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSendGridAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("sendgrid_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: sendgridAccessCredentialStep1() + sendgridAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_sendgrid_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_sendgrid_access_credential.test", "name", "test-sendgrid-cred",
					),
				),
			},
		},
	})
}

func sendgridAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestSendGridAPIKey)
	return `
resource "hush_sendgrid_access_credential" "test" {
  name           = "test-sendgrid-cred"
  description    = "test sendgrid credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
}
`
}

func sendgridAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestSendGridAPIKey)
	return `
resource "hush_sendgrid_access_credential" "test" {
  name           = "test-sendgrid-cred-updated"
  description    = "updated sendgrid credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
}
`
}

const sendgridAccessCredentialDataSource = `
data "hush_sendgrid_access_credential" "test" {
  id = hush_sendgrid_access_credential.test.id
}
`
