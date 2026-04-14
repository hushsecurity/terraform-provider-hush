package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSalesforceAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("salesforce_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: salesforceAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_salesforce_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_credential.test", "name", "test-salesforce-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_credential.test", "description", "test salesforce credential",
					),
				),
			},
			{
				Config: salesforceAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_salesforce_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_credential.test", "name", "test-salesforce-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_salesforce_access_credential.test", "description", "updated salesforce credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSalesforceAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("salesforce_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: salesforceAccessCredentialStep1() + salesforceAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_salesforce_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_salesforce_access_credential.test", "name", "test-salesforce-cred",
					),
				),
			},
		},
	})
}

func salesforceAccessCredentialStep1() string {
	return `
resource "hush_salesforce_access_credential" "test" {
  name            = "test-salesforce-cred"
  description     = "test salesforce credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  instance_url    = "https://mock-instance.salesforce.com"
  client_id       = "mock-salesforce-client-id"
  client_secret   = "mock-salesforce-client-secret"
}
`
}

func salesforceAccessCredentialStep2() string {
	return `
resource "hush_salesforce_access_credential" "test" {
  name            = "test-salesforce-cred-updated"
  description     = "updated salesforce credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  instance_url    = "https://mock-instance.salesforce.com"
  client_id       = "mock-salesforce-client-id"
  client_secret   = "mock-salesforce-client-secret"
}
`
}

const salesforceAccessCredentialDataSource = `
data "hush_salesforce_access_credential" "test" {
  id = hush_salesforce_access_credential.test.id
}
`
