package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestSalesforceInstanceURL = "HUSH_TEST_SALESFORCE_INSTANCE_URL"
const envHushTestSalesforceClientID = "HUSH_TEST_SALESFORCE_CLIENT_ID"
const envHushTestSalesforceClientSecret = "HUSH_TEST_SALESFORCE_CLIENT_SECRET"

func testAccSalesforceAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestSalesforceInstanceURL) == "" {
		t.Fatalf("%s env var must be set", envHushTestSalesforceInstanceURL)
	}
	if os.Getenv(envHushTestSalesforceClientID) == "" {
		t.Fatalf("%s env var must be set", envHushTestSalesforceClientID)
	}
	if os.Getenv(envHushTestSalesforceClientSecret) == "" {
		t.Fatalf("%s env var must be set", envHushTestSalesforceClientSecret)
	}
}

func TestAccResourceSalesforceAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccSalesforceAccessCredentialPreCheck(t) },
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
		PreCheck:          func() { testAccSalesforceAccessCredentialPreCheck(t) },
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
	deploymentID := os.Getenv(envHushTestDeploymentID)
	instanceURL := os.Getenv(envHushTestSalesforceInstanceURL)
	clientID := os.Getenv(envHushTestSalesforceClientID)
	clientSecret := os.Getenv(envHushTestSalesforceClientSecret)
	return `
resource "hush_salesforce_access_credential" "test" {
  name            = "test-salesforce-cred"
  description     = "test salesforce credential"
  deployment_ids  = ["` + deploymentID + `"]
  instance_url    = "` + instanceURL + `"
  client_id       = "` + clientID + `"
  client_secret   = "` + clientSecret + `"
}
`
}

func salesforceAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	instanceURL := os.Getenv(envHushTestSalesforceInstanceURL)
	clientID := os.Getenv(envHushTestSalesforceClientID)
	clientSecret := os.Getenv(envHushTestSalesforceClientSecret)
	return `
resource "hush_salesforce_access_credential" "test" {
  name            = "test-salesforce-cred-updated"
  description     = "updated salesforce credential"
  deployment_ids  = ["` + deploymentID + `"]
  instance_url    = "` + instanceURL + `"
  client_id       = "` + clientID + `"
  client_secret   = "` + clientSecret + `"
}
`
}

const salesforceAccessCredentialDataSource = `
data "hush_salesforce_access_credential" "test" {
  id = hush_salesforce_access_credential.test.id
}
`
