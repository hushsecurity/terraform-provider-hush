package acc_tests

import (
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const (
	mockGithubAppClientID       = "Iv23liMockClientId01"
	mockGithubAppInstallationID = "176836371"
	mockGithubAppPrivateKey     = "mock-github-app-private-key"
	// An older client ID format, with a dot.
	mockGithubAppLegacyClientID = "Iv1.8a61f9b3a7aba766"
)

// githubAppPatchedKeys records, per credential name, the private_key on the
// stored credential after each update. The create stored the old key, so the
// new one appearing here means the PATCH sent it.
var githubAppPatchedKeys sync.Map

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("access_credentials/github_app", testutil.OpUpdate, func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			if name, ok := obj["name"].(string); ok {
				githubAppPatchedKeys.Store(name, obj["private_key"])
			}
			return nil
		})
	})
}

// checkGithubAppPatchedKey asserts that the last update of the named
// credential sent the given private key.
func checkGithubAppPatchedKey(name, want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		got, ok := githubAppPatchedKeys.Load(name)
		if !ok {
			return fmt.Errorf("no update of credential %q reached the API", name)
		}
		if got != want {
			return fmt.Errorf("credential %q: the update sent private_key %v, want %q", name, got, want)
		}
		return nil
	}
}

func TestAccResourceGithubAppAccessCredential(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_github_app_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "name", "test-github-app-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "client_id", mockGithubAppClientID,
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "installation_id", mockGithubAppInstallationID,
					),
					checkSecretStoreID("hush_github_app_access_credential.test"),
					recordID("hush_github_app_access_credential.test", &id),
				),
			},
			{
				// changes the plain private_key in place
				Config: githubAppAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "name", "test-github-app-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "description", "updated github app credential",
					),
					// the numeric App ID is accepted as the client id too
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "client_id", "5539289",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "installation_id", "42",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "private_key", mockGithubAppPrivateKey+"-rotated",
					),
					checkGithubAppPatchedKey("test-github-app-cred-updated", mockGithubAppPrivateKey+"-rotated"),
					checkIDUnchanged("hush_github_app_access_credential.test", &id),
				),
			},
			{
				// switches from private_key to private_key_wo in place, with
				// an older, dotted client ID
				Config: githubAppAccessCredentialStep3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "client_id", mockGithubAppLegacyClientID,
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "private_key", "",
					),
					resource.TestCheckNoResourceAttr(
						"hush_github_app_access_credential.test", "private_key_wo",
					),
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "private_key_wo_version", "1",
					),
					checkGithubAppPatchedKey("test-github-app-cred-updated", mockGithubAppPrivateKey+"-wo"),
					checkIDUnchanged("hush_github_app_access_credential.test", &id),
				),
			},
		},
	})
}

func TestAccResourceGithubAppAccessCredential_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessCredentialWithIDs(mockGithubAppClientID, mockGithubAppInstallationID),
			},
			{
				ResourceName:            "hush_github_app_access_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "private_key_wo", "private_key_wo_version"},
			},
		},
	})
}

func TestAccDataSourceGithubAppAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessCredentialStep1() + githubAppAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_github_app_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_github_app_access_credential.test", "client_id", mockGithubAppClientID,
					),
					resource.TestCheckResourceAttr(
						"data.hush_github_app_access_credential.test", "installation_id", mockGithubAppInstallationID,
					),
				),
			},
		},
	})
}

// Write-only private key rotation: bumping private_key_wo_version must
// trigger Update and converge with no perpetual diff (each step's post-apply
// plan must be empty).
func TestAccResourceGithubAppAccessCredential_WOPrivateKeyRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessCredentialWO("1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "private_key_wo_version", "1",
					),
					resource.TestCheckNoResourceAttr(
						"hush_github_app_access_credential.test", "private_key_wo",
					),
				),
			},
			{
				Config: githubAppAccessCredentialWO("2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_github_app_access_credential.test", "private_key_wo_version", "2",
					),
					checkGithubAppPatchedKey("test-github-app-cred-wo", mockGithubAppPrivateKey+"-2"),
				),
			},
		},
	})
}

