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

func TestAccResourceAccessPolicy_withEnvDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyEnvDeliveryStep1(),
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
				Config: accessPolicyEnvDeliveryStep2(),
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

func TestAccResourceAccessPolicy_withEnvDeliveryTemplate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyEnvDeliveryTemplateStep(),
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

func TestAccDataSourceAccessPolicy_withEnvDelivery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyEnvDeliveryStep1() + accessPolicyEnvDeliveryDataSource,
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

func accessPolicyEnvDeliveryStep1() string {
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

func accessPolicyEnvDeliveryStep2() string {
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

func accessPolicyEnvDeliveryTemplateStep() string {
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

const accessPolicyEnvDeliveryDataSource = `
data "hush_access_policy" "test" {
  id = hush_access_policy.test.id
}
`

func TestAccResourceAccessPolicy_withBothDeliveryConfigs(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      accessPolicyBothDeliveryConfigs(),
				ExpectError: regexp.MustCompile(`"env_delivery_config": only one of`),
			},
		},
	})
}

func accessPolicyBothDeliveryConfigs() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-both"
  description          = "should fail with both delivery configs"
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

  volume_delivery_config {
    mount_point = "/etc/secrets"

    item {
      path = "db_password"
      key  = "password"
      type = "key"
    }
  }
}
`
}

func TestAccResourceAccessPolicy_withNoDeliveryConfig(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      accessPolicyNoDeliveryConfig(),
				ExpectError: regexp.MustCompile(`"env_delivery_config": one of`),
			},
		},
	})
}

func accessPolicyNoDeliveryConfig() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-no-delivery"
  description          = "should fail without any delivery config"
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }
}
`
}

func TestAccResourceAccessPolicy_withVolumeDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyVolumeDeliveryStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-volume",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "test volume delivery policy",
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
						"hush_access_policy.test", "volume_delivery_config.0.mount_point", "/etc/secrets",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "volume_delivery_config.0.item.0.path", "db_password",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "volume_delivery_config.0.item.0.key", "password",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "volume_delivery_config.0.item.0.type", "key",
					),
				),
			},
			{
				Config: accessPolicyVolumeDeliveryStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-volume-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "updated volume delivery policy",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "volume_delivery_config.0.mount_point", "/var/secrets",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "volume_delivery_config.0.item.0.path", "api_key",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "volume_delivery_config.0.item.0.key", "password",
					),
				),
			},
		},
	})
}

func TestAccResourceAccessPolicy_withVolumeDeliveryTemplate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyVolumeDeliveryTemplateStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.template", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "volume_delivery_config.0.mount_point", "/etc/secrets",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "volume_delivery_config.0.item.0.path", "db_config.json",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "volume_delivery_config.0.item.0.type", "template",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "volume_delivery_config.0.item.0.key", "postgresql://${username}:${password}@host:5432/db",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "volume_delivery_config.0.item.1.path", "port",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.template", "volume_delivery_config.0.item.1.type", "key",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAccessPolicy_withVolumeDelivery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyVolumeDeliveryStep1() + accessPolicyVolumeDeliveryDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "name", "test-policy-volume",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "description", "test volume delivery policy",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "attestation_criteria.0.type", "k8s:ns",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "volume_delivery_config.0.mount_point", "/etc/secrets",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "volume_delivery_config.0.item.0.path", "db_password",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_access_policy.test", "status",
					),
				),
			},
		},
	})
}

func accessPolicyVolumeDeliveryStep1() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-volume"
  description          = "test volume delivery policy"
  enabled              = true
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  volume_delivery_config {
    mount_point = "/etc/secrets"

    item {
      path = "db_password"
      key  = "password"
      type = "key"
    }
  }
}
`
}

func accessPolicyVolumeDeliveryStep2() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-volume-updated"
  description          = "updated volume delivery policy"
  enabled              = true
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  volume_delivery_config {
    mount_point = "/var/secrets"

    item {
      path = "api_key"
      key  = "password"
      type = "key"
    }
  }
}
`
}

