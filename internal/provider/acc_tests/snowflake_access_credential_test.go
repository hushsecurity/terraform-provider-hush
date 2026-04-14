package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSnowflakeAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("snowflake_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: snowflakeAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_snowflake_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "name", "test-snowflake-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "description", "test snowflake credential",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "account", "TESTORG-TESTACCOUNT",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "warehouse", "COMPUTE_WH",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "database", "TESTDB",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "schema", "PUBLIC",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "username", "testuser",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "auth_method", "password",
					),
				),
			},
			{
				Config: snowflakeAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_snowflake_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "name", "test-snowflake-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_snowflake_access_credential.test", "description", "updated snowflake credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSnowflakeAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("snowflake_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: snowflakeAccessCredentialStep1() + snowflakeAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_snowflake_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_snowflake_access_credential.test", "name", "test-snowflake-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_snowflake_access_credential.test", "account", "TESTORG-TESTACCOUNT",
					),
					resource.TestCheckResourceAttr(
						"data.hush_snowflake_access_credential.test", "database", "TESTDB",
					),
				),
			},
		},
	})
}

func snowflakeAccessCredentialStep1() string {
	return `
resource "hush_snowflake_access_credential" "test" {
  name           = "test-snowflake-cred"
  description    = "test snowflake credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  account        = "TESTORG-TESTACCOUNT"
  warehouse      = "COMPUTE_WH"
  database       = "TESTDB"
  schema         = "PUBLIC"
  username       = "testuser"
  password       = "TestPassword123!"
  auth_method    = "password"
}
`
}

func snowflakeAccessCredentialStep2() string {
	return `
resource "hush_snowflake_access_credential" "test" {
  name           = "test-snowflake-cred-updated"
  description    = "updated snowflake credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  account        = "TESTORG-TESTACCOUNT"
  warehouse      = "COMPUTE_WH"
  database       = "TESTDB"
  schema         = "PUBLIC"
  username       = "testuser"
  password       = "TestPassword123!"
  auth_method    = "password"
}
`
}

const snowflakeAccessCredentialDataSource = `
data "hush_snowflake_access_credential" "test" {
  id = hush_snowflake_access_credential.test.id
}
`
