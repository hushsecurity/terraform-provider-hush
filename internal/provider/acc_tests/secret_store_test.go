package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSecretStore(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("secret_store", "v1/secret_stores"),
		Steps: []resource.TestStep{
			{
				Config: secretStoreStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_secret_store.test", "id", regexp.MustCompile(`^sst-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "name", "test-secret-store",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "aws_sm.0.prefix", "hush",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "aws_sm.0.region", "eu-west-1",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "deployment_ids.0", mockDeploymentID,
					),
				),
			},
			{
				Config: secretStoreStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "name", "test-secret-store-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "description", "updated description",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.test", "aws_sm.0.region", "eu-west-1",
					),
				),
			},
			{
				ResourceName:      "hush_secret_store.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecretStoreK8s(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("secret_store", "v1/secret_stores"),
		Steps: []resource.TestStep{
			{
				Config: secretStoreK8sStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_secret_store.k8s", "id", regexp.MustCompile(`^sst-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.k8s", "k8s_secrets.0.prefix", "hush",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.k8s", "k8s_secrets.0.namespace", "hush-secrets",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSecretStore(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: secretStoreDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_secret_store.test", "id", regexp.MustCompile(`^sst-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_secret_store.test", "name", "test-secret-store-ds",
					),
					resource.TestCheckResourceAttr(
						"data.hush_secret_store.test", "gcp_sm.0.project_id", "my-project",
					),
				),
			},
		},
	})
}

const secretStoreStep1 = `
resource "hush_secret_store" "test" {
  name           = "test-secret-store"
  deployment_ids = ["` + mockDeploymentID + `"]

  aws_sm {
    prefix = "hush"
    region = "eu-west-1"
  }
}
`

const secretStoreStep2 = `
resource "hush_secret_store" "test" {
  name           = "test-secret-store-updated"
  description    = "updated description"
  deployment_ids = ["` + mockDeploymentID + `"]

  aws_sm {
    prefix = "hush"
    region = "eu-west-1"
  }
}
`

const secretStoreK8sStep1 = `
resource "hush_secret_store" "k8s" {
  name = "test-secret-store-k8s"

  k8s_secrets {
    prefix    = "hush"
    namespace = "hush-secrets"
  }
}
`

const secretStoreDataSource = `
resource "hush_secret_store" "ds_source" {
  name = "test-secret-store-ds"

  gcp_sm {
    prefix     = "hush"
    project_id = "my-project"
  }
}

data "hush_secret_store" "test" {
  name = hush_secret_store.ds_source.name
}
`

func TestAccResourceSecretStoreAwsSSM(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("secret_store", "v1/secret_stores"),
		Steps: []resource.TestStep{
			{
				Config: secretStoreAwsSSMStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_secret_store.ssm", "id", regexp.MustCompile(`^sst-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.ssm", "aws_ssm.0.prefix", "hush",
					),
					resource.TestCheckResourceAttr(
						"hush_secret_store.ssm", "aws_ssm.0.region", "us-east-1",
					),
				),
			},
		},
	})
}

const secretStoreAwsSSMStep1 = `
resource "hush_secret_store" "ssm" {
  name = "test-secret-store-ssm"

  aws_ssm {
    prefix = "hush"
    region = "us-east-1"
  }
}
`
