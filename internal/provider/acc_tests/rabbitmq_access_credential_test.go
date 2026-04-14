package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const mockRabbitmqHost = "amqp://mock-rabbitmq.example.com:5672"
const mockRabbitmqUsername = "mock-rabbitmq-user"
const mockRabbitmqPassword = "mock-rabbitmq-password"

func TestAccResourceRabbitmqAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessCredentialStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_rabbitmq_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "name", "test-rabbitmq-cred",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "description", "test rabbitmq credential",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "port", "5672",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "management_port", "15672",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "vhost", "/",
					),
				),
			},
			{
				Config: rabbitmqAccessCredentialStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_rabbitmq_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "name", "test-rabbitmq-cred-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "description", "updated rabbitmq credential",
					),
					resource.TestCheckResourceAttr(
						"hush_rabbitmq_access_credential.test", "vhost", "/test",
					),
				),
			},
		},
	})
}

func TestAccDataSourceRabbitmqAccessCredential(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessCredentialStep1 + rabbitmqAccessCredentialDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_rabbitmq_access_credential.test", "id", regexp.MustCompile(`^acr-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_rabbitmq_access_credential.test", "name", "test-rabbitmq-cred",
					),
				),
			},
		},
	})
}

const rabbitmqAccessCredentialStep1 = `
resource "hush_rabbitmq_access_credential" "test" {
  name            = "test-rabbitmq-cred"
  description     = "test rabbitmq credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  host            = "` + mockRabbitmqHost + `"
  port            = 5672
  management_port = 15672
  username        = "` + mockRabbitmqUsername + `"
  password        = "` + mockRabbitmqPassword + `"
  vhost           = "/"
  tls             = false
}
`

const rabbitmqAccessCredentialStep2 = `
resource "hush_rabbitmq_access_credential" "test" {
  name            = "test-rabbitmq-cred-updated"
  description     = "updated rabbitmq credential"
  deployment_ids  = ["` + mockDeploymentID + `"]
  host            = "` + mockRabbitmqHost + `"
  port            = 5672
  management_port = 15672
  username        = "` + mockRabbitmqUsername + `"
  password        = "` + mockRabbitmqPassword + `"
  vhost           = "/test"
  tls             = false
}
`

const rabbitmqAccessCredentialDataSource = `
data "hush_rabbitmq_access_credential" "test" {
  id = hush_rabbitmq_access_credential.test.id
}
`
