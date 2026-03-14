package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestBedrockAccessKeyID = "HUSH_TEST_BEDROCK_ACCESS_KEY_ID"
const envHushTestBedrockSecretAccessKey = "HUSH_TEST_BEDROCK_SECRET_ACCESS_KEY"
const envHushTestBedrockRegion = "HUSH_TEST_BEDROCK_REGION"

func testAccBedrockAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestBedrockAccessKeyID) == "" {
		t.Fatalf("%s env var must be set", envHushTestBedrockAccessKeyID)
	}
	if os.Getenv(envHushTestBedrockSecretAccessKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestBedrockSecretAccessKey)
	}
	if os.Getenv(envHushTestBedrockRegion) == "" {
		t.Fatalf("%s env var must be set", envHushTestBedrockRegion)
	}
}

func TestAccResourceBedrockAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccBedrockAccessCredentialPreCheck(t) },
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
		PreCheck:          func() { testAccBedrockAccessCredentialPreCheck(t) },
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
	deploymentID := os.Getenv(envHushTestDeploymentID)
	accessKeyID := os.Getenv(envHushTestBedrockAccessKeyID)
	secretAccessKey := os.Getenv(envHushTestBedrockSecretAccessKey)
	region := os.Getenv(envHushTestBedrockRegion)
	return `
resource "hush_bedrock_access_credential" "test" {
  name           = "test-bedrock-cred"
  description    = "test bedrock credential"
  deployment_ids = ["` + deploymentID + `"]
  region         = "` + region + `"
  access_key_id  = "` + accessKeyID + `"

  secret_access_key = "` + secretAccessKey + `"
}
`
}

func bedrockAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	accessKeyID := os.Getenv(envHushTestBedrockAccessKeyID)
	secretAccessKey := os.Getenv(envHushTestBedrockSecretAccessKey)
	region := os.Getenv(envHushTestBedrockRegion)
	return `
resource "hush_bedrock_access_credential" "test" {
  name           = "test-bedrock-cred-updated"
  description    = "updated bedrock credential"
  deployment_ids = ["` + deploymentID + `"]
  region         = "` + region + `"
  access_key_id  = "` + accessKeyID + `"

  secret_access_key = "` + secretAccessKey + `"
}
`
}

const bedrockAccessCredentialDataSource = `
data "hush_bedrock_access_credential" "test" {
  id = hush_bedrock_access_credential.test.id
}
`
