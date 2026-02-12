package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestGeminiServiceAccountKey = "HUSH_TEST_GEMINI_SERVICE_ACCOUNT_KEY"

func testAccGeminiAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestGeminiServiceAccountKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestGeminiServiceAccountKey)
	}
}

func TestAccResourceGeminiAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccGeminiAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gemini_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: geminiAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gemini_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gemini_access_credential.test", "name", "test-gemini-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_gemini_access_credential.test", "description", "test gemini credential",
					),
				),
			},
			{
				Config: geminiAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gemini_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gemini_access_credential.test", "name", "test-gemini-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gemini_access_credential.test", "description", "updated gemini credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGeminiAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccGeminiAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gemini_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: geminiAccessCredentialStep1() + geminiAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gemini_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gemini_access_credential.test", "name", "test-gemini-cred",
					),
				),
			},
		},
	})
}

func geminiAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	serviceAccountKey := os.Getenv(envHushTestGeminiServiceAccountKey)
	return `
resource "hush_gemini_access_credential" "test" {
  name                = "test-gemini-cred"
  description         = "test gemini credential"
  deployment_ids      = ["` + deploymentID + `"]
  project_id          = "test-gcp-project-1"
  service_account_key = "` + serviceAccountKey + `"
}
`
}

func geminiAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	serviceAccountKey := os.Getenv(envHushTestGeminiServiceAccountKey)
	return `
resource "hush_gemini_access_credential" "test" {
  name                = "test-gemini-cred-updated"
  description         = "updated gemini credential"
  deployment_ids      = ["` + deploymentID + `"]
  project_id          = "test-gcp-project-1"
  service_account_key = "` + serviceAccountKey + `"
}
`
}

const geminiAccessCredentialDataSource = `
data "hush_gemini_access_credential" "test" {
  id = hush_gemini_access_credential.test.id
}
`
