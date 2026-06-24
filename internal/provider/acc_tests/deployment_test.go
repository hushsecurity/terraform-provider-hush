package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDeployment(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "name", "test-deployment",
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "description", "test deployment description",
					),
				),
			},
			{
				Config: deploymentStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "name", "test-deployment-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_deployment.test", "description", "updated deployment description",
					),
				),
			},
		},
	})
}

func TestAccDataSourceDeployment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentStep1 + deploymentDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_deployment.test", "id", regexp.MustCompile("^dep-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_deployment.test", "name", "test-deployment",
					),
					resource.TestCheckResourceAttr(
						"data.hush_deployment.test", "description", "test deployment description",
					),
				),
			},
		},
	})
}

const (
	deploymentStep1 = `
resource "hush_deployment" "test" {
  name        = "test-deployment"
  description = "test deployment description"
  kind        = "k8s"
}
`

	deploymentStep2 = `
resource "hush_deployment" "test" {
  name        = "test-deployment-updated"
  description = "updated deployment description"
  kind        = "k8s"
}
`

	deploymentDataSource = `
data "hush_deployment" "test" {
  id = hush_deployment.test.id
}
`
)

const (
	oidcIssuer    = "https://oidc.eks.us-east-1.amazonaws.com/id/D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9"
	oidcAudience  = "https://kubernetes.default.svc"
	oidcIssuer2   = "https://oidc.eks.eu-west-1.amazonaws.com/id/A1B2C3D4E5F60718293A4B5C6D7E8F90"
	oidcAudience2 = "sts.amazonaws.com"

	// deploymentNoOIDC is the same deployment without an oidc_provider block,
	// used to verify removal.
	deploymentNoOIDC = `
resource "hush_deployment" "test" {
  name = "test-deployment-oidc"
  kind = "k8s"
}
`
)

// deploymentOIDCConfig renders a deployment with an oidc_provider block.
// allowedSubjects is injected verbatim as an HCL list literal.
func deploymentOIDCConfig(issuer, audience, allowedSubjects string) string {
	return fmt.Sprintf(`
resource "hush_deployment" "test" {
  name = "test-deployment-oidc"
  kind = "k8s"

  oidc_provider {
    issuer           = %q
    audience         = %q
    allowed_subjects = %s
  }
}
`, issuer, audience, allowedSubjects)
}

// TestAccResourceDeploymentOIDC exercises the full oidc_provider lifecycle:
// create with the block, widen allowed_subjects, change issuer + audience, and
// remove the block (explicit null to the API).
func TestAccResourceDeploymentOIDC(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentOIDCConfig(oidcIssuer, oidcAudience, `["system:serviceaccount:hush-security:*"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.issuer", oidcIssuer),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.audience", oidcAudience),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.allowed_subjects.#", "1"),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.allowed_subjects.0", "system:serviceaccount:hush-security:*"),
				),
			},
			{
				Config: deploymentOIDCConfig(oidcIssuer, oidcAudience, `["system:serviceaccount:hush-security:*", "system:serviceaccount:other:*"]`),
				Check:  resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.allowed_subjects.#", "2"),
			},
			{
				Config: deploymentOIDCConfig(oidcIssuer2, oidcAudience2, `["system:serviceaccount:hush-security:*", "system:serviceaccount:other:*"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.issuer", oidcIssuer2),
					resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.0.audience", oidcAudience2),
				),
			},
			{
				Config: deploymentNoOIDC,
				Check:  resource.TestCheckResourceAttr("hush_deployment.test", "oidc_provider.#", "0"),
			},
		},
	})
}

// TestAccDataSourceDeploymentOIDC verifies the data source surfaces oidc_provider.
func TestAccDataSourceDeploymentOIDC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("deployment", "v1/deployments"),
		Steps: []resource.TestStep{
			{
				Config: deploymentOIDCConfig(oidcIssuer, oidcAudience, `["system:serviceaccount:hush-security:*"]`) + deploymentDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hush_deployment.test", "oidc_provider.#", "1"),
					resource.TestCheckResourceAttr("data.hush_deployment.test", "oidc_provider.0.issuer", oidcIssuer),
					resource.TestCheckResourceAttr("data.hush_deployment.test", "oidc_provider.0.audience", oidcAudience),
				),
			},
		},
	})
}
