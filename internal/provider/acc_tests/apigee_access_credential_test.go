package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const mockApigeeServiceAccountKey = `{"type":"service_account","project_id":"mock"}`

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("access_credentials/apigee", testutil.OpCreate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			if _, ok := obj["service_account_key"]; ok {
				obj["has_provider_credentials"] = true
			}
			return nil
		})
	})
}

func TestAccResourceApigeeAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("apigee_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: apigeeAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_apigee_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_credential.test", "name", "test-apigee-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_credential.test", "description", "test apigee credential",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_credential.test", "has_provider_credentials", "true",
					),
				),
			},
			{
				Config: apigeeAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_apigee_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_credential.test", "name", "test-apigee-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_apigee_access_credential.test", "description", "updated apigee credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceApigeeAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("apigee_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: apigeeAccessCredentialStep1() + apigeeAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_apigee_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_apigee_access_credential.test", "name", "test-apigee-cred",
					),
				),
			},
		},
	})
}

func apigeeAccessCredentialStep1() string {
	return `
resource "hush_apigee_access_credential" "test" {
  name              = "test-apigee-cred"
  description       = "test apigee credential"
  deployment_ids    = ["` + mockDeploymentID + `"]
  service_account_key = <<-EOF
` + mockApigeeServiceAccountKey + `
EOF
}
`
}

func apigeeAccessCredentialStep2() string {
	return `
resource "hush_apigee_access_credential" "test" {
  name              = "test-apigee-cred-updated"
  description       = "updated apigee credential"
  deployment_ids    = ["` + mockDeploymentID + `"]
  service_account_key = <<-EOF
` + mockApigeeServiceAccountKey + `
EOF
}
`
}

const apigeeAccessCredentialDataSource = `
data "hush_apigee_access_credential" "test" {
  id = hush_apigee_access_credential.test.id
}
`
