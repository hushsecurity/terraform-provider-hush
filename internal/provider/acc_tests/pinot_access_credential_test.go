package acc_tests

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

const (
	pinotCred         = "hush_pinot_access_credential.test"
	mockPinotHost     = "mock-pinot-controller.example.com"
	mockPinotBroker   = "http://mock-pinot-broker.example.com:8099"
	mockPinotUsername = "admin"
	mockPinotPassword = "mock-pinot-password"
)

// pinotPatchedPasswords holds, per credential id, the password its last PATCH
// left on the stored object.
var pinotPatchedPasswords sync.Map

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		// The published fixtures predate Pinot; the routes are Elasticsearch's.
		ms.AliasRoutes("access_credentials/elasticsearch", "access_credentials/pinot")
		// midgard stores broker_url without trailing slashes.
		normalize := func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
			obj["type"] = "pinot"
			if u, ok := obj["broker_url"].(string); ok {
				obj["broker_url"] = strings.TrimRight(u, "/")
			}
			return nil
		}
		ms.OnOperation("access_credentials/pinot", testutil.OpCreate, normalize)
		ms.OnOperation("access_credentials/pinot", testutil.OpUpdate, normalize)
		ms.OnOperation("access_credentials/pinot", testutil.OpUpdate,
			func(_ testutil.Operation, obj map[string]any) *testutil.HookError {
				if id, ok := obj["id"].(string); ok {
					pinotPatchedPasswords.Store(id, obj["password"])
				}
				return nil
			})
	})
}

