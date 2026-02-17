package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestAccessCredentialID = "HUSH_TEST_ACCESS_CREDENTIAL_ID"
const envHushTestAccessPrivilegeID = "HUSH_TEST_ACCESS_PRIVILEGE_ID"

func testAccAccessPolicyPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestAccessCredentialID) == "" {
		t.Fatalf("%s env var must be set", envHushTestAccessCredentialID)
	}
	if os.Getenv(envHushTestAccessPrivilegeID) == "" {
		t.Fatalf("%s env var must be set", envHushTestAccessPrivilegeID)
	}
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
}

func TestAccResourceAccessPolicy(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "test policy description",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "enabled", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "attestation_criteria.0.type", "k8s:ns",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "attestation_criteria.0.value", "default",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "env_delivery_config.0.key", "port",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "env_delivery_config.0.name", "PORT",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "env_delivery_config.0.type", "key",
					),
				),
			},
			{
				Config: accessPolicyStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "updated policy description",
					),
				),
			},
		},
	})
}

func TestAccResourceAccessPolicy_withTemplate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyTemplateStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.template", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "env_delivery_config.0.type", "template",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "env_delivery_config.0.key", "postgresql://${username}:${password}@host:5432/db",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "env_delivery_config.0.name", "DATABASE_URL",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAccessPolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyStep1() + accessPolicyDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "name", "test-policy",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "description", "test policy description",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "attestation_criteria.0.type", "k8s:ns",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "env_delivery_config.0.type", "key",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_access_policy.test", "status",
					),
				),
			},
		},
	})
}

func accessPolicyStep1() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy"
  description          = "test policy description"
  enabled              = true
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  env_delivery_config {
    key  = "port"
    name = "PORT"
    type = "key"
  }
}
`
}

func accessPolicyStep2() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-updated"
  description          = "updated policy description"
  enabled              = true
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  env_delivery_config {
    key  = "port"
    name = "PORT"
    type = "key"
  }
}
`
}

func accessPolicyTemplateStep() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "template" {
  name                 = "test-policy-template"
  description          = "policy with template delivery"
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  env_delivery_config {
    key  = "postgresql://$${username}:$${password}@host:5432/db"
    name = "DATABASE_URL"
    type = "template"
  }

  env_delivery_config {
    key  = "port"
    name = "PORT"
    type = "key"
  }
}
`
}

const accessPolicyDataSource = `
data "hush_access_policy" "test" {
  id = hush_access_policy.test.id
}
`
