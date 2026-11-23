package acc_tests

import (
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const (
	r2Cred          = "hush_cloudflare_r2_access_credential.test"
	mockR2AccountID = "0123456789abcdef0123456789abcdef"
	mockR2Account2  = "fedcba9876543210fedcba9876543210"
	mockR2APIToken  = "mock-r2-root-token"
)

// r2SentTokens holds, per credential name, the api_token the last accepted
// create or PATCH sent, or nil when it sent none.
var r2SentTokens sync.Map

// r2Refused names the credentials whose next PATCH the mock refuses with 409,
// as the API refuses an account change while a policy uses the credential.
var r2Refused sync.Map

// r2SyncFails names the credentials whose next PATCH the mock takes, then
// reports in error status, as when the access manager refuses the token.
var r2SyncFails sync.Map

// r2Stored holds, per credential id, a copy of what the mock last stored: the
// mock merges a PATCH before the hooks run, and a refused one changes nothing.
var r2Stored sync.Map

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Cloudflare R2; the routes are SendGrid's.
		ms.AliasRoutes("access_credentials/sendgrid", "access_credentials/cloudflare_r2")
		// The Hush API never stores or returns api_token, so dropping it here
		// makes its presence on the merged object mean this request sent it.
		record := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["type"] = "cloudflare_r2"
			if name, ok := obj["name"].(string); ok {
				r2SentTokens.Store(name, obj["api_token"])
			}
			delete(obj, "api_token")
			return nil
		}
		ms.OnOperation("access_credentials/cloudflare_r2", testutil.OpCreate, record)
		ms.OnOperation("access_credentials/cloudflare_r2", testutil.OpUpdate, record)
		inject := func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			name, _ := obj["name"].(string)
			id, _ := obj["id"].(string)
			if op == testutil.OpUpdate {
				if _, ok := r2Refused.LoadAndDelete(name); ok {
					prev, _ := r2Stored.Load(id)
					clear(obj)
					maps.Copy(obj, prev.(map[string]any))
					r2SentTokens.Delete(name)
					return &testutil.HookError{Status: http.StatusConflict, Detail: "Resource is busy"}
				}
				obj["status"], obj["status_detail"] = "ok", ""
				if _, ok := r2SyncFails.LoadAndDelete(name); ok {
					obj["status"], obj["status_detail"] = "error", "the access manager refused the root token"
				}
			}
			r2Stored.Store(id, maps.Clone(obj))
			return nil
		}
		ms.OnOperation("access_credentials/cloudflare_r2", testutil.OpCreate, inject)
		ms.OnOperation("access_credentials/cloudflare_r2", testutil.OpUpdate, inject)
	})
}

// checkR2SentToken asserts what api_token the last request for the named
// credential sent: want, or none when want is nil.
func checkR2SentToken(name string, want any) resource.TestCheckFunc {
	return func(*terraform.State) error {
		got, ok := r2SentTokens.Load(name)
		if !ok {
			return fmt.Errorf("no request for credential %q reached the API", name)
		}
		if got != want {
			return fmt.Errorf("credential %q: the last request sent api_token %v, want %v", name, got, want)
		}
		return nil
	}
}

