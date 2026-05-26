package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceTemporalCloudAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("temporal_cloud_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: temporalCloudAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_temporal_cloud_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_credential.test", "name", "test-temporal-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_credential.test", "description", "test temporal cloud credential",
					),
				),
			},
			{
				Config: temporalCloudAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_temporal_cloud_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_credential.test", "name", "test-temporal-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_temporal_cloud_access_credential.test", "description", "updated temporal cloud credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceTemporalCloudAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("temporal_cloud_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: temporalCloudAccessCredentialStep1() + temporalCloudAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_temporal_cloud_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_temporal_cloud_access_credential.test", "name", "test-temporal-cred",
					),
				),
			},
		},
	})
}

func temporalCloudAccessCredentialStep1() string {
	return `
resource "hush_temporal_cloud_access_credential" "test" {
  name           = "test-temporal-cred"
  description    = "test temporal cloud credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "test-temporal-api-key"
}
`
}

func temporalCloudAccessCredentialStep2() string {
	return `
resource "hush_temporal_cloud_access_credential" "test" {
  name           = "test-temporal-cred-updated"
  description    = "updated temporal cloud credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  api_key        = "test-temporal-api-key"
}
`
}

const temporalCloudAccessCredentialDataSource = `
data "hush_temporal_cloud_access_credential" "test" {
  id = hush_temporal_cloud_access_credential.test.id
}
`
