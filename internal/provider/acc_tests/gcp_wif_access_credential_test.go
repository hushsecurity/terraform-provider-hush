package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccGcpWifAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestDeploymentID2) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID2)
	}
}

func TestAccResourceGcpWifAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccGcpWifAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gcp_wif_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gcpWifAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gcp_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "name", "test-gcp-wif-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "description", "test gcp wif credential",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "deployment_ids.#", "1",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "deployment_ids.0", os.Getenv(envHushTestDeploymentID),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "project_number", "123456789012",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "pool_id", "my-wif-pool",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "workload_provider_id", "my-wif-provider",
					),
					resource.TestMatchResourceAttr(
						"hush_gcp_wif_access_credential.test", "issuer_url", regexp.MustCompile(`^https://.+$`),
					),
				),
			},
			{
				Config: gcpWifAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_gcp_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "name", "test-gcp-wif-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "description", "updated gcp wif credential",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "deployment_ids.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "deployment_ids.0", os.Getenv(envHushTestDeploymentID),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "deployment_ids.1", os.Getenv(envHushTestDeploymentID2),
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "project_number", "987654321098",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "pool_id", "my-wif-pool-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_gcp_wif_access_credential.test", "workload_provider_id", "my-wif-provider-updated",
					),
				),
			},
		},
	})
}

func TestAccDataSourceGcpWifAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccGcpWifAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("gcp_wif_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: gcpWifAccessCredentialStep1() + gcpWifAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_gcp_wif_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_wif_access_credential.test", "name", "test-gcp-wif-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_wif_access_credential.test", "project_number", "123456789012",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_wif_access_credential.test", "pool_id", "my-wif-pool",
					),
					resource.TestCheckResourceAttr(
						"data.hush_gcp_wif_access_credential.test", "workload_provider_id", "my-wif-provider",
					),
				),
			},
		},
	})
}

func gcpWifAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_gcp_wif_access_credential" "test" {
  name                 = "test-gcp-wif-cred"
  description          = "test gcp wif credential"
  deployment_ids       = ["` + deploymentID + `"]
  project_number       = "123456789012"
  pool_id              = "my-wif-pool"
  workload_provider_id = "my-wif-provider"
}
`
}

func gcpWifAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	deploymentID2 := os.Getenv(envHushTestDeploymentID2)
	return `
resource "hush_gcp_wif_access_credential" "test" {
  name                 = "test-gcp-wif-cred-updated"
  description          = "updated gcp wif credential"
  deployment_ids       = ["` + deploymentID + `", "` + deploymentID2 + `"]
  project_number       = "987654321098"
  pool_id              = "my-wif-pool-updated"
  workload_provider_id = "my-wif-provider-updated"
}
`
}

const gcpWifAccessCredentialDataSource = `
data "hush_gcp_wif_access_credential" "test" {
  id = hush_gcp_wif_access_credential.test.id
}
`
