package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestGCPSAKey = "HUSH_TEST_GCP_SA_KEY"

func testAccGCPSAAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestGCPSAKey) == "" {
		t.Fatalf("%s env var must be set", envHushTestGCPSAKey)
	}
}

func TestAccResourceGCPSAAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccGCPSAAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gcp_sa_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gcpSAAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gcp_sa_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_credential.test", "name", "test-gcp-sa-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_credential.test", "description", "test gcp service account credential",
					),
				),
			},
			{
				Config: gcpSAAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gcp_sa_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_credential.test", "name", "test-gcp-sa-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_sa_access_credential.test", "description", "updated gcp service account credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGCPSAAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccGCPSAAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gcp_sa_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gcpSAAccessCredentialStep1() + gcpSAAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gcp_sa_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_sa_access_credential.test", "name", "test-gcp-sa-cred",
					),
				),
			},
		},
	})
}

func gcpSAAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	serviceAccountKey := os.Getenv(envHushTestGCPSAKey)
	return `
resource "hush_gcp_sa_access_credential" "test" {
  name                = "test-gcp-sa-cred"
  description         = "test gcp service account credential"
  deployment_ids      = ["` + deploymentID + `"]
  service_account_key = "` + serviceAccountKey + `"
}
`
}

func gcpSAAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	serviceAccountKey := os.Getenv(envHushTestGCPSAKey)
	return `
resource "hush_gcp_sa_access_credential" "test" {
  name                = "test-gcp-sa-cred-updated"
  description         = "updated gcp service account credential"
  deployment_ids      = ["` + deploymentID + `"]
  service_account_key = "` + serviceAccountKey + `"
}
`
}

const gcpSAAccessCredentialDataSource = `
data "hush_gcp_sa_access_credential" "test" {
  id = hush_gcp_sa_access_credential.test.id
}
`
