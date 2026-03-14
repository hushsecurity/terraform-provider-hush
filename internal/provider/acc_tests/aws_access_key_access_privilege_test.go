package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAWSAccessKeyAccessPrivilege(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_access_key_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: awsAccessKeyAccessPrivilegeStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_aws_access_key_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_privilege.test", "name", "test-aws-priv",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_privilege.test", "description", "test aws privilege",
					),
				),
			},
			{
				Config: awsAccessKeyAccessPrivilegeStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_aws_access_key_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_privilege.test", "name", "test-aws-priv-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_aws_access_key_access_privilege.test", "description", "updated aws privilege",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAWSAccessKeyAccessPrivilege(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("aws_access_key_access_privilege", "v1/access_privileges"),
		Steps: []resource.TestStep{
			{
				Config: awsAccessKeyAccessPrivilegeStep1() + awsAccessKeyAccessPrivilegeDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_aws_access_key_access_privilege.test", "id", regexp.MustCompile(`^apr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_aws_access_key_access_privilege.test", "name", "test-aws-priv",
					),
				),
			},
		},
	})
}

func awsAccessKeyAccessPrivilegeStep1() string {
	return `
resource "hush_aws_access_key_access_privilege" "test" {
  name        = "test-aws-priv"
  description = "test aws privilege"
  policies    = ["arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"]
}
`
}

func awsAccessKeyAccessPrivilegeStep2() string {
	return `
resource "hush_aws_access_key_access_privilege" "test" {
  name        = "test-aws-priv-updated"
  description = "updated aws privilege"
  policies    = ["arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"]
}
`
}

const awsAccessKeyAccessPrivilegeDataSource = `
data "hush_aws_access_key_access_privilege" "test" {
  id = hush_aws_access_key_access_privilege.test.id
}
`
