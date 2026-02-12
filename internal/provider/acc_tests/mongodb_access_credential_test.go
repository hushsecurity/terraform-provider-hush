package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccMongoDBAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
}

func TestAccResourceMongoDBAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccMongoDBAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "name", "test-mongodb-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "description", "test mongodb credential",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "db_name", "testdb",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "host", "test-mongo.example.com",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "port", "27017",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "username", "testuser",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "auth_source", "admin",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "tls", "false",
					),
				),
			},
			{
				Config: mongodbAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "name", "test-mongodb-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_access_credential.test", "description", "updated mongodb credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMongoDBAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccMongoDBAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAccessCredentialStep1() + mongodbAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mongodb_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_access_credential.test", "name", "test-mongodb-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_access_credential.test", "db_name", "testdb",
					),
				),
			},
		},
	})
}

func mongodbAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_mongodb_access_credential" "test" {
  name           = "test-mongodb-cred"
  description    = "test mongodb credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-mongo.example.com"
  port           = 27017
  username       = "testuser"
  password       = "testpassword123"
  auth_source    = "admin"
  tls            = false
}
`
}

func mongodbAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_mongodb_access_credential" "test" {
  name           = "test-mongodb-cred-updated"
  description    = "updated mongodb credential"
  deployment_ids = ["` + deploymentID + `"]
  db_name        = "testdb"
  host           = "test-mongo.example.com"
  port           = 27017
  username       = "testuser"
  password       = "testpassword123"
  auth_source    = "admin"
  tls            = false
}
`
}

const mongodbAccessCredentialDataSource = `
data "hush_mongodb_access_credential" "test" {
  id = hush_mongodb_access_credential.test.id
}
`
