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
	cldCred          = "hush_cloudinary_access_credential.test"
	mockCldAccountID = "abcde1fg-2345-6789-0abc-d1efg2hi3jkl"
	mockCldAccount2  = "other-account-2"
	mockCldAPIKey    = "123456789012345"
	mockCldAPISecret = "mock-cld-root-secret"
)

// cldSent holds, per credential name, the api_key and api_secret the last
// accepted create or PATCH sent, each nil when it sent none.
var cldSent sync.Map

// cldRefused names the credentials whose next PATCH the mock refuses with 409,
// as the API refuses an account change while a policy uses the credential.
var cldRefused sync.Map

// cldSyncFails names the credentials whose next PATCH the mock takes, then
// reports in error status, as when the access manager refuses the key.
var cldSyncFails sync.Map

// cldStored holds, per credential id, a copy of what the mock last stored: the
// mock merges a PATCH before the hooks run, and a refused one changes nothing.
var cldStored sync.Map

type cldSentKeys struct{ key, secret any }

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Cloudinary; the routes are SendGrid's.
		ms.AliasRoutes("access_credentials/sendgrid", "access_credentials/cloudinary")
		// The Hush API never stores or returns api_key or api_secret, so dropping
		// them here makes their presence mean this request sent them.
		record := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["type"] = "cloudinary"
			if name, ok := obj["name"].(string); ok {
				cldSent.Store(name, cldSentKeys{obj["api_key"], obj["api_secret"]})
			}
			delete(obj, "api_key")
			delete(obj, "api_secret")
			return nil
		}
		ms.OnOperation("access_credentials/cloudinary", testutil.OpCreate, record)
		ms.OnOperation("access_credentials/cloudinary", testutil.OpUpdate, record)
		inject := func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			name, _ := obj["name"].(string)
			id, _ := obj["id"].(string)
			if op == testutil.OpUpdate {
				if _, ok := cldRefused.LoadAndDelete(name); ok {
					prev, _ := cldStored.Load(id)
					clear(obj)
					maps.Copy(obj, prev.(map[string]any))
					cldSent.Delete(name)
					return &testutil.HookError{Status: http.StatusConflict, Detail: "Resource is busy"}
				}
				obj["status"], obj["status_detail"] = "ok", ""
				if _, ok := cldSyncFails.LoadAndDelete(name); ok {
					obj["status"], obj["status_detail"] = "error", "the access manager refused the account management key"
				}
			}
			cldStored.Store(id, maps.Clone(obj))
			return nil
		}
		ms.OnOperation("access_credentials/cloudinary", testutil.OpCreate, inject)
		ms.OnOperation("access_credentials/cloudinary", testutil.OpUpdate, inject)
	})
}

// checkCldSent asserts what api_key and api_secret the last request for the
// named credential sent: the given values, or none where nil.
func checkCldSent(name string, key, secret any) resource.TestCheckFunc {
	return func(*terraform.State) error {
		got, ok := cldSent.Load(name)
		if !ok {
			return fmt.Errorf("no request for credential %q reached the API", name)
		}
		want := cldSentKeys{key, secret}
		if got != want {
			return fmt.Errorf("credential %q: the last request sent %+v, want %+v", name, got, want)
		}
		return nil
	}
}

