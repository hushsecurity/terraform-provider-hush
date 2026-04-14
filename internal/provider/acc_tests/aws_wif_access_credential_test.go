package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("access_credentials/aws_wif", testutil.OpCreate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["audience"] = "sts.amazonaws.com"
			obj["issuer_url"] = "https://hush-oidc.example.com/" + fmt.Sprintf("%v", obj["id"])
			return nil
		})
	})
}

func TestAccResourceAwsWifAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
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
						"hush_aws_wif_access_credential.test", "deployment_ids.0", mockDeploymentID,
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
						"hush_aws_wif_access_credential.test", "deployment_ids.0", mockDeploymentID,
					),
					resource.TestCheckResourceAttr(
						"hush_aws_wif_access_credential.test", "deployment_ids.1", mockDeploymentID2,
					),
				),
			},
		},
	})
}

func TestAccDataSourceAwsWifAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
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
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred"
  description    = "test aws wif credential"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`
}

func awsWifAccessCredentialStep2() string {
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred-updated"
  description    = "updated aws wif credential"
  deployment_ids = ["` + mockDeploymentID + `", "` + mockDeploymentID2 + `"]
}
`
}

const awsWifAccessCredentialDataSource = `
data "hush_aws_wif_access_credential" "test" {
  id = hush_aws_wif_access_credential.test.id
}
`
