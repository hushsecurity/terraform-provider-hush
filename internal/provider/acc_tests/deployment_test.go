package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDeployment(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "name", "test-deployment",
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "description", "test deployment description",
					),
				),
			},
			{
				Config: deploymentStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "name", "test-deployment-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "description", "updated deployment description",
					),
				),
			},
		},
	})
}

func TestAccDataSourceDeployment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentStep1 + deploymentDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_deployment.test", "name", "test-deployment",
					),
					resource.TestCheckResourceAttr(
						"data.hush_deployment.test", "description", "test deployment description",
					),
				),
			},
		},
	})
}

const (
	deploymentStep1 = `
resource "hush_deployment" "test" {
  name        = "test-deployment"
  description = "test deployment description"
}
`

	deploymentStep2 = `
resource "hush_deployment" "test" {
  name        = "test-deployment-updated"
  description = "updated deployment description"
}
`

	deploymentDataSource = `
data "hush_deployment" "test" {
  id = hush_deployment.test.id
}
`
)