// A trailing slash on broker_url is stored without it and does not show as a
// change; a later change is sent, and removing it unsets it.
func TestAccResourcePinotAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessCredential("test-pinot-cred", "test pinot credential",
					`broker_url = "`+mockPinotBroker+`/"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(pinotCred, "id", regexp.MustCompile(`^acr-.+$`)),
					resource.TestCheckResourceAttr(pinotCred, "name", "test-pinot-cred"),
					resource.TestCheckResourceAttr(pinotCred, "description", "test pinot credential"),
					resource.TestCheckResourceAttr(pinotCred, "host", mockPinotHost),
					resource.TestCheckResourceAttr(pinotCred, "port", "9000"),
					resource.TestCheckResourceAttr(pinotCred, "tls", "false"),
					resource.TestCheckResourceAttr(pinotCred, "broker_url", mockPinotBroker),
					resource.TestCheckResourceAttr(pinotCred, "type", "pinot"),
					checkSecretStoreID(pinotCred),
				),
			},
			{
				Config: pinotAccessCredential("test-pinot-cred-updated", "updated pinot credential",
					`broker_url = "https://pinot-broker.example.com"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pinotCred, "name", "test-pinot-cred-updated"),
					resource.TestCheckResourceAttr(pinotCred, "description", "updated pinot credential"),
					resource.TestCheckResourceAttr(pinotCred, "broker_url", "https://pinot-broker.example.com"),
				),
			},
			{
				Config: pinotAccessCredential("test-pinot-cred-updated", "updated pinot credential", ""),
				Check:  resource.TestCheckResourceAttr(pinotCred, "broker_url", ""),
			},
			{
				ResourceName:            pinotCred,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDataSourcePinotAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessCredential("test-pinot-cred", "test pinot credential",
					`broker_url = "`+mockPinotBroker+`"`) + `
data "hush_pinot_access_credential" "test" {
  id = hush_pinot_access_credential.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_pinot_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_pinot_access_credential.test", "name", "test-pinot-cred",
					),
					resource.TestCheckResourceAttr(
						"data.hush_pinot_access_credential.test", "broker_url", mockPinotBroker,
					),
				),
			},
		},
	})
}

// Bumping password_wo_version must trigger Update and converge with no
// perpetual diff.
func TestAccResourcePinotAccessCredential_WOPasswordRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessCredentialWO("secret-v1", "1"),
				Check:  resource.TestCheckResourceAttr(pinotCred, "password_wo_version", "1"),
			},
			{
				Config: pinotAccessCredentialWO("secret-v2", "2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pinotCred, "password_wo_version", "2"),
					checkPinotPatchedPassword("secret-v2"),
				),
			},
			{
				ResourceName:            pinotCred,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password_wo_version"},
			},
		},
	})
}

// Moving from password to password_wo is an in-place update that sends the
// write-only value.
func TestAccResourcePinotAccessCredential_PasswordToWO(t *testing.T) {
	var id string
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessCredentialFields(`password = "` + mockPinotPassword + `"`),
				Check:  recordID(pinotCred, &id),
			},
			{
				Config: pinotAccessCredentialFields(`
  password_wo         = "secret-wo"
  password_wo_version = "1"`),
				Check: resource.ComposeTestCheckFunc(
					checkIDUnchanged(pinotCred, &id),
					resource.TestCheckResourceAttr(pinotCred, "password", ""),
					resource.TestCheckResourceAttr(pinotCred, "password_wo_version", "1"),
					checkPinotPatchedPassword("secret-wo"),
				),
			},
		},
	})
}

// Every plan-time refusal of the connection fields. None reaches the mock.
func TestAccResourcePinotAccessCredential_FieldValidation(t *testing.T) {
	password := `password = "` + mockPinotPassword + `"`
	cases := map[string]struct {
		fields string
		err    string
	}{
		"port 0":     {password + "\n  port = 0", `expected "port" to be a valid port number`},
		"port 70000": {password + "\n  port = 70000", `expected "port" to be a valid port number`},
		"empty username": {
			`username = ""` + "\n  " + password,
			`expected "username" to not be an empty string`,
		},
		"long password": {
			`password = "` + strings.Repeat("p", 257) + `"`,
			`expected length of password to be in the range \(1 - 256\)`,
		},
		"long password_wo": {
			`password_wo = "` + strings.Repeat("p", 257) + `"` + "\n  password_wo_version = \"1\"",
			`expected length of password_wo to be in the range \(1 - 256\)`,
		},
		"both passwords": {
			password + "\n  password_wo = \"x\"\n  password_wo_version = \"1\"",
			`only one of .password,password_wo. can be specified`,
		},
		"no password": {"", `one of .password,password_wo. must be specified`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{
					{
						Config:      pinotAccessCredentialFields(c.fields),
						ExpectError: regexp.MustCompile(c.err),
					},
				},
			})
		})
	}
}

// checkPinotPatchedPassword asserts the credential's last PATCH carried want.
func checkPinotPatchedPassword(want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[pinotCred]
		if !ok {
			return fmt.Errorf("resource not found: %s", pinotCred)
		}
		got, _ := pinotPatchedPasswords.Load(rs.Primary.ID)
		if got != want {
			return fmt.Errorf("PATCH left password %v, want %q", got, want)
		}
		return nil
	}
}

func TestAccResourcePinotAccessCredential_BrokerURLValidation(t *testing.T) {
	const notURL = `broker_url must be an http\(s\) URL`
	const extra = `broker_url cannot carry a query or a fragment`
	const userinfo = `broker_url cannot carry a username or password`
	const space = `broker_url cannot contain whitespace or control characters`
	for name, c := range map[string]struct{ url, err string }{
		"scheme":         {"ftp://pinot-broker.example.com", notURL},
		"host":           {"http://", notURL},
		"bare":           {"pinot-broker.example.com:8099", notURL},
		"non-numeric":    {"http://pinot-broker.example.com:abc", notURL},
		"port too big":   {"http://pinot-broker.example.com:99999", `broker_url has an invalid port`},
		"query":          {"http://pinot-broker.example.com:8099?x=1", extra},
		"empty query":    {"http://pinot-broker.example.com:8099?", extra},
		"fragment":       {"http://pinot-broker.example.com:8099#f", extra},
		"empty fragment": {"http://pinot-broker.example.com:8099#", extra},
		"password":       {"http://u:p@pinot-broker.example.com:8099", userinfo},
		"username":       {"http://u@pinot-broker.example.com:8099", userinfo},
		"trailing space": {"http://pinot-broker.example.com:8099/ ", space},
		"space in path":  {"http://pinot-broker.example.com:8099/a b", space},
		"zero width":     {"http://pinot-broker.example.com\u200b:8099", space},
		"port only":      {"http://:8099", notURL},
	} {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{
					{
						Config: pinotAccessCredential("test-pinot-cred", "test pinot credential",
							`broker_url = "`+c.url+`"`),
						ExpectError: regexp.MustCompile(c.err),
					},
				},
			})
		})
	}
}

func TestAccResourcePinotAccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("pinot_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: pinotAccessCredential("test-pinot-cred", "test pinot credential", ""),
			},
			{
				Config: strings.Replace(
					pinotAccessCredential("test-pinot-cred", "test pinot credential", ""),
					mockDeploymentID, mockDeploymentID2, 1),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func pinotAccessCredential(name, description, extra string) string {
	return `
resource "hush_pinot_access_credential" "test" {
  name            = "` + name + `"
  description     = "` + description + `"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "` + mockSecretStoreID + `"
  host            = "` + mockPinotHost + `"
  username        = "` + mockPinotUsername + `"
  password        = "` + mockPinotPassword + `"
  ` + extra + `
}
`
}

func pinotAccessCredentialWO(password, version string) string {
	return `
resource "hush_pinot_access_credential" "test" {
  name                = "test-pinot-wo-password"
  deployment_ids      = ["` + mockDeploymentID + `"]
  host                = "` + mockPinotHost + `"
  username            = "` + mockPinotUsername + `"
  password_wo         = "` + password + `"
  password_wo_version = "` + version + `"
}
`
}

// pinotAccessCredentialFields leaves the credential fields, and username when
// they name it, to the caller.
func pinotAccessCredentialFields(fields string) string {
	username := `username = "` + mockPinotUsername + `"`
	if strings.Contains(fields, "username") {
		username = ""
	}
	return `
resource "hush_pinot_access_credential" "test" {
  name           = "test-pinot-fields"
  deployment_ids = ["` + mockDeploymentID + `"]
  host           = "` + mockPinotHost + `"
  ` + username + `
  ` + fields + `
}
`
}
