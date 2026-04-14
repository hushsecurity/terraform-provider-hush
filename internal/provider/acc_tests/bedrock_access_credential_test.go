package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const mockBedrockAccessKeyID = "AKIAMOCKBEDROCK12345"
const mockBedrockSecretAccessKey = "mock-bedrock-secret"
const mockBedrockRegion = "us-east-1"

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("access_credentials/bedrock", testutil.OpCreate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			if _, ok := obj["access_key_id"]; ok {
				obj["has_provider_credentials"] = true
			}
			return nil
		})
	})
}

func TestAccResourceBedrockAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("bedrock_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: bedrockAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_bedrock_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_bedrock_access_credential.test", "name", "test-bedrock-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_bedrock_access_credential.test", "description", "test bedrock credential",
					),
					resource.TestCheckResourceAttr(
						"hush_bedrock_access_credential.test", "has_provider_credentials", "true",
					),
				),
			},
			{
				Config: bedrockAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_bedrock_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_bedrock_access_credential.test", "name", "test-bedrock-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_bedrock_access_credential.test", "description", "updated bedrock credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceBedrockAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("bedrock_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: bedrockAccessCredentialStep1() + bedrockAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_bedrock_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_bedrock_access_credential.test", "name", "test-bedrock-cred",
					),
				),
			},
		},
	})
}

func bedrockAccessCredentialStep1() string {
	return `
resource "hush_bedrock_access_credential" "test" {
  name           = "test-bedrock-cred"
  description    = "test bedrock credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  region         = "` + mockBedrockRegion + `"
  access_key_id  = "` + mockBedrockAccessKeyID + `"

  secret_access_key = "` + mockBedrockSecretAccessKey + `"
}
`
}

func bedrockAccessCredentialStep2() string {
	return `
resource "hush_bedrock_access_credential" "test" {
  name           = "test-bedrock-cred-updated"
  description    = "updated bedrock credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  region         = "` + mockBedrockRegion + `"
  access_key_id  = "` + mockBedrockAccessKeyID + `"

  secret_access_key = "` + mockBedrockSecretAccessKey + `"
}
`
}

const bedrockAccessCredentialDataSource = `
data "hush_bedrock_access_credential" "test" {
  id = hush_bedrock_access_credential.test.id
}
`
