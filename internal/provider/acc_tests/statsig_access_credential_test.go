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

const statsigCred = "hush_statsig_access_credential.test"

// statsigPatchedAPIKeys holds, per credential id, the api_key its last PATCH
// left on the stored object.
var statsigPatchedAPIKeys sync.Map

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Statsig; the routes are SendGrid's.
		ms.AliasRoutes("access_credentials/sendgrid", "access_credentials/statsig")
		setType := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["type"] = "statsig"
			return nil
		}
		ms.OnOperation("access_credentials/statsig", testutil.OpCreate, setType)
		ms.OnOperation("access_credentials/statsig", testutil.OpUpdate, setType)
		ms.OnOperation("access_credentials/statsig", testutil.OpUpdate,
			func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
				if id, ok := obj["id"].(string); ok {
					statsigPatchedAPIKeys.Store(id, obj["api_key"])
				}
				return nil
			})
	})
}

func TestAccResourceStatsigAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("statsig_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: statsigAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_statsig_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_statsig_access_credential.test", "name", "test-statsig-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_statsig_access_credential.test", "description", "test statsig credential",
					),
					checkSecretStoreID("hush_statsig_access_credential.test"),
				),
			},
			{
				Config: statsigAccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_statsig_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_statsig_access_credential.test", "name", "test-statsig-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_statsig_access_credential.test", "description", "updated statsig credential",
					),
					checkStatsigPatchedAPIKey("console-mock-statsig-api-key-2"),
				),
			},
		},
	})
}

func TestAccDataSourceStatsigAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("statsig_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: statsigAccessCredentialStep1() + statsigAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_statsig_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_statsig_access_credential.test", "name", "test-statsig-cred",
					),
				),
			},
		},
	})
}

func statsigAccessCredentialStep1() string {
	return `
resource "hush_statsig_access_credential" "test" {
  name            = "test-statsig-cred"
  description     = "test statsig credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  api_key         = "console-mock-statsig-api-key"
}
`
}

func statsigAccessCredentialStep2() string {
	return `
resource "hush_statsig_access_credential" "test" {
  name            = "test-statsig-cred-updated"
  description     = "updated statsig credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  api_key         = "console-mock-statsig-api-key-2"
}
`
}

const statsigAccessCredentialDataSource = `
data "hush_statsig_access_credential" "test" {
  id = hush_statsig_access_credential.test.id
}
`

// Bumping api_key_wo_version must trigger Update and converge with no
// perpetual diff.
func TestAccResourceStatsigAccessCredential_WOAPIKeyRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("statsig_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: statsigAccessCredentialWO("console-v1", "1"),
				Check:  resource.TestCheckResourceAttr(statsigCred, "api_key_wo_version", "1"),
			},
			{
				Config: statsigAccessCredentialWO("console-v2", "2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statsigCred, "api_key_wo_version", "2"),
					checkStatsigPatchedAPIKey("console-v2"),
				),
			},
			{
				ResourceName:            statsigCred,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key_wo_version"},
			},
		},
	})
}

func statsigAccessCredentialWO(apiKey, version string) string {
	return `
resource "hush_statsig_access_credential" "test" {
  name               = "test-statsig-wo"
  deployment_ids     = ["` + mockDeploymentID + `"]
  api_key_wo         = "` + apiKey + `"
  api_key_wo_version = "` + version + `"
}
`
}

func checkStatsigPatchedAPIKey(want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[statsigCred]
		if !ok {
			return fmt.Errorf("resource not found: %s", statsigCred)
		}
		got, _ := statsigPatchedAPIKeys.Load(rs.Primary.ID)
		if got != want {
			return fmt.Errorf("PATCH left api_key %v, want %q", got, want)
		}
		return nil
	}
}
