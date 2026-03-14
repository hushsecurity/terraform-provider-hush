package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestApigeeServiceAccountKey = "HUSH_TEST_APIGEE_SERVICE_ACCOUNT_KEY"

func testAccApigeeAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestApigeeServiceAccountKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestApigeeServiceAccountKey)
	}
}

func TestAccResourceApigeeAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccApigeeAccessCredentialPreCheck(t) },
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
		PreCheck:          func() { testAccApigeeAccessCredentialPreCheck(t) },
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
	deploymentID := os.Getenv(envHushTestDeploymentID)
	serviceAccountKey := os.Getenv(envHushTestApigeeServiceAccountKey)
	return `
resource "hush_apigee_access_credential" "test" {
  name              = "test-apigee-cred"
  description       = "test apigee credential"
  deployment_ids    = ["` + deploymentID + `"]
  service_account_key = <<-EOF
` + serviceAccountKey + `
EOF
}
`
}

func apigeeAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	serviceAccountKey := os.Getenv(envHushTestApigeeServiceAccountKey)
	return `
resource "hush_apigee_access_credential" "test" {
  name              = "test-apigee-cred-updated"
  description       = "updated apigee credential"
  deployment_ids    = ["` + deploymentID + `"]
  service_account_key = <<-EOF
` + serviceAccountKey + `
EOF
}
`
}

const apigeeAccessCredentialDataSource = `
data "hush_apigee_access_credential" "test" {
  id = hush_apigee_access_credential.test.id
}
`
