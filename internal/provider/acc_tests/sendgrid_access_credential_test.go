package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSendGridAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
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
	return `
resource "hush_sendgrid_access_credential" "test" {
  name           = "test-sendgrid-cred"
  description    = "test sendgrid credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "mock-sendgrid-api-key"
}
`
}

func sendgridAccessCredentialStep2() string {
	return `
resource "hush_sendgrid_access_credential" "test" {
  name           = "test-sendgrid-cred-updated"
  description    = "updated sendgrid credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "mock-sendgrid-api-key"
}
`
}

const sendgridAccessCredentialDataSource = `
data "hush_sendgrid_access_credential" "test" {
  id = hush_sendgrid_access_credential.test.id
}
`
