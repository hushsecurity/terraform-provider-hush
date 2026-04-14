package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockGCPSAKey = `{"type":"service_account","project_id":"mock"}`

func TestAccResourceGCPSAAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
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
	return `
resource "hush_gcp_sa_access_credential" "test" {
  name                = "test-gcp-sa-cred"
  description         = "test gcp service account credential"
  deployment_ids      = ["` + mockDeploymentID + `"]
  service_account_key = <<-EOF
` + mockGCPSAKey + `
EOF
}
`
}

func gcpSAAccessCredentialStep2() string {
	return `
resource "hush_gcp_sa_access_credential" "test" {
  name                = "test-gcp-sa-cred-updated"
  description         = "updated gcp service account credential"
  deployment_ids      = ["` + mockDeploymentID + `"]
  service_account_key = <<-EOF
` + mockGCPSAKey + `
EOF
}
`
}

const gcpSAAccessCredentialDataSource = `
data "hush_gcp_sa_access_credential" "test" {
  id = hush_gcp_sa_access_credential.test.id
}
`
