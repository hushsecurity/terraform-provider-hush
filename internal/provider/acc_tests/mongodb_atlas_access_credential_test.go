package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMongoDBAtlasAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "name", "test-atlas-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "description", "test atlas credential",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "group_id", "5e2211c17a3e5a48f5497de3",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "db_name", "testdb",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "host", "cluster0.abcde.mongodb.net",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "client_id", "mdb_sa_id_abc123",
					),
				),
			},
			{
				Config: mongodbAtlasAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "name", "test-atlas-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "description", "updated atlas credential",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMongoDBAtlasAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessCredentialStep1() + mongodbAtlasAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_mongodb_atlas_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_credential.test", "name", "test-atlas-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_mongodb_atlas_access_credential.test", "group_id", "5e2211c17a3e5a48f5497de3",
					),
				),
			},
		},
	})
}

func mongodbAtlasAccessCredentialStep1() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-cred"
  description    = "test atlas credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
  client_id      = "mdb_sa_id_abc123"
  client_secret  = "test-client-secret-123"
}
`
}

func mongodbAtlasAccessCredentialStep2() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-cred-updated"
  description    = "updated atlas credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
  client_id      = "mdb_sa_id_abc123"
  client_secret  = "test-client-secret-123"
}
`
}

const mongodbAtlasAccessCredentialDataSource = `
data "hush_mongodb_atlas_access_credential" "test" {
  id = hush_mongodb_atlas_access_credential.test.id
}
`

// Negative test: validateAtlasAuth (CustomizeDiff) must reject configs that
// don't specify exactly one auth method. These fail at plan time, before any
// request reaches the mock.
func TestAccResourceMongoDBAtlasAccessCredential_InvalidAuth(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// No auth method at all.
				Config:      mongodbAtlasAccessCredentialNoAuth(),
				ExpectError: regexp.MustCompile(`use either client_id`),
			},
			{
				// Service-account id without its secret.
				Config:      mongodbAtlasAccessCredentialClientIDOnly(),
				ExpectError: regexp.MustCompile(`client_id and client_secret must both be set`),
			},
		},
	})
}

// Negative test: deployment_ids is immutable (credutil.ForbidDeploymentIDsChange).
// Step 1 creates the credential; step 2 changes only deployment_ids and must
// error at plan time.
func TestAccResourceMongoDBAtlasAccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessCredentialStep1(),
			},
			{
				Config:      mongodbAtlasAccessCredentialDeploymentChanged(),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func mongodbAtlasAccessCredentialNoAuth() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-noauth"
  deployment_ids = ["` + mockDeploymentID + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
}
`
}

func mongodbAtlasAccessCredentialClientIDOnly() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-idonly"
  deployment_ids = ["` + mockDeploymentID + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
  client_id      = "mdb_sa_id_abc123"
}
`
}

// Identical to step 1 except deployment_ids, to isolate the immutability check.
func mongodbAtlasAccessCredentialDeploymentChanged() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-cred"
  description    = "test atlas credential"
  deployment_ids = ["` + mockDeploymentID2 + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
  client_id      = "mdb_sa_id_abc123"
  client_secret  = "test-client-secret-123"
}
`
}

// Exercises the API-key auth path (public_key + private_key), the other valid
// branch of validateAtlasAuth not covered by the service-account happy path.
func TestAccResourceMongoDBAtlasAccessCredential_APIKey(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessCredentialAPIKeyStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "name", "test-atlas-apikey",
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "public_key", "abcdefgh",
					),
					// API-key path -> client_id should be absent from state.
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "client_id", "",
					),
				),
			},
			{
				Config: mongodbAtlasAccessCredentialAPIKeyStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "description", "updated atlas api-key credential",
					),
				),
			},
		},
	})
}

// Exercises write-only secret rotation via _wo_version. Bumping the version
// must trigger Update to re-send the new secret; the framework's post-apply
// plan check verifies there is no perpetual diff on _wo or _wo_version.
func TestAccResourceMongoDBAtlasAccessCredential_WOSecretRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("mongodb_atlas_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: mongodbAtlasAccessCredentialWOSecretStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "client_secret_wo_version", "1",
					),
				),
			},
			{
				// Same resource, bumped version, new wo secret -> rotation path.
				Config: mongodbAtlasAccessCredentialWOSecretStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_mongodb_atlas_access_credential.test", "client_secret_wo_version", "2",
					),
				),
			},
		},
	})
}

func mongodbAtlasAccessCredentialAPIKeyStep1() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-apikey"
  description    = "test atlas api-key credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
  public_key     = "abcdefgh"
  private_key    = "test-private-key-123"
}
`
}

func mongodbAtlasAccessCredentialAPIKeyStep2() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name           = "test-atlas-apikey"
  description    = "updated atlas api-key credential"
  deployment_ids = ["` + mockDeploymentID + `"]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "testdb"
  host           = "cluster0.abcde.mongodb.net"
  public_key     = "abcdefgh"
  private_key    = "test-private-key-123"
}
`
}

func mongodbAtlasAccessCredentialWOSecretStep1() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name                     = "test-atlas-wo-rotate"
  deployment_ids           = ["` + mockDeploymentID + `"]
  group_id                 = "5e2211c17a3e5a48f5497de3"
  db_name                  = "testdb"
  host                     = "cluster0.abcde.mongodb.net"
  client_id                = "mdb_sa_id_abc123"
  client_secret_wo         = "initial-secret"
  client_secret_wo_version = "1"
}
`
}

func mongodbAtlasAccessCredentialWOSecretStep2() string {
	return `
resource "hush_mongodb_atlas_access_credential" "test" {
  name                     = "test-atlas-wo-rotate"
  deployment_ids           = ["` + mockDeploymentID + `"]
  group_id                 = "5e2211c17a3e5a48f5497de3"
  db_name                  = "testdb"
  host                     = "cluster0.abcde.mongodb.net"
  client_id                = "mdb_sa_id_abc123"
  client_secret_wo         = "rotated-secret"
  client_secret_wo_version = "2"
}
`
}
