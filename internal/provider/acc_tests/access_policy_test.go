package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockAccessCredentialID = "acr-mock-1234"
const mockAccessPrivilegeID = "apr-mock-1234"

func TestAccResourceAccessPolicy_withEnvDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
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
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy"
  description          = "test policy description"
  enabled              = true
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-updated"
  description          = "updated policy description"
  enabled              = true
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_access_policy" "template" {
  name                 = "test-policy-template"
  description          = "policy with template delivery"
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-both"
  description          = "should fail with both delivery configs"
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-no-delivery"
  description          = "should fail without any delivery config"
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }
}
`
}

func TestAccResourceAccessPolicy_withVolumeDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
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
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-volume"
  description          = "test volume delivery policy"
  enabled              = true
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_access_policy" "test" {
  name                 = "test-policy-volume-updated"
  description          = "updated volume delivery policy"
  enabled              = true
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_access_policy" "template" {
  name                 = "test-policy-volume-template"
  description          = "policy with volume template delivery"
  access_credential_id = "` + mockAccessCredentialID + `"
  access_privilege_ids = ["` + mockAccessPrivilegeID + `"]
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred"
  description    = "AWS WIF credential for access policy test"
  deployment_ids = ["` + mockDeploymentID + `"]
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-aws-wif"
  description          = "test AWS WIF delivery policy"
  enabled              = true
  access_credential_id = hush_aws_wif_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `"]

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
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred"
  description    = "AWS WIF credential for access policy test"
  deployment_ids = ["` + mockDeploymentID + `", "` + mockDeploymentID2 + `"]
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-aws-wif-updated"
  description          = "updated AWS WIF delivery policy"
  enabled              = true
  access_credential_id = hush_aws_wif_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `", "` + mockDeploymentID2 + `"]

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
	return `
resource "hush_aws_wif_access_credential" "test" {
  name           = "test-aws-wif-cred-sa"
  description    = "AWS WIF credential for service account test"
  deployment_ids = ["` + mockDeploymentID + `"]
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-aws-wif-sa"
  description          = "AWS WIF delivery with service account"
  enabled              = true
  access_credential_id = hush_aws_wif_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `"]

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

func TestAccResourceAccessPolicy_withGcpWifDelivery(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyGcpWifDeliveryStep1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-gcp-wif",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "test GCP WIF delivery policy",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.subject_kind", "hush_subject",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.subject", "my-test-subject",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.service_account_token_lifetime", "3600",
					),
				),
			},
			{
				Config: accessPolicyGcpWifDeliveryStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "name", "test-policy-gcp-wif-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "description", "updated GCP WIF delivery policy",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.subject", "my-updated-subject",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.service_account_token_lifetime", "7200",
					),
				),
			},
		},
	})
}

func TestAccResourceAccessPolicy_withGcpWifDeliveryServiceAccount(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyGcpWifDeliveryServiceAccountStep(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.subject_kind", "service_account",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.service_account", "my-sa@my-project.iam.gserviceaccount.com",
					),
					resource.TestCheckResourceAttr(
						"hush_access_policy.test", "gcp_wif_delivery_config.0.service_account_token_lifetime", "7200",
					),
				),
			},
		},
	})
}

func TestAccDataSourceAccessPolicy_withGcpWifDelivery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("access_policy", "v1/access_policies"),
		Steps: []resource.TestStep{
			{
				Config: accessPolicyGcpWifDeliveryStep1() + accessPolicyGcpWifDeliveryDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_access_policy.test", "id", regexp.MustCompile("^apl-.+$"),
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "name", "test-policy-gcp-wif",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "gcp_wif_delivery_config.0.subject_kind", "hush_subject",
					),
					resource.TestCheckResourceAttr(
						"data.hush_access_policy.test", "gcp_wif_delivery_config.0.subject", "my-test-subject",
					),
					resource.TestCheckResourceAttrSet(
						"data.hush_access_policy.test", "status",
					),
				),
			},
		},
	})
}

func accessPolicyGcpWifDeliveryStep1() string {
	return `
resource "hush_gcp_wif_access_credential" "test" {
  name                 = "test-gcp-wif-cred-policy"
  description          = "GCP WIF credential for access policy test"
  deployment_ids       = ["` + mockDeploymentID + `"]
  project_number       = "123456789012"
  pool_id              = "my-wif-pool"
  workload_provider_id = "my-wif-provider"
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-gcp-wif"
  description          = "test GCP WIF delivery policy"
  enabled              = true
  access_credential_id = hush_gcp_wif_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  gcp_wif_delivery_config {
    subject_kind = "hush_subject"
    subject      = "my-test-subject"
  }
}
`
}

func accessPolicyGcpWifDeliveryStep2() string {
	return `
resource "hush_gcp_wif_access_credential" "test" {
  name                 = "test-gcp-wif-cred-policy"
  description          = "GCP WIF credential for access policy test"
  deployment_ids       = ["` + mockDeploymentID + `"]
  project_number       = "123456789012"
  pool_id              = "my-wif-pool"
  workload_provider_id = "my-wif-provider"
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-gcp-wif-updated"
  description          = "updated GCP WIF delivery policy"
  enabled              = true
  access_credential_id = hush_gcp_wif_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  gcp_wif_delivery_config {
    subject_kind                  = "hush_subject"
    subject                       = "my-updated-subject"
    service_account_token_lifetime = 7200
  }
}
`
}

func accessPolicyGcpWifDeliveryServiceAccountStep() string {
	return `
resource "hush_gcp_wif_access_credential" "test" {
  name                 = "test-gcp-wif-cred-sa"
  description          = "GCP WIF credential for service account test"
  deployment_ids       = ["` + mockDeploymentID + `"]
  project_number       = "123456789012"
  pool_id              = "my-wif-pool"
  workload_provider_id = "my-wif-provider"
}

resource "hush_access_policy" "test" {
  name                 = "test-policy-gcp-wif-sa"
  description          = "GCP WIF delivery with service account"
  enabled              = true
  access_credential_id = hush_gcp_wif_access_credential.test.id
  deployment_ids       = ["` + mockDeploymentID + `"]

  attestation_criteria {
    type  = "k8s:ns"
    value = "default"
  }

  attestation_criteria {
    type  = "k8s:sa"
    value = "app-service-account"
  }

  gcp_wif_delivery_config {
    subject_kind                  = "service_account"
    service_account               = "my-sa@my-project.iam.gserviceaccount.com"
    service_account_token_lifetime = 7200
  }
}
`
}

const accessPolicyGcpWifDeliveryDataSource = `
data "hush_access_policy" "test" {
  id = hush_access_policy.test.id
}
`
