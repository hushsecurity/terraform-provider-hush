package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("access_credentials/azure_wif", testutil.OpCreate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["audience"] = "api://AzureADTokenExchange"
			obj["issuer_url"] = "https://hush-oidc.example.com/" + fmt.Sprintf("%v", obj["id"])
			return nil
		})
	})
}

func TestAccResourceAzureWifAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_wif_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: azureWifAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_azure_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "name", "test-azure-wif-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "description", "test azure wif credential",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "deployment_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "deployment_ids.0", mockDeploymentID,
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "audience", "api://AzureADTokenExchange",
					),
					resource.TestMatchResourceAttr(
						"hush_azure_wif_access_credential.test", "issuer_url", regexp.MustCompile(`^https://.+$`),
					),
				),
			},
			{
				Config: azureWifAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_azure_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "name", "test-azure-wif-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "description", "updated azure wif credential",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "deployment_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_azure_wif_access_credential.test", "deployment_ids.0", mockDeploymentID2,
					),
				),
			},
		},
	})
}

func TestAccDataSourceAzureWifAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("azure_wif_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: azureWifAccessCredentialStep1() + azureWifAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_azure_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_azure_wif_access_credential.test", "name", "test-azure-wif-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_azure_wif_access_credential.test", "audience", "api://AzureADTokenExchange",
					),
				),
			},
		},
	})
}

func azureWifAccessCredentialStep1() string {
	return `
resource "hush_azure_wif_access_credential" "test" {
  name           = "test-azure-wif-cred"
  description    = "test azure wif credential"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`
}

func azureWifAccessCredentialStep2() string {
	return `
resource "hush_azure_wif_access_credential" "test" {
  name           = "test-azure-wif-cred-updated"
  description    = "updated azure wif credential"
  deployment_ids = ["` + mockDeploymentID2 + `"]
}
`
}

const azureWifAccessCredentialDataSource = `
data "hush_azure_wif_access_credential" "test" {
  id = hush_azure_wif_access_credential.test.id
}
`
