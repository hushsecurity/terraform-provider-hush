package acc_tests

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const envHushTestRabbitmqHost = "HUSH_TEST_RABBITMQ_HOST"
const envHushTestRabbitmqUsername = "HUSH_TEST_RABBITMQ_USERNAME"
const envHushTestRabbitmqPassword = "HUSH_TEST_RABBITMQ_PASSWORD"

func testAccRabbitmqAccessCredentialPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv(envHushTestDeploymentID) == "" {
		t.Fatalf("%s env var must be set", envHushTestDeploymentID)
	}
	if os.Getenv(envHushTestRabbitmqHost) == "" {
		t.Fatalf("%s env var must be set", envHushTestRabbitmqHost)
	}
	if os.Getenv(envHushTestRabbitmqUsername) == "" {
		t.Fatalf("%s env var must be set", envHushTestRabbitmqUsername)
	}
	if os.Getenv(envHushTestRabbitmqPassword) == "" {
		t.Fatalf("%s env var must be set", envHushTestRabbitmqPassword)
	}
}

func TestAccResourceRabbitmqAccessCredential(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccRabbitmqAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessCredentialStep1(),
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
				Config: rabbitmqAccessCredentialStep2(),
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
		PreCheck:          func() { testAccRabbitmqAccessCredentialPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("rabbitmq_access_credential", "v1/access_credentials"),
		Steps: []resource.TestStep{
			{
				Config: rabbitmqAccessCredentialStep1() + rabbitmqAccessCredentialDataSource,
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

func rabbitmqAccessCredentialStep1() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	host := os.Getenv(envHushTestRabbitmqHost)
	username := os.Getenv(envHushTestRabbitmqUsername)
	password := os.Getenv(envHushTestRabbitmqPassword)
	return `
resource "hush_rabbitmq_access_credential" "test" {
  name            = "test-rabbitmq-cred"
  description     = "test rabbitmq credential"
  deployment_ids  = ["` + deploymentID + `"]
  host            = "` + host + `"
  port            = 5672
  management_port = 15672
  username        = "` + username + `"
  password        = "` + password + `"
  vhost           = "/"
  tls             = false
}
`
}

func rabbitmqAccessCredentialStep2() string {
	deploymentID := os.Getenv(envHushTestDeploymentID)
	host := os.Getenv(envHushTestRabbitmqHost)
	username := os.Getenv(envHushTestRabbitmqUsername)
	password := os.Getenv(envHushTestRabbitmqPassword)
	return `
resource "hush_rabbitmq_access_credential" "test" {
  name            = "test-rabbitmq-cred-updated"
  description     = "updated rabbitmq credential"
  deployment_ids  = ["` + deploymentID + `"]
  host            = "` + host + `"
  port            = 5672
  management_port = 15672
  username        = "` + username + `"
  password        = "` + password + `"
  vhost           = "/test"
  tls             = false
}
`
}

const rabbitmqAccessCredentialDataSource = `
data "hush_rabbitmq_access_credential" "test" {
  id = hush_rabbitmq_access_credential.test.id
}
`