// Negative tests: the private_key ExactlyOneOf and RequiredWith groups and the
// field formats. The deployment_ids CustomizeDiff
// (credutil.ForbidDeploymentIDsChange) is covered by _DeploymentIDsImmutable.
func TestAccResourceGithubAppAccessCredential_Validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      githubAppAccessCredentialKeys(`private_key = "a"` + "\n" + `private_key_wo = "b"` + "\n" + `private_key_wo_version = "1"`),
				ExpectError: regexp.MustCompile(`only one of`),
			},
			{
				Config:      githubAppAccessCredentialKeys(""),
				ExpectError: regexp.MustCompile(`one of .*private_key.*private_key_wo.* must be specified`),
			},
			{
				Config:      githubAppAccessCredentialKeys(`private_key_wo = "b"`),
				ExpectError: regexp.MustCompile(`"private_key_wo": all of\s+.private_key_wo,private_key_wo_version.\s+must\s+be\s+specified`),
			},
			{
				Config:      githubAppAccessCredentialKeys(`private_key = "a"` + "\n" + `private_key_wo_version = "1"`),
				ExpectError: regexp.MustCompile(`"private_key_wo_version": all of\s+.private_key_wo,private_key_wo_version.\s+must\s+be\s+specified`),
			},
			{
				Config:      githubAppAccessCredentialWithIDs(mockGithubAppClientID, "12a"),
				ExpectError: regexp.MustCompile(`installation_id must be a positive integer`),
			},
			{
				Config:      githubAppAccessCredentialWithIDs(mockGithubAppClientID, "0"),
				ExpectError: regexp.MustCompile(`installation_id must be a positive integer`),
			},
			{
				Config:      githubAppAccessCredentialWithIDs(mockGithubAppClientID, "012"),
				ExpectError: regexp.MustCompile(`installation_id must be a positive integer of at most 19 digits, without leading zeros`),
			},
			{
				Config:      githubAppAccessCredentialWithIDs(mockGithubAppClientID, "12345678901234567890"),
				ExpectError: regexp.MustCompile(`installation_id must be a positive integer`),
			},
			{
				Config:      githubAppAccessCredentialWithIDs("Iv23 li", mockGithubAppInstallationID),
				ExpectError: regexp.MustCompile(`client_id must be 1-64 letters, digits or dots`),
			},
			{
				Config:      githubAppAccessCredentialWithIDs("Iv1/8a61", mockGithubAppInstallationID),
				ExpectError: regexp.MustCompile(`client_id must be 1-64 letters, digits or dots`),
			},
			{
				// a dotted client ID passes validation
				Config:             githubAppAccessCredentialWithIDs(mockGithubAppLegacyClientID, mockGithubAppInstallationID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccResourceGithubAppAccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("github_app_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: githubAppAccessCredentialStep1(),
			},
			{
				Config: regexp.MustCompile(regexp.QuoteMeta(mockDeploymentID)).
					ReplaceAllString(githubAppAccessCredentialStep1(), mockDeploymentID2),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func githubAppAccessCredentialStep1() string {
	return `
resource "hush_github_app_access_credential" "test" {
  name            = "test-github-app-cred"
  description     = "test github app credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  client_id       = "` + mockGithubAppClientID + `"
  installation_id = "` + mockGithubAppInstallationID + `"
  private_key     = "` + mockGithubAppPrivateKey + `"
}
`
}

const githubAppAccessCredentialStep2 = `
resource "hush_github_app_access_credential" "test" {
  name            = "test-github-app-cred-updated"
  description     = "updated github app credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  client_id       = "5539289"
  installation_id = "42"
  private_key     = "` + mockGithubAppPrivateKey + `-rotated"
}
`

const githubAppAccessCredentialStep3 = `
resource "hush_github_app_access_credential" "test" {
  name                   = "test-github-app-cred-updated"
  description            = "updated github app credential"
  deployment_ids         = ["` + mockDeploymentID + `"]
  secret_store_id        = "sst-mock-store-1"
  client_id              = "` + mockGithubAppLegacyClientID + `"
  installation_id        = "42"
  private_key_wo         = "` + mockGithubAppPrivateKey + `-wo"
  private_key_wo_version = "1"
}
`

const githubAppAccessCredentialDataSource = `
data "hush_github_app_access_credential" "test" {
  id = hush_github_app_access_credential.test.id
}
`

func githubAppAccessCredentialWO(version string) string {
	return `
resource "hush_github_app_access_credential" "test" {
  name                   = "test-github-app-cred-wo"
  deployment_ids         = ["` + mockDeploymentID + `"]
  client_id              = "` + mockGithubAppClientID + `"
  installation_id        = "` + mockGithubAppInstallationID + `"
  private_key_wo         = "` + mockGithubAppPrivateKey + `-` + version + `"
  private_key_wo_version = "` + version + `"
}
`
}

// githubAppAccessCredentialKeys renders a credential with the given private
// key attributes.
func githubAppAccessCredentialKeys(keys string) string {
	return `
resource "hush_github_app_access_credential" "test" {
  name            = "test-github-app-cred-keys"
  deployment_ids  = ["` + mockDeploymentID + `"]
  client_id       = "` + mockGithubAppClientID + `"
  installation_id = "` + mockGithubAppInstallationID + `"
  ` + keys + `
}
`
}

func githubAppAccessCredentialWithIDs(clientID, installationID string) string {
	return `
resource "hush_github_app_access_credential" "test" {
  name            = "test-github-app-cred-ids"
  deployment_ids  = ["` + mockDeploymentID + `"]
  client_id       = "` + clientID + `"
  installation_id = "` + installationID + `"
  private_key     = "` + mockGithubAppPrivateKey + `"
}
`
}