func TestAccResourceCloudinaryAccessCredential(t *testing.T) {
	const renamed = "test-cld-cred-renamed"
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(cldCred, "id", regexp.MustCompile(`^acr-.+$`)),
					resource.TestCheckResourceAttr(cldCred, "name", "test-cld-cred"),
					resource.TestCheckResourceAttr(cldCred, "description", "test cloudinary credential"),
					resource.TestCheckResourceAttr(cldCred, "account_id", mockCldAccountID),
					resource.TestCheckResourceAttr(cldCred, "type", "cloudinary"),
					checkSecretStoreID(cldCred),
					checkCldSent("test-cld-cred", mockCldAPIKey, mockCldAPISecret),
					recordID(cldCred, &id),
				),
			},
			{
				// a change that leaves the key alone does not send it: that
				// would issue new credentials to every policy
				Config: cloudinaryAccessCredential(renamed, mockCldAccountID,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldCred, "name", renamed),
					checkCldSent(renamed, nil, nil),
					checkIDUnchanged(cldCred, &id),
				),
			},
			{
				// a new secret alone sends the secret alone
				Config: cloudinaryAccessCredential(renamed, mockCldAccountID,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret+"-rotated")),
				Check: resource.ComposeTestCheckFunc(
					checkCldSent(renamed, nil, mockCldAPISecret+"-rotated"),
					checkIDUnchanged(cldCred, &id),
				),
			},
			{
				// another account, with its own key and secret, in place
				Config: cloudinaryAccessCredential(renamed, mockCldAccount2,
					cldKey("999"), cldSecret(mockCldAPISecret+"-account2")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldCred, "account_id", mockCldAccount2),
					checkCldSent(renamed, "999", mockCldAPISecret+"-account2"),
					checkIDUnchanged(cldCred, &id),
				),
			},
			{
				Config: cloudinaryAccessCredential(renamed, mockCldAccount2,
					cldKeyWO("999", "1"), cldSecretWO(mockCldAPISecret+"-wo1", "1")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldCred, "api_key", ""),
					resource.TestCheckNoResourceAttr(cldCred, "api_key_wo"),
					resource.TestCheckResourceAttr(cldCred, "api_key_wo_version", "1"),
					resource.TestCheckResourceAttr(cldCred, "api_secret", ""),
					resource.TestCheckNoResourceAttr(cldCred, "api_secret_wo"),
					resource.TestCheckResourceAttr(cldCred, "api_secret_wo_version", "1"),
					checkCldSent(renamed, "999", mockCldAPISecret+"-wo1"),
					checkIDUnchanged(cldCred, &id),
				),
			},
			{
				// a bumped secret version sends the secret alone
				Config: cloudinaryAccessCredential(renamed, mockCldAccount2,
					cldKeyWO("999", "1"), cldSecretWO(mockCldAPISecret+"-wo2", "2")),
				Check: resource.ComposeTestCheckFunc(
					checkCldSent(renamed, nil, mockCldAPISecret+"-wo2"),
					checkIDUnchanged(cldCred, &id),
				),
			},
			{
				Config: cloudinaryAccessCredential(renamed, mockCldAccount2,
					cldKeyWO("998", "2"), cldSecretWO(mockCldAPISecret+"-wo2", "2")),
				Check: resource.ComposeTestCheckFunc(
					checkCldSent(renamed, "998", nil),
					checkIDUnchanged(cldCred, &id),
				),
			},
		},
	})
}

// Each write-only pair rotates on its version alone, and the plan after the
// rotation is empty.
func TestAccResourceCloudinaryAccessCredential_WriteOnlyRotation(t *testing.T) {
	const name = "test-cld-cred-rotation"
	config := func(keyVer, secretVer string) string {
		return cloudinaryAccessCredential(name, mockCldAccountID,
			cldKeyWO("key-"+keyVer, keyVer), cldSecretWO("secret-"+secretVer, secretVer))
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: config("1", "1"),
				Check:  checkCldSent(name, "key-1", "secret-1"),
			},
			{
				Config: config("2", "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldCred, "api_key_wo_version", "2"),
					checkCldSent(name, "key-2", nil),
				),
			},
			{
				Config:   config("2", "1"),
				PlanOnly: true,
			},
			{
				Config: config("2", "2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(cldCred, "api_secret_wo_version", "2"),
					checkCldSent(name, nil, "secret-2"),
				),
			},
			{
				Config:   config("2", "2"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceCloudinaryAccessCredential_AccountChangeNeedsKeyAndSecret(t *testing.T) {
	const name = "test-cld-cred-account"
	wantErr := regexp.MustCompile(`changing account_id requires a new api_key and a new api_secret`)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessCredential(name, mockCldAccountID,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret)),
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccount2,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret)),
				ExpectError: wantErr,
			},
			{
				// a key of another account comes with its own secret
				Config: cloudinaryAccessCredential(name, mockCldAccount2,
					cldKey("999"), cldSecret(mockCldAPISecret)),
				ExpectError: wantErr,
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccount2,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret+"-2")),
				ExpectError: wantErr,
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccountID,
					cldKeyWO(mockCldAPIKey, "1"), cldSecretWO(mockCldAPISecret, "1")),
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccount2,
					cldKeyWO(mockCldAPIKey, "2"), cldSecretWO(mockCldAPISecret, "1")),
				ExpectError: wantErr,
			},
		},
	})
}

