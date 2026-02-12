package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccPostgresAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
}

func TestAccResourcePostgresAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPostgresAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_postgres_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "name", "test-postgres-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "description", "test postgres credential",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "db_name", "testdb",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "host", "test-db.example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "port", "5432",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "ssl_mode", "prefer",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "username", "testuser",
					),
				),
			},
			{
				Config: postgresAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_postgres_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "name", "test-postgres-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_postgres_access_credential.test", "description", "updated postgres credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourcePostgresAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPostgresAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("postgres_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: postgresAccessCredentialStep1() + postgresAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_postgres_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_credential.test", "name", "test-postgres-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_credential.test", "db_name", "testdb",
					),
					resource.TestCheckResourceAttr(
						"data.hush_postgres_access_credential.test", "host", "test-db.example.com",
					),
				),
			},
		},
	})
}

func postgresAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_postgres_access_credential" "test" {
  name           = "test-postgres-cred"
  description    = "test postgres credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-db.example.com"
  port           = 5432
  ssl_mode       = "prefer"
  username       = "testuser"
  password       = "testpassword123"
}
`
}

func postgresAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_postgres_access_credential" "test" {
  name           = "test-postgres-cred-updated"
  description    = "updated postgres credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-db.example.com"
  port           = 5432
  ssl_mode       = "prefer"
  username       = "testuser"
  password       = "testpassword123"
}
`
}

const postgresAccessCredentialDataSource = `
data "hush_postgres_access_credential" "test" {
  id = hush_postgres_access_credential.test.id
}
`
