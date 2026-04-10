package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccAwsWifAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestDeploymentID2) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID2)
	}
}

func TestAccResourceAwsWifAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAwsWifAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_wif_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: awsWifAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_aws_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "name", "test-aws-wif-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "description", "test aws wif credential",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "deployment_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "deployment_ids.0", os.Getenv(envHushTestDeploymentID),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "audience", "sts.amazonaws.com",
					),
					resource.TestMatchResourceAttr(
						"hush_aws_wif_access_credential.test", "issuer_url", regexp.MustCompile(`^https://.+$`),
					),
				),
			},
			{
				Config: awsWifAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_aws_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "name", "test-aws-wif-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "description", "updated aws wif credential",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "deployment_ids.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "deployment_ids.0", os.Getenv(envHushTestDeploymentID),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "deployment_ids.1", os.Getenv(envHushTestDeploymentID2),
					),
				),
			},
		},
	})
}

func TestAccDataSourceAwsWifAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAwsWifAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_wif_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: awsWifAccessCredentialStep1() + awsWifAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_aws_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_aws_wif_access_credential.test", "name", "test-aws-wif-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_aws_wif_access_credential.test", "audience", "sts.amazonaws.com",
					),
				),
			},
		},
	})
}

func awsWifAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred"
  description    = "test aws wif credential"
  deployment_ids = ["` + deploymentID + `"]
}
`
}

func awsWifAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	deploymentID2 := os.Getenv(envHushTestDeploymentID2)
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred-updated"
  description    = "updated aws wif credential"
  deployment_ids = ["` + deploymentID + `", "` + deploymentID2 + `"]
}
`
}

const awsWifAccessCredentialDataSource = `
data "hush_aws_wif_access_credential" "test" {
  id = hush_aws_wif_access_credential.test.id
}
`
