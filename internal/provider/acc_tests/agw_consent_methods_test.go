package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const consentMethodsDeploymentID = "dep-consent-1"

func consentMethodsConfig(methods string) string {
	return fmt.Sprintf(`
resource "hush_agw_consent_methods" "test" {
  deployment_id = %q
  methods       = %s
}
`, consentMethodsDeploymentID, methods)
}

// Set, change, import, and destroy. Destroy has no API call behind it, so the
// methods stay stored after it: the check says so rather than expecting them
// gone.
func TestAccResourceAgwConsentMethods(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy: func(*terraform.State) error {
			heimdall.mu.Lock()
			defer heimdall.mu.Unlock()
			if _, ok := heimdall.consentMethods[consentMethodsDeploymentID]; !ok {
				return fmt.Errorf("destroy is expected to leave the methods stored, have %v", heimdall.consentMethods)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: consentMethodsConfig(`["oidc_callback", "push_slack"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_agw_consent_methods.test", "id", consentMethodsDeploymentID),
					resource.TestCheckResourceAttr("hush_agw_consent_methods.test", "methods.#", "2"),
					resource.TestCheckTypeSetElemAttr("hush_agw_consent_methods.test", "methods.*", "oidc_callback"),
					resource.TestCheckTypeSetElemAttr("hush_agw_consent_methods.test", "methods.*", "push_slack"),
				),
			},
			// Order carries no meaning, so reordering the list plans nothing.
			{
				Config:   consentMethodsConfig(`["push_slack", "oidc_callback"]`),
				PlanOnly: true,
			},
			// heimdall accepts mcp_elicitation, so the provider does too.
			{
				Config: consentMethodsConfig(`["mcp_elicitation"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_agw_consent_methods.test", "methods.#", "1"),
					resource.TestCheckTypeSetElemAttr("hush_agw_consent_methods.test", "methods.*", "mcp_elicitation"),
				),
			},
			{
				ResourceName:      "hush_agw_consent_methods.test",
				ImportState:       true,
				ImportStateId:     consentMethodsDeploymentID,
				ImportStateVerify: true,
			},
		},
	})
}

// A method Hush does not know is refused at plan time.
func TestAccResourceAgwConsentMethods_unknownMethod(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      consentMethodsConfig(`["email"]`),
				ExpectError: regexp.MustCompile(`expected methods\.\d+ to be one of`),
			},
		},
	})
}
