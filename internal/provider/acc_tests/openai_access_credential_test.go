package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestOpenAIAPIKey = "HUSH_TEST_OPENAI_API_KEY"
const envHushTestOpenAIProjectID = "HUSH_TEST_OPENAI_PROJECT_ID"

func testAccOpenAIAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestOpenAIAPIKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestOpenAIAPIKey)
	}
	if os.Getenv(envHushTestOpenAIProjectID) == "" {
		t.Fatalf("%s env var must be set", envHushTestOpenAIProjectID)
	}
}

func TestAccResourceOpenAIAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccOpenAIAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_openai_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_credential.test", "name", "test-openai-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_credential.test", "description", "test openai credential",
					),
				),
			},
			{
				Config: openaiAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_openai_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_credential.test", "name", "test-openai-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_openai_access_credential.test", "description", "updated openai credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceOpenAIAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccOpenAIAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessCredentialStep1() + openaiAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_openai_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_openai_access_credential.test", "name", "test-openai-cred",
					),
				),
			},
		},
	})
}

func openaiAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestOpenAIAPIKey)
	projectID := os.Getenv(envHushTestOpenAIProjectID)
	return `
resource "hush_openai_access_credential" "test" {
  name           = "test-openai-cred"
  description    = "test openai credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
  project_id     = "` + projectID + `"
}
`
}

func openaiAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestOpenAIAPIKey)
	projectID := os.Getenv(envHushTestOpenAIProjectID)
	return `
resource "hush_openai_access_credential" "test" {
  name           = "test-openai-cred-updated"
  description    = "updated openai credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
  project_id     = "` + projectID + `"
}
`
}

const openaiAccessCredentialDataSource = `
data "hush_openai_access_credential" "test" {
  id = hush_openai_access_credential.test.id
}
`