func accessPolicyVolumeDeliveryTemplateStep() string {
	credID := os.Getenv(envHushTestAccessCredentialID)
	privID := os.Getenv(envHushTestAccessPrivilegeID)
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_access_policy" "template" {
  name                 = "test-policy-volume-template"
  description          = "policy with volume template delivery"
  access_credential_id = "` + credID + `"
  access_privilege_ids = ["` + privID + `"]
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  volume_delivery_config {
    mount_point = "/etc/secrets"

    item {
      path = "db_config.json"
      key  = "postgresql://$${username}:$${password}@host:5432/db"
      type = "template"
    }

    item {
      path = "port"
      key  = "port"
      type = "key"
    }
  }
}
`
}

const accessPolicyVolumeDeliveryDataSource = `
data "hush_access_policy" "test" {
  id = hush_access_policy.test.id
}
`

func TestAccResourceAccessPolicy_withAwsWifDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyAwsWifDeliveryStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-aws-wif",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "test AWS WIF delivery policy",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.role_arn", "arn:aws:iam::123456789012:role/test-role",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.subject_kind", "hush_subject",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.subject", "my-test-subject",
					),
				),
			},
			{
				Config: accessPolicyAwsWifDeliveryStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-aws-wif-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "updated AWS WIF delivery policy",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "deployment_ids.#", "2",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.role_arn", "arn:aws:iam::123456789012:role/updated-role",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.subject", "my-updated-subject",
					),
				),
			},
		},
	})
}

func TestAccResourceAccessPolicy_withAwsWifDeliveryServiceAccount(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyAwsWifDeliveryServiceAccountStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.role_arn", "arn:aws:iam::123456789012:role/sa-role",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "aws_wif_delivery_config.0.subject_kind", "service_account",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAccessPolicy_withAwsWifDelivery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAccessPolicyPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyAwsWifDeliveryStep1() + accessPolicyAwsWifDeliveryDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "name", "test-policy-aws-wif",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "aws_wif_delivery_config.0.role_arn", "arn:aws:iam::123456789012:role/test-role",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "aws_wif_delivery_config.0.subject_kind", "hush_subject",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_access_policy.test", "status",
					),
				),
			},
		},
	})
}

func accessPolicyAwsWifDeliveryStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred"
  description    = "AWS WIF credential for access policy test"
  deployment_ids = ["` + deploymentID + `"]
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-aws-wif"
  description          = "test AWS WIF delivery policy"
  enabled              = true
  access_credential_id = hush_aws_wif_access_credential.test.id
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  aws_wif_delivery_config {
    role_arn     = "arn:aws:iam::123456789012:role/test-role"
    subject_kind = "hush_subject"
    subject      = "my-test-subject"
  }
}
`
}

func accessPolicyAwsWifDeliveryStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	deploymentID2 := os.Getenv(envHushTestDeploymentID2)
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred"
  description    = "AWS WIF credential for access policy test"
  deployment_ids = ["` + deploymentID + `", "` + deploymentID2 + `"]
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-aws-wif-updated"
  description          = "updated AWS WIF delivery policy"
  enabled              = true
  access_credential_id = hush_aws_wif_access_credential.test.id
  deployment_ids       = ["` + deploymentID + `", "` + deploymentID2 + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  aws_wif_delivery_config {
    role_arn     = "arn:aws:iam::123456789012:role/updated-role"
    subject_kind = "hush_subject"
    subject      = "my-updated-subject"
  }
}
`
}

func accessPolicyAwsWifDeliveryServiceAccountStep() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred-sa"
  description    = "AWS WIF credential for service account test"
  deployment_ids = ["` + deploymentID + `"]
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-aws-wif-sa"
  description          = "AWS WIF delivery with service account"
  enabled              = true
  access_credential_id = hush_aws_wif_access_credential.test.id
  deployment_ids       = ["` + deploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  attestation_criteria {
    type  = "k8s:sa"
    value = "app-service-account"
  }

  aws_wif_delivery_config {
    role_arn     = "arn:aws:iam::123456789012:role/sa-role"
    subject_kind = "service_account"
  }
}
`
}

const accessPolicyAwsWifDeliveryDataSource = `
data "hush_access_policy" "test" {
  id = hush_access_policy.test.id
}
`
