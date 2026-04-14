package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockAzureTenantID = "b2c3d4e5-f6a7-8901-bcde-f12345678901"
const mockAzureClientID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
const mockAzureClientSecret = "mock-azure-client-secret"

func TestAccResourceAzureAppAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: azureAppAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_azure_app_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_credential.test", "name", "test-azure-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_credential.test", "description", "test azure credential",
					),
				),
			},
			{
				Config: azureAppAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_azure_app_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_credential.test", "name", "test-azure-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_app_access_credential.test", "description", "updated azure credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAzureAppAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: azureAppAccessCredentialStep1 + azureAppAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_azure_app_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_azure_app_access_credential.test", "name", "test-azure-cred",
					),
				),
			},
		},
	})
}

const azureAppAccessCredentialStep1 = `
resource "hush_azure_app_access_credential" "test" {
  name            = "test-azure-cred"
  description     = "test azure credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  tenant_id       = "` + mockAzureTenantID + `"
  client_id       = "` + mockAzureClientID + `"
  client_secret   = "` + mockAzureClientSecret + `"
}
`

const azureAppAccessCredentialStep2 = `
resource "hush_azure_app_access_credential" "test" {
  name            = "test-azure-cred-updated"
  description     = "updated azure credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  tenant_id       = "` + mockAzureTenantID + `"
  client_id       = "` + mockAzureClientID + `"
  client_secret   = "` + mockAzureClientSecret + `"
}
`

const azureAppAccessCredentialDataSource = `
data "hush_azure_app_access_credential" "test" {
  id = hush_azure_app_access_credential.test.id
}
`
