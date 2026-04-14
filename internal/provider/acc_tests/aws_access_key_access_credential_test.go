package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockAWSAccessKeyID = "AKIAMOCKKEY123456789"
const mockAWSSecretAccessKey = "mock-secret-key-1234567890"

func TestAccResourceAWSAccessKeyAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_access_key_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: awsAccessKeyAccessCredentialStep1,
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
				Config: awsAccessKeyAccessCredentialStep2,
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
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_access_key_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: awsAccessKeyAccessCredentialStep1 + awsAccessKeyAccessCredentialDataSource,
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

const awsAccessKeyAccessCredentialStep1 = `
resource "hush_aws_access_key_access_credential" "test" {
  name                = "test-aws-cred"
  description         = "test aws credential"
  deployment_ids      = ["` + mockDeploymentID + `"]
  access_key_id_value = "` + mockAWSAccessKeyID + `"
  secret_access_key   = "` + mockAWSSecretAccessKey + `"
}
`

const awsAccessKeyAccessCredentialStep2 = `
resource "hush_aws_access_key_access_credential" "test" {
  name                = "test-aws-cred-updated"
  description         = "updated aws credential"
  deployment_ids      = ["` + mockDeploymentID + `"]
  access_key_id_value = "` + mockAWSAccessKeyID + `"
  secret_access_key   = "` + mockAWSSecretAccessKey + `"
}
`

const awsAccessKeyAccessCredentialDataSource = `
data "hush_aws_access_key_access_credential" "test" {
  id = hush_aws_access_key_access_credential.test.id
}
`
