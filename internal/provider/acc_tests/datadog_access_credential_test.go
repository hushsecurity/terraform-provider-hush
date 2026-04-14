package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestDatadogAPIKey = "HUSH_TEST_DATADOG_API_KEY"
const envHushTestDatadogAppKey = "HUSH_TEST_DATADOG_APP_KEY"

func testAccDatadogAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestDatadogAPIKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestDatadogAPIKey)
	}
	if os.Getenv(envHushTestDatadogAppKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestDatadogAppKey)
	}
}

func TestAccResourceDatadogAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccDatadogAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("datadog_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: datadogAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_datadog_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_credential.test", "name", "test-datadog-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_credential.test", "description", "test datadog credential",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_credential.test", "site", "us5.datadoghq.com",
					),
				),
			},
			{
				Config: datadogAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_datadog_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_credential.test", "name", "test-datadog-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_credential.test", "description", "updated datadog credential",
					),
					resource.TestCheckResourceAttr(
						"hush_datadog_access_credential.test", "site", "datadoghq.com",
					),
				),
			},
		},
	})
}

func TestAccDataSourceDatadogAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccDatadogAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("datadog_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: datadogAccessCredentialStep1() + datadogAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_datadog_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_datadog_access_credential.test", "name", "test-datadog-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_datadog_access_credential.test", "site", "us5.datadoghq.com",
					),
				),
			},
		},
	})
}

func datadogAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestDatadogAPIKey)
	appKey := os.Getenv(envHushTestDatadogAppKey)
	return `
resource "hush_datadog_access_credential" "test" {
  name           = "test-datadog-cred"
  description    = "test datadog credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
  app_key        = "` + appKey + `"
  site           = "us5.datadoghq.com"
}
`
}

func datadogAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestDatadogAPIKey)
	appKey := os.Getenv(envHushTestDatadogAppKey)
	return `
resource "hush_datadog_access_credential" "test" {
  name           = "test-datadog-cred-updated"
  description    = "updated datadog credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
  app_key        = "` + appKey + `"
  site           = "datadoghq.com"
}
`
}

const datadogAccessCredentialDataSource = `
data "hush_datadog_access_credential" "test" {
  id = hush_datadog_access_credential.test.id
}
`
