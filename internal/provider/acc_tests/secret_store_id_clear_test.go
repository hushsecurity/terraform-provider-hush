package acc_tests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// Mirror midgard: secret_store_id must be a valid sst- id or explicit null.
// The empty string reaches the mock only if the provider clears the field with
// "" instead of null, so rejecting it here makes every credential update test
// guard the three-state PATCH contract. Registered on the computed-fields key so
// it fires for every access-credential subtype.
func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("access_credential", testutil.OpUpdate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			if v, ok := obj["secret_store_id"]; ok {
				if s, isStr := v.(string); isStr && s == "" {
					return &testutil.HookError{
						Status: 400,
						Detail: "secret_store_id must be a valid sst- id or null",
					}
				}
			}
			return nil
		})
	})
}

// TestAccResourcePlaintextAccessCredentialSecretStoreIDClear verifies that
// detaching a credential from its secret store is an in-place update that sends
// null (not ""), so the credential keeps its id and the field is cleared.
func TestAccResourcePlaintextAccessCredentialSecretStoreIDClear(t *testing.T) {
	var id string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("plaintext_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: plaintextSecretStoreClearStep1,
				Check: resource.ComposeTestCheckFunc(
					checkSecretStoreID("hush_plaintext_access_credential.clear"),
					recordID("hush_plaintext_access_credential.clear", &id),
				),
			},
			{
				Config: plaintextSecretStoreClearStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_plaintext_access_credential.clear", "secret_store_id", "",
					),
					checkIDUnchanged("hush_plaintext_access_credential.clear", &id),
				),
			},
		},
	})
}

const plaintextSecretStoreClearStep1 = `
resource "hush_plaintext_access_credential" "clear" {
  name            = "test-plaintext-clear"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  secret          = "s3cr3t-value"
}
`

const plaintextSecretStoreClearStep2 = `
resource "hush_plaintext_access_credential" "clear" {
  name           = "test-plaintext-clear"
  deployment_ids = ["` + mockDeploymentID + `"]
  secret         = "s3cr3t-value"
}
`
