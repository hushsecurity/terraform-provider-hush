package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestGrokAPIKey = "HUSH_TEST_GROK_API_KEY"
const envHushTestGrokTeamID = "HUSH_TEST_GROK_TEAM_ID"

func testAccGrokAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestGrokAPIKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestGrokAPIKey)
	}
	if os.Getenv(envHushTestGrokTeamID) == "" {
		t.Fatalf("%s env var must be set", envHushTestGrokTeamID)
	}
}

func TestAccResourceGrokAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccGrokAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("grok_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: grokAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_grok_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_credential.test", "name", "test-grok-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_credential.test", "description", "test grok credential",
					),
				),
			},
			{
				Config: grokAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_grok_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_credential.test", "name", "test-grok-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_grok_access_credential.test", "description", "updated grok credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGrokAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccGrokAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("grok_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: grokAccessCredentialStep1() + grokAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_grok_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_grok_access_credential.test", "name", "test-grok-cred",
					),
				),
			},
		},
	})
}

func grokAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestGrokAPIKey)
	teamID := os.Getenv(envHushTestGrokTeamID)
	return `
resource "hush_grok_access_credential" "test" {
  name           = "test-grok-cred"
  description    = "test grok credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
  team_id        = "` + teamID + `"
}
`
}

func grokAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	apiKey := os.Getenv(envHushTestGrokAPIKey)
	teamID := os.Getenv(envHushTestGrokTeamID)
	return `
resource "hush_grok_access_credential" "test" {
  name           = "test-grok-cred-updated"
  description    = "updated grok credential"
  deployment_ids = ["` + deploymentID + `"]
  api_key        = "` + apiKey + `"
  team_id        = "` + teamID + `"
}
`
}

const grokAccessCredentialDataSource = `
data "hush_grok_access_credential" "test" {
  id = hush_grok_access_credential.test.id
}
`
