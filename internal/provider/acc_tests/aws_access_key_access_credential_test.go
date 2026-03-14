package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestAWSAccessKeyID = "HUSH_TEST_AWS_ACCESS_KEY_ID"
const envHushTestAWSSecretAccessKey = "HUSH_TEST_AWS_SECRET_ACCESS_KEY"

func testAccAWSAccessKeyAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestAWSAccessKeyID) == "" {
		t.Fatalf("%s env var must be set", envHushTestAWSAccessKeyID)
	}
	if os.Getenv(envHushTestAWSSecretAccessKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestAWSSecretAccessKey)
	}
}

func TestAccResourceAWSAccessKeyAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAWSAccessKeyAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_access_key_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: awsAccessKeyAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_aws_access_key_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_credential.test", "name", "test-aws-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_credential.test", "description", "test aws credential",
					),
				),
			},
			{
				Config: awsAccessKeyAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_aws_access_key_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_credential.test", "name", "test-aws-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_credential.test", "description", "updated aws credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAWSAccessKeyAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAWSAccessKeyAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_access_key_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: awsAccessKeyAccessCredentialStep1() + awsAccessKeyAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_aws_access_key_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_aws_access_key_access_credential.test", "name", "test-aws-cred",
					),
				),
			},
		},
	})
}

func awsAccessKeyAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	accessKeyID := os.Getenv(envHushTestAWSAccessKeyID)
	secretAccessKey := os.Getenv(envHushTestAWSSecretAccessKey)
	return `
resource "hush_aws_access_key_access_credential" "test" {
  name                = "test-aws-cred"
  description         = "test aws credential"
  deployment_ids      = ["` + deploymentID + `"]
  access_key_id_value = "` + accessKeyID + `"
  secret_access_key   = "` + secretAccessKey + `"
}
`
}

func awsAccessKeyAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	accessKeyID := os.Getenv(envHushTestAWSAccessKeyID)
	secretAccessKey := os.Getenv(envHushTestAWSSecretAccessKey)
	return `
resource "hush_aws_access_key_access_credential" "test" {
  name                = "test-aws-cred-updated"
  description         = "updated aws credential"
  deployment_ids      = ["` + deploymentID + `"]
  access_key_id_value = "` + accessKeyID + `"
  secret_access_key   = "` + secretAccessKey + `"
}
`
}

const awsAccessKeyAccessCredentialDataSource = `
data "hush_aws_access_key_access_credential" "test" {
  id = hush_aws_access_key_access_credential.test.id
}
`
