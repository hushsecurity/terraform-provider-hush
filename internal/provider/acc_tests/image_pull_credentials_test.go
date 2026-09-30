package acc_tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// The credentials are an organization-wide singleton served straight from the
// mock fixtures, so there is nothing to create first and nothing to destroy.
func TestAccDataSourceImagePullCredentials(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: imagePullCredentialsDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hush_image_pull_credentials.test", "id", "image_pull_credentials",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_image_pull_credentials.test", "username",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_image_pull_credentials.test", "password",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_image_pull_credentials.test", "registry",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_image_pull_credentials.test", "token",
					),
				),
			},
		},
	})
}

const imagePullCredentialsDataSource = `
data "hush_image_pull_credentials" "test" {}
`
