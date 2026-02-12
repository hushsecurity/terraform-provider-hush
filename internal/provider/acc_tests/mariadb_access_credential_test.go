package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccMariaDBAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
}

func TestAccResourceMariaDBAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccMariaDBAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mariadb_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mariadbAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mariadb_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "name", "test-mariadb-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "description", "test mariadb credential",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "db_name", "testdb",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "host", "test-mariadb.example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "port", "3306",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "ssl_mode", "preferred",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "username", "testuser",
					),
				),
			},
			{
				Config: mariadbAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mariadb_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "name", "test-mariadb-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mariadb_access_credential.test", "description", "updated mariadb credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMariaDBAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccMariaDBAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mariadb_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mariadbAccessCredentialStep1() + mariadbAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mariadb_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mariadb_access_credential.test", "name", "test-mariadb-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mariadb_access_credential.test", "db_name", "testdb",
					),
				),
			},
		},
	})
}

func mariadbAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_mariadb_access_credential" "test" {
  name           = "test-mariadb-cred"
  description    = "test mariadb credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-mariadb.example.com"
  port           = 3306
  ssl_mode       = "preferred"
  username       = "testuser"
  password       = "testpassword123"
}
`
}

func mariadbAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_mariadb_access_credential" "test" {
  name           = "test-mariadb-cred-updated"
  description    = "updated mariadb credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-mariadb.example.com"
  port           = 3306
  ssl_mode       = "preferred"
  username       = "testuser"
  password       = "testpassword123"
}
`
}

const mariadbAccessCredentialDataSource = `
data "hush_mariadb_access_credential" "test" {
  id = hush_mariadb_access_credential.test.id
}
`
