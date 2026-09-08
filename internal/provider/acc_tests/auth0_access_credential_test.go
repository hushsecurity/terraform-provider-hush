package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAuth0AccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("auth0_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: auth0AccessCredentialStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_auth0_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "name", "test-auth0-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "domain", "acme.us.auth0.com",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "client_id", "mgmtClientId",
					),
					checkSecretStoreID("hush_auth0_access_credential.test"),
				),
			},
			{
				Config: auth0AccessCredentialStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "name", "test-auth0-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "description", "updated auth0 credential",
					),
				),
			},
		},
	})
}

// Removing custom_domain must clear it: midgard types the field
// DomainStr | None, so an empty string is a 422.
func TestAccResourceAuth0AccessCredential_CustomDomainSetThenUnset(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("auth0_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: auth0AccessCredentialCustomDomain(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "domain", "acme.us.auth0.com",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "custom_domain", "auth.acme.com",
					),
				),
			},
			{
				// Same resource, custom_domain removed from the config.
				Config: auth0AccessCredentialCustomDomainRemoved(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "domain", "acme.us.auth0.com",
					),
					resource.TestCheckResourceAttr(
						"hush_auth0_access_credential.test", "custom_domain", "",
					),
				),
			},
		},
	})
}

// A custom domain in `domain` must be refused at plan time, not at apply.
func TestAccResourceAuth0AccessCredential_RejectsCustomDomainAsDomain(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      auth0AccessCredentialBadDomain(),
				ExpectError: regexp.MustCompile(`canonical Auth0 domain`),
				PlanOnly:    true,
			},
		},
	})
}

// Bumping the version must converge with no perpetual diff.
func TestAccResourceAuth0AccessCredential_WOSecretRotation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("auth0_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: auth0AccessCredentialWOSecretStep1(),
				Check: resource.TestCheckResourceAttr(
					"hush_auth0_access_credential.test", "client_secret_wo_version", "1",
				),
			},
			{
				Config: auth0AccessCredentialWOSecretStep2(),
				Check: resource.TestCheckResourceAttr(
					"hush_auth0_access_credential.test", "client_secret_wo_version", "2",
				),
			},
		},
	})
}

// deployment_ids is immutable (credutil.ForbidDeploymentIDsChange).
func TestAccResourceAuth0AccessCredential_DeploymentIDsImmutable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("auth0_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: auth0AccessCredentialStep1(),
			},
			{
				Config:      auth0AccessCredentialDeploymentChanged(),
				ExpectError: regexp.MustCompile(`deployment_ids cannot be changed after creation`),
			},
		},
	})
}

func auth0AccessCredentialStep1() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name            = "test-auth0-cred"
  description     = "test auth0 credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  domain          = "acme.us.auth0.com"
  client_id       = "mgmtClientId"
  client_secret   = "test-auth0-client-secret"
}
`
}

func auth0AccessCredentialStep2() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name            = "test-auth0-cred-updated"
  description     = "updated auth0 credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  secret_store_id = "sst-mock-store-1"
  domain          = "acme.us.auth0.com"
  client_id       = "mgmtClientId"
  client_secret   = "test-auth0-client-secret"
}
`
}

func auth0AccessCredentialCustomDomain() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name           = "test-auth0-cred-custom"
  deployment_ids = ["` + mockDeploymentID + `"]
  domain         = "acme.us.auth0.com"
  custom_domain  = "auth.acme.com"
  client_id      = "mgmtClientId"
  client_secret  = "test-auth0-client-secret"
}
`
}

func auth0AccessCredentialCustomDomainRemoved() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name           = "test-auth0-cred-custom"
  deployment_ids = ["` + mockDeploymentID + `"]
  domain         = "acme.us.auth0.com"
  client_id      = "mgmtClientId"
  client_secret  = "test-auth0-client-secret"
}
`
}

func auth0AccessCredentialBadDomain() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name           = "test-auth0-cred-bad-domain"
  deployment_ids = ["` + mockDeploymentID + `"]
  domain         = "auth.acme.com"
  client_id      = "mgmtClientId"
  client_secret  = "test-auth0-client-secret"
}
`
}

func auth0AccessCredentialWOSecretStep1() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name                     = "test-auth0-wo-secret"
  deployment_ids           = ["` + mockDeploymentID + `"]
  domain                   = "acme.us.auth0.com"
  client_id                = "mgmtClientId"
  client_secret_wo         = "secret-v1"
  client_secret_wo_version = "1"
}
`
}

func auth0AccessCredentialWOSecretStep2() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name                     = "test-auth0-wo-secret"
  deployment_ids           = ["` + mockDeploymentID + `"]
  domain                   = "acme.us.auth0.com"
  client_id                = "mgmtClientId"
  client_secret_wo         = "secret-v2"
  client_secret_wo_version = "2"
}
`
}

func auth0AccessCredentialDeploymentChanged() string {
	return `
resource "hush_auth0_access_credential" "test" {
  name            = "test-auth0-cred"
  description     = "test auth0 credential"
  deployment_ids  = ["` + mockDeploymentID2 + `"]
  secret_store_id = "sst-mock-store-1"
  domain          = "acme.us.auth0.com"
  client_id       = "mgmtClientId"
  client_secret   = "test-auth0-client-secret"
}
`
}
