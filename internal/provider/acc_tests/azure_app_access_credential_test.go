package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestAzureTenantID = "HUSH_TEST_AZURE_TENANT_ID"
const envHushTestAzureClientID = "HUSH_TEST_AZURE_CLIENT_ID"
const envHushTestAzureClientSecret = "HUSH_TEST_AZURE_CLIENT_SECRET"

func testAccAzureAppAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestAzureTenantID) == "" {
		t.Fatalf("%s env var must be set", envHushTestAzureTenantID)
	}
	if os.Getenv(envHushTestAzureClientID) == "" {
		t.Fatalf("%s env var must be set", envHushTestAzureClientID)
	}
	if os.Getenv(envHushTestAzureClientSecret) == "" {
		t.Fatalf("%s env var must be set", envHushTestAzureClientSecret)
	}
}

func TestAccResourceAzureAppAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAzureAppAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: azureAppAccessCredentialStep1(),
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
				Config: azureAppAccessCredentialStep2(),
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
		PreCheck:          func() { testAccAzureAppAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: azureAppAccessCredentialStep1() + azureAppAccessCredentialDataSource,
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

func azureAppAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	tenantID := os.Getenv(envHushTestAzureTenantID)
	clientID := os.Getenv(envHushTestAzureClientID)
	clientSecret := os.Getenv(envHushTestAzureClientSecret)
	return `
resource "hush_azure_app_access_credential" "test" {
  name            = "test-azure-cred"
  description     = "test azure credential"
  deployment_ids  = ["` + deploymentID + `"]
  tenant_id       = "` + tenantID + `"
  client_id       = "` + clientID + `"
  client_secret   = "` + clientSecret + `"
}
`
}

func azureAppAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	tenantID := os.Getenv(envHushTestAzureTenantID)
	clientID := os.Getenv(envHushTestAzureClientID)
	clientSecret := os.Getenv(envHushTestAzureClientSecret)
	return `
resource "hush_azure_app_access_credential" "test" {
  name            = "test-azure-cred-updated"
  description     = "updated azure credential"
  deployment_ids  = ["` + deploymentID + `"]
  tenant_id       = "` + tenantID + `"
  client_id       = "` + clientID + `"
  client_secret   = "` + clientSecret + `"
}
`
}

const azureAppAccessCredentialDataSource = `
data "hush_azure_app_access_credential" "test" {
  id = hush_azure_app_access_credential.test.id
}
`