// A refused PATCH leaves the prior key and secret in the state, so the same
// config retries a refused account or key change.
func TestAccResourceCloudinaryAccessCredential_RefusedUpdateRetries(t *testing.T) {
	const name = "test-cld-cred-refused"
	refuseNext := func() { cldRefused.Store(name, true) }
	account2 := cloudinaryAccessCredential(name, mockCldAccount2,
		cldKey("999"), cldSecret(mockCldAPISecret+"-account2"))
	writeOnly := cloudinaryAccessCredential(name, mockCldAccount2,
		cldKeyWO("999", "1"), cldSecretWO(mockCldAPISecret+"-wo", "1"))
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessCredential(name, mockCldAccountID,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret)),
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
					resource.TestCheckResourceAttr(cldCred, "account_id", mockCldAccount2),
					checkCldSent(name, "999", mockCldAPISecret+"-account2"),
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
					resource.TestCheckResourceAttr(cldCred, "api_key_wo_version", "1"),
					resource.TestCheckResourceAttr(cldCred, "api_secret_wo_version", "1"),
					checkCldSent(name, "999", mockCldAPISecret+"-wo"),
				),
			},
		},
	})
}

// A failed status wait follows an accepted PATCH: the state keeps the new
// secret, and only a new secret version sends it again.
func TestAccResourceCloudinaryAccessCredential_FailedSyncKeepsPlan(t *testing.T) {
	const name = "test-cld-cred-sync-fails"
	version := func(v string) string {
		return cloudinaryAccessCredential(name, mockCldAccountID,
			cldKeyWO(mockCldAPIKey, "1"), cldSecretWO(mockCldAPISecret+"-wo"+v, v))
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: version("1"),
			},
			{
				PreConfig:   func() { cldSyncFails.Store(name, true) },
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
				Check:  checkCldSent(name, nil, mockCldAPISecret+"-wo3"),
			},
		},
	})
}

func TestAccResourceCloudinaryAccessCredential_Validation(t *testing.T) {
	const name = "test-cld-cred-bad"
	key, secret := cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret)
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// the access manager puts the account in API paths
				Config:      cloudinaryAccessCredential(name, "acct/1", key, secret),
				ExpectError: regexp.MustCompile(`account_id must be 1 to 128 letters, digits, '_' or '-'`),
			},
			{
				Config:      cloudinaryAccessCredential(name, "", key, secret),
				ExpectError: regexp.MustCompile(`account_id must be 1 to 128 letters, digits, '_' or '-'`),
			},
			{
				// a pasted key often ends in a newline
				Config:      cloudinaryAccessCredential(name, mockCldAccountID, `api_key = "123\n"`, secret),
				ExpectError: regexp.MustCompile(`api_key cannot contain whitespace or control characters`),
			},
			{
				Config:      cloudinaryAccessCredential(name, mockCldAccountID, key, `api_secret = "s e"`),
				ExpectError: regexp.MustCompile(`api_secret cannot contain whitespace or control characters`),
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccountID, key,
					`api_secret_wo = "s\n"`+"\n"+`api_secret_wo_version = "1"`),
				ExpectError: regexp.MustCompile(`api_secret_wo cannot contain whitespace or control characters`),
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccountID,
					`api_key = "`+fmt.Sprintf("%0257d", 0)+`"`, secret),
				ExpectError: regexp.MustCompile(`api_key must be 1 to 256 characters`),
			},
			{
				Config: cloudinaryAccessCredential(name, mockCldAccountID,
					`api_key = "a"`+"\n"+`api_key_wo = "b"`+"\n"+`api_key_wo_version = "1"`, secret),
				ExpectError: regexp.MustCompile(`only one of`),
			},
			{
				Config:      cloudinaryAccessCredential(name, mockCldAccountID, "", secret),
				ExpectError: regexp.MustCompile(`one of .*api_key.*api_key_wo.* must be specified`),
			},
			{
				Config:      cloudinaryAccessCredential(name, mockCldAccountID, key, ""),
				ExpectError: regexp.MustCompile(`one of .*api_secret.*api_secret_wo.* must be specified`),
			},
		},
	})
}

func TestAccResourceCloudinaryAccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessCredentialStep1(),
			},
			{
				Config: regexp.MustCompile(regexp.QuoteMeta(mockDeploymentID)).
					ReplaceAllString(cloudinaryAccessCredentialStep1(), mockDeploymentID2),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func TestAccResourceCloudinaryAccessCredential_Import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessCredential("test-cld-cred-import", mockCldAccountID,
					cldKey(mockCldAPIKey), cldSecret(mockCldAPISecret)),
			},
			{
				ResourceName:      cldCred,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_key", "api_key_wo", "api_key_wo_version",
					"api_secret", "api_secret_wo", "api_secret_wo_version",
				},
			},
		},
	})
}

// The data source reads back what the API stored, so it shows an update was sent.
func TestAccDataSourceCloudinaryAccessCredential(t *testing.T) {
	const data = "data.hush_cloudinary_access_credential.test"
	dataSource := `
data "hush_cloudinary_access_credential" "test" {
  id = hush_cloudinary_access_credential.test.id
}
`
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("cloudinary_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: cloudinaryAccessCredentialStep1() + dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(data, "id", regexp.MustCompile(`^acr-.+$`)),
					resource.TestCheckResourceAttr(data, "name", "test-cld-cred"),
					resource.TestCheckResourceAttr(data, "account_id", mockCldAccountID),
					resource.TestCheckResourceAttr(data, "type", "cloudinary"),
				),
			},
			{
				Config: cloudinaryAccessCredential("test-cld-cred-ds", mockCldAccount2,
					cldKey("999"), cldSecret(mockCldAPISecret+"-2")) + dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data, "name", "test-cld-cred-ds"),
					resource.TestCheckResourceAttr(data, "account_id", mockCldAccount2),
				),
			},
		},
	})
}

func cloudinaryAccessCredentialStep1() string {
	return `
resource "hush_cloudinary_access_credential" "test" {
  name            = "test-cld-cred"
  description     = "test cloudinary credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  account_id      = "` + mockCldAccountID + `"
  api_key         = "` + mockCldAPIKey + `"
  api_secret      = "` + mockCldAPISecret + `"
}
`
}

// cloudinaryAccessCredential renders a credential with the given account and
// key and secret attributes.
func cloudinaryAccessCredential(name, accountID, key, secret string) string {
	return `
resource "hush_cloudinary_access_credential" "test" {
  name           = "` + name + `"
  deployment_ids = ["` + mockDeploymentID + `"]
  account_id     = "` + accountID + `"
  ` + key + `
  ` + secret + `
}
`
}

func cldKey(v string) string    { return `api_key = "` + v + `"` }
func cldSecret(v string) string { return `api_secret = "` + v + `"` }

func cldKeyWO(v, version string) string {
	return `api_key_wo = "` + v + `"` + "\n" + `api_key_wo_version = "` + version + `"`
}

func cldSecretWO(v, version string) string {
	return `api_secret_wo = "` + v + `"` + "\n" + `api_secret_wo_version = "` + version + `"`
}
