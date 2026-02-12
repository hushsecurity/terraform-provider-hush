package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccMySQLAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
}

func TestAccResourceMySQLAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccMySQLAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mysql_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mysqlAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mysql_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "name", "test-mysql-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "description", "test mysql credential",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "db_name", "testdb",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "host", "test-mysql.example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "port", "3306",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "ssl_mode", "preferred",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "username", "testuser",
					),
				),
			},
			{
				Config: mysqlAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mysql_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "name", "test-mysql-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mysql_access_credential.test", "description", "updated mysql credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMySQLAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccMySQLAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mysql_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mysqlAccessCredentialStep1() + mysqlAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mysql_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mysql_access_credential.test", "name", "test-mysql-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mysql_access_credential.test", "db_name", "testdb",
					),
				),
			},
		},
	})
}

func mysqlAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_mysql_access_credential" "test" {
  name           = "test-mysql-cred"
  description    = "test mysql credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-mysql.example.com"
  port           = 3306
  ssl_mode       = "preferred"
  username       = "testuser"
  password       = "testpassword123"
}
`
}

func mysqlAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_mysql_access_credential" "test" {
  name           = "test-mysql-cred-updated"
  description    = "updated mysql credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-mysql.example.com"
  port           = 3306
  ssl_mode       = "preferred"
  username       = "testuser"
  password       = "testpassword123"
}
`
}

const mysqlAccessCredentialDataSource = `
data "hush_mysql_access_credential" "test" {
  id = hush_mysql_access_credential.test.id
}
`
