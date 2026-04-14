package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	mockOpenAIAPIKey    = "sk-mock-openai-key-1234567890"
	mockOpenAIProjectID = "proj_mock_openai"
)

func TestAccResourceOpenAIAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessCredentialStep1,
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
				Config: openaiAccessCredentialStep2,
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
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("openai_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: openaiAccessCredentialStep1 + openaiAccessCredentialDataSource,
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

const openaiAccessCredentialStep1 = `
resource "hush_openai_access_credential" "test" {
  name           = "test-openai-cred"
  description    = "test openai credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "` + mockOpenAIAPIKey + `"
  project_id     = "` + mockOpenAIProjectID + `"
}
`

const openaiAccessCredentialStep2 = `
resource "hush_openai_access_credential" "test" {
  name           = "test-openai-cred-updated"
  description    = "updated openai credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "` + mockOpenAIAPIKey + `"
  project_id     = "` + mockOpenAIProjectID + `"
}
`

const openaiAccessCredentialDataSource = `
data "hush_openai_access_credential" "test" {
  id = hush_openai_access_credential.test.id
}
`
