package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

// Mirror the API's rules on the OIDC fields, and pin the provider's own side of
// the contract. The hook runs on the merged object, so it sees exactly the state
// a request would leave behind, and it is registered from init() so it applies
// to every deployment update in the package rather than only to this test.
//
// Without it the mock accepts anything and each of the three mistakes below
// passes here while the real API refuses it.
func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("deployment", testutil.OpUpdate,
			func(op testutil.Operation, obj map[string]any) *testutil.HookError {
				// API rule: a change must not leave the deployment holding
				// both fields. Fires when a write of the list fails to clear
				// the singular field a deployment already holds.
				if obj["oidc_provider"] != nil && obj["oidc_providers"] != nil {
					return &testutil.HookError{
						Status: 422,
						Detail: "set oidc_provider or oidc_providers, not both",
					}
				}

				// API rule: the list has a minimum length, so removal is an
				// explicit null. Belt and braces -- removal takes the nil path
				// and Go marshals a nil slice to null on its own, so this only
				// fires for a non-nil empty slice. The property is covered by
				// TestUpdateDeploymentInput_OidcProvidersMarshaling; this keeps
				// the mock faithful to the API rather than to what the provider
				// happens to reach today.
				if entries, ok := obj["oidc_providers"].([]any); ok &&
					len(entries) == 0 {
					return &testutil.HookError{
						Status: 422,
						Detail: "oidc_providers must not be empty",
					}
				}

				// Provider contract rather than an API rule: this provider
				// writes the list and never the singular field. The API would
				// accept a lone singular value, so nothing else would notice a
				// regression to writing it.
				if obj["oidc_provider"] != nil {
					return &testutil.HookError{
						Status: 422,
						Detail: "provider must not write oidc_provider",
					}
				}

				return nil
			})
	})
}

// legacyOIDCDeploymentName marks the deployment the create hook below rewrites.
// Keyed on the name so the rewrite reaches one test and leaves every other
// deployment in the package alone.
const legacyOIDCDeploymentName = "tf-acc-legacy-oidc"

// Manufacture a deployment that holds the API's singular field, which this
// provider never writes and so could not otherwise appear in a test. It stands
// in for one created before the list shipped, and it is the only state in which
// failing to clear the singular field on update is observable at all.
//
// The hook runs before the object is stored and before the response is encoded,
// so both the stored deployment and the read-back carry the singular field.
func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("deployment", testutil.OpCreate,
			func(op testutil.Operation, obj map[string]any) *testutil.HookError {
				if obj["name"] != legacyOIDCDeploymentName {
					return nil
				}
				if entries, ok := obj["oidc_providers"].([]any); ok &&
					len(entries) == 1 {
					obj["oidc_provider"] = entries[0]
					delete(obj, "oidc_providers")
				}
				return nil
			})
	})
}

// TestAccResourceDeploymentOIDCMigratesFromTheSingularField is the migration
// this commit exists for: a deployment that already holds the singular field
// gains a second issuer. The update has to write the list and clear the
// singular in one request, because the API resolves an absent field from the
// stored deployment and refuses a change that would leave both set.
//
// Deleting the clear from deploymentUpdate makes step 2 fail here with the
// API's own "not both". No other test in the suite can see that.
func TestAccResourceDeploymentOIDCMigratesFromTheSingularField(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				// Created with one block; the hook moves it to the singular
				// field, so the stored deployment looks pre-list.
				Config: deploymentLegacyOIDCConfig(oidcIssuer),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.legacy",
						"oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.legacy",
						"oidc_provider.0.issuer", oidcIssuer),
				),
			},
			{
				// The second cluster joins. This is the request that has to
				// carry the list and a null singular together.
				Config: deploymentLegacyOIDCConfig(oidcIssuer, oidcIssuer2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.legacy",
						"oidc_provider.#", "2"),
					resource.TestCheckResourceAttr("hush_deployment.legacy",
						"oidc_provider.1.issuer", oidcIssuer2),
				),
			},
		},
	})
}

// TestAccResourceDeploymentMultipleOIDCProviders walks the block from one
// issuer to two and back, then removes it.
//
// Every step differs only in the OIDC blocks, so a failure cannot be blamed on
// anything else changing at the same time.
func TestAccResourceDeploymentMultipleOIDCProviders(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentMultiOIDCConfig(oidcIssuer),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.0.issuer", oidcIssuer),
				),
			},
			{
				// The second cluster joins.
				Config: deploymentMultiOIDCConfig(oidcIssuer, oidcIssuer2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.#", "2"),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.0.issuer", oidcIssuer),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.1.issuer", oidcIssuer2),
				),
			},
			{
				// The old cluster is gone.
				Config: deploymentMultiOIDCConfig(oidcIssuer2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test",
						"oidc_provider.0.issuer", oidcIssuer2),
				),
			},
			{
				// The same issuer twice is refused at plan time, before any
				// request: selection matches on the issuer and takes the first
				// entry, so the second would be unreachable.
				Config: deploymentMultiOIDCConfig(oidcIssuer, oidcIssuer),
				ExpectError: regexp.MustCompile(
					`issuer .* is configured more than once`),
			},
			{
				// Removing every block clears both fields, which the API
				// accepts as two explicit nulls.
				Config: deploymentMultiOIDCConfig(),
				Check: resource.TestCheckResourceAttr("hush_deployment.test",
					"oidc_provider.#", "0"),
			},
		},
	})
}

// TestAccDataSourceDeploymentMultipleOIDCProviders verifies the data source
// surfaces every block.
func TestAccDataSourceDeploymentMultipleOIDCProviders(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentMultiOIDCConfig(oidcIssuer, oidcIssuer2) +
					deploymentDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hush_deployment.test",
						"oidc_provider.#", "2"),
					resource.TestCheckResourceAttr("data.hush_deployment.test",
						"oidc_provider.0.issuer", oidcIssuer),
					resource.TestCheckResourceAttr("data.hush_deployment.test",
						"oidc_provider.1.issuer", oidcIssuer2),
				),
			},
		},
	})
}

// deploymentMultiOIDCConfig renders one oidc_provider block per issuer, at a
// fixed deployment name so no step changes anything but the blocks.
func deploymentMultiOIDCConfig(issuers ...string) string {
	return deploymentOIDCBlocksConfig("test", "tf-acc-multi-oidc", issuers...)
}

// deploymentLegacyOIDCConfig renders the same shape at the name the create hook
// rewrites, so the deployment under test holds the API's singular field.
func deploymentLegacyOIDCConfig(issuers ...string) string {
	return deploymentOIDCBlocksConfig(
		"legacy", legacyOIDCDeploymentName, issuers...)
}

func deploymentOIDCBlocksConfig(
	label, name string, issuers ...string,
) string {
	blocks := ""
	for _, issuer := range issuers {
		blocks += fmt.Sprintf(`
  oidc_provider {
    issuer           = %q
    audience         = %q
    allowed_subjects = ["system:serviceaccount:hush-security:*"]
  }
`, issuer, oidcAudience)
	}

	return fmt.Sprintf(`
resource "hush_deployment" %q {
  name     = %q
  env_type = "dev"
  kind     = "k8s"
%s}
`, label, name, blocks)
}