func TestAccResourceCloudflareR2AccessCredential(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(r2Cred, "id", regexp.MustCompile(`^acr-.+$`)),
					resource.TestCheckResourceAttr(r2Cred, "name", "test-r2-cred"),
					resource.TestCheckResourceAttr(r2Cred, "description", "test r2 credential"),
					resource.TestCheckResourceAttr(r2Cred, "account_id", mockR2AccountID),
					resource.TestCheckResourceAttr(r2Cred, "type", "cloudflare_r2"),
					checkSecretStoreID(r2Cred),
					checkR2SentToken("test-r2-cred", mockR2APIToken),
					recordID(r2Cred, &id),
				),
			},
			{
				// a change that leaves the token alone does not send it:
				// that would issue new credentials to every policy
				Config: cloudflareR2AccessCredential("test-r2-cred-renamed", mockR2AccountID,
					`api_token = "`+mockR2APIToken+`"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Cred, "name", "test-r2-cred-renamed"),
					checkR2SentToken("test-r2-cred-renamed", nil),
					checkIDUnchanged(r2Cred, &id),
				),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-renamed", mockR2AccountID,
					`api_token = "`+mockR2APIToken+`-rotated"`),
				Check: resource.ComposeTestCheckFunc(
					checkR2SentToken("test-r2-cred-renamed", mockR2APIToken+"-rotated"),
					checkIDUnchanged(r2Cred, &id),
				),
			},
			{
				// another account, with its own root token, in place
				Config: cloudflareR2AccessCredential("test-r2-cred-renamed", mockR2Account2,
					`api_token = "`+mockR2APIToken+`-account2"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Cred, "account_id", mockR2Account2),
					checkR2SentToken("test-r2-cred-renamed", mockR2APIToken+"-account2"),
					checkIDUnchanged(r2Cred, &id),
				),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-renamed", mockR2Account2,
					`api_token_wo = "`+mockR2APIToken+`-wo1"`+"\n"+`api_token_wo_version = "1"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Cred, "api_token", ""),
					resource.TestCheckNoResourceAttr(r2Cred, "api_token_wo"),
					resource.TestCheckResourceAttr(r2Cred, "api_token_wo_version", "1"),
					checkR2SentToken("test-r2-cred-renamed", mockR2APIToken+"-wo1"),
					checkIDUnchanged(r2Cred, &id),
				),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-renamed", mockR2Account2,
					`api_token_wo = "`+mockR2APIToken+`-wo2"`+"\n"+`api_token_wo_version = "2"`),
				Check: resource.ComposeTestCheckFunc(
					checkR2SentToken("test-r2-cred-renamed", mockR2APIToken+"-wo2"),
					checkIDUnchanged(r2Cred, &id),
				),
			},
		},
	})
}

func TestAccResourceCloudflareR2AccessCredential_AccountChangeNeedsToken(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-account", mockR2AccountID,
					`api_token = "`+mockR2APIToken+`"`),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-account", mockR2Account2,
					`api_token = "`+mockR2APIToken+`"`),
				ExpectError: regexp.MustCompile(`changing account_id requires a new api_token`),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-account", mockR2AccountID,
					`api_token_wo = "`+mockR2APIToken+`"`+"\n"+`api_token_wo_version = "1"`),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-account", mockR2Account2,
					`api_token_wo = "`+mockR2APIToken+`"`+"\n"+`api_token_wo_version = "1"`),
				ExpectError: regexp.MustCompile(`changing account_id requires a new api_token`),
			},
		},
	})
}

// A refused PATCH leaves the prior token in the state, so the same config
// retries a refused account or token change.
func TestAccResourceCloudflareR2AccessCredential_RefusedUpdateRetries(t *testing.T) {
	const name = "test-r2-cred-refused"
	refuseNext := func() { r2Refused.Store(name, true) }
	account2 := cloudflareR2AccessCredential(name, mockR2Account2,
		`api_token = "`+mockR2APIToken+`-account2"`)
	writeOnly := cloudflareR2AccessCredential(name, mockR2Account2,
		`api_token_wo = "`+mockR2APIToken+`-wo"`+"\n"+`api_token_wo_version = "1"`)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredential(name, mockR2AccountID,
					`api_token = "`+mockR2APIToken+`"`),
			},
			{
				// a policy uses the credential
				PreConfig:   refuseNext,
				Config:      account2,
				ExpectError: regexp.MustCompile(`Resource is busy`),
			},
			{
				// the policy is gone
				Config: account2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Cred, "account_id", mockR2Account2),
					checkR2SentToken(name, mockR2APIToken+"-account2"),
				),
			},
			{
				PreConfig:   refuseNext,
				Config:      writeOnly,
				ExpectError: regexp.MustCompile(`Resource is busy`),
			},
			{
				Config: writeOnly,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(r2Cred, "api_token_wo_version", "1"),
					checkR2SentToken(name, mockR2APIToken+"-wo"),
				),
			},
		},
	})
}

// A failed status wait follows an accepted PATCH: the state keeps the new
// token, and only a new token version sends it again.
func TestAccResourceCloudflareR2AccessCredential_FailedSyncKeepsPlan(t *testing.T) {
	const name = "test-r2-cred-sync-fails"
	version := func(v string) string {
		return cloudflareR2AccessCredential(name, mockR2AccountID,
			`api_token_wo = "`+mockR2APIToken+`-wo`+v+`"`+"\n"+`api_token_wo_version = "`+v+`"`)
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: version("1"),
			},
			{
				PreConfig:   func() { r2SyncFails.Store(name, true) },
				Config:      version("2"),
				ExpectError: regexp.MustCompile(`resource entered error status`),
			},
			{
				// sending it again would issue new credentials to every policy
				Config:   version("2"),
				PlanOnly: true,
			},
			{
				Config: version("3"),
				Check:  checkR2SentToken(name, mockR2APIToken+"-wo3"),
			},
		},
	})
}

func TestAccResourceCloudflareR2AccessCredential_Validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-bad", "0123456789ABCDEF0123456789ABCDEF",
					`api_token = "`+mockR2APIToken+`"`),
				ExpectError: regexp.MustCompile(`account_id must be 32 lowercase hex characters`),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-bad", "0123456789abcdef",
					`api_token = "`+mockR2APIToken+`"`),
				ExpectError: regexp.MustCompile(`account_id must be 32 lowercase hex characters`),
			},
			{
				// a pasted token often ends in a newline
				Config: cloudflareR2AccessCredential("test-r2-cred-bad", mockR2AccountID,
					`api_token = "`+mockR2APIToken+`\n"`),
				ExpectError: regexp.MustCompile(`api_token cannot contain whitespace or control characters`),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-bad", mockR2AccountID,
					`api_token_wo = "`+mockR2APIToken+`\n"`+"\n"+`api_token_wo_version = "1"`),
				ExpectError: regexp.MustCompile(`api_token_wo cannot contain whitespace or control characters`),
			},
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-bad", mockR2AccountID,
					`api_token = "a"`+"\n"+`api_token_wo = "b"`+"\n"+`api_token_wo_version = "1"`),
				ExpectError: regexp.MustCompile(`only one of`),
			},
			{
				Config:      cloudflareR2AccessCredential("test-r2-cred-bad", mockR2AccountID, ""),
				ExpectError: regexp.MustCompile(`one of .*api_token.*api_token_wo.* must be specified`),
			},
		},
	})
}

func TestAccResourceCloudflareR2AccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredentialStep1(),
			},
			{
				Config: regexp.MustCompile(regexp.QuoteMeta(mockDeploymentID)).
					ReplaceAllString(cloudflareR2AccessCredentialStep1(), mockDeploymentID2),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func TestAccResourceCloudflareR2AccessCredential_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredential("test-r2-cred-import", mockR2AccountID,
					`api_token = "`+mockR2APIToken+`"`),
			},
			{
				ResourceName:            r2Cred,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_token", "api_token_wo", "api_token_wo_version"},
			},
		},
	})
}

func TestAccDataSourceCloudflareR2AccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudflare_r2_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudflareR2AccessCredentialStep1() + `
data "hush_cloudflare_r2_access_credential" "test" {
  id = hush_cloudflare_r2_access_credential.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_cloudflare_r2_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_credential.test", "name", "test-r2-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_credential.test", "account_id", mockR2AccountID,
					),
					resource.TestCheckResourceAttr(
						"data.hush_cloudflare_r2_access_credential.test", "type", "cloudflare_r2",
					),
				),
			},
		},
	})
}

func cloudflareR2AccessCredentialStep1() string {
	return `
resource "hush_cloudflare_r2_access_credential" "test" {
  name            = "test-r2-cred"
  description     = "test r2 credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  account_id      = "` + mockR2AccountID + `"
  api_token       = "` + mockR2APIToken + `"
}
`
}

// cloudflareR2AccessCredential renders a credential with the given account
// and token attributes.
func cloudflareR2AccessCredential(name, accountID, token string) string {
	return `
resource "hush_cloudflare_r2_access_credential" "test" {
  name           = "` + name + `"
  deployment_ids = ["` + mockDeploymentID + `"]
  account_id     = "` + accountID + `"
  ` + token + `
}
`
}
