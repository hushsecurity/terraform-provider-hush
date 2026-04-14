package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	mockGrokAPIKey = "xai-mock-grok-key"
	mockGrokTeamID = "mock-team-id"
)

func TestAccResourceGrokAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("grok_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: grokAccessCredentialStep1,
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
				Config: grokAccessCredentialStep2,
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
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("grok_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: grokAccessCredentialStep1 + grokAccessCredentialDataSource,
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

const grokAccessCredentialStep1 = `
resource "hush_grok_access_credential" "test" {
  name           = "test-grok-cred"
  description    = "test grok credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "` + mockGrokAPIKey + `"
  team_id        = "` + mockGrokTeamID + `"
}
`

const grokAccessCredentialStep2 = `
resource "hush_grok_access_credential" "test" {
  name           = "test-grok-cred-updated"
  description    = "updated grok credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "` + mockGrokAPIKey + `"
  team_id        = "` + mockGrokTeamID + `"
}
`

const grokAccessCredentialDataSource = `
data "hush_grok_access_credential" "test" {
  id = hush_grok_access_credential.test.id
}
`
