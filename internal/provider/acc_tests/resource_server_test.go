package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceConfig(server, app string) string {
	return `
resource "hush_resource_server" "test" {
  name         = "acc-clientworks"
  audience     = "https://kong.example.com/clientworks"
` + server + `}

resource "hush_resource_application" "test" {
  resource       = hush_resource_server.test.name
  deployment_ids = ["` + mockDeploymentID + `"]
` + app + `}
`
}

func serversDestroyed(names ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		for _, name := range names {
			if _, ok := heimdall.servers[name]; ok {
				return fmt.Errorf("resource server %s still exists", name)
			}
		}
		return nil
	}
}

func TestAccResourceServerAndApplication(t *testing.T) {
	const server, app = "hush_resource_server.test", "hush_resource_application.test"
	step1 := resourceConfig(`
  display_name = "Clientworks"
  scope {
    name        = "accounts:read"
    description = "Read accounts"
  }
  scope {
    name = "trades:write"
  }
  token_claim {
    name      = "teams"
    attribute = "groups"
  }
`, `
  display_name = "Clientworks - Advisors"
  scopes       = ["accounts:read"]
  assignment {
    condition {
      match {
        source   = "user"
        property = "email"
        op       = "eq"
        value    = "dana@example.com"
      }
    }
  }
`)
	step2 := resourceConfig(`
  display_name   = "Clientworks API"
  description    = "second"
  token_lifetime = 900
  scope {
    name = "accounts:read"
  }
`, `
  display_name = "Clientworks - everyone"
  enabled      = false
  assign_all   = true
`)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			mcpAppsDestroyed("Clientworks - Advisors", "Clientworks - everyone"),
			serversDestroyed("acc-clientworks"),
		),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(server, "scope.#", "2"),
					resource.TestCheckResourceAttr(server, "scope.0.description", "Read accounts"),
					resource.TestCheckResourceAttr(server, "token_claim.0.attribute", "groups"),
					resource.TestCheckResourceAttr(server, "token_lifetime", "0"),
					resource.TestCheckResourceAttr(app, "resource", "acc-clientworks"),
					resource.TestCheckResourceAttr(app, "scopes.#", "1"),
					resource.TestCheckResourceAttr(app, "enabled", "true"),
					resource.TestCheckResourceAttr(app, "assignment.#", "1"),
				),
			},
			{
				Config: step2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(server, "display_name", "Clientworks API"),
					resource.TestCheckResourceAttr(server, "token_lifetime", "900"),
					resource.TestCheckResourceAttr(server, "scope.#", "1"),
					resource.TestCheckResourceAttr(server, "token_claim.#", "0"),
					resource.TestCheckResourceAttr(app, "scopes.#", "0"),
					resource.TestCheckResourceAttr(app, "enabled", "false"),
					resource.TestCheckResourceAttr(app, "assign_all", "true"),
					resource.TestCheckResourceAttr(app, "assignment.#", "0"),
				),
			},
			{ResourceName: server, ImportState: true, ImportStateVerify: true},
			{ResourceName: app, ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccResourceApplicationUnknownScope(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      serversDestroyed("acc-clientworks-2"),
		Steps: []resource.TestStep{{
			Config: `
resource "hush_resource_server" "test" {
  name         = "acc-clientworks-2"
  display_name = "Clientworks 2"
  audience     = "https://kong.example.com/clientworks-2"
  scope {
    name = "accounts:read"
  }
}

resource "hush_resource_application" "test" {
  resource       = hush_resource_server.test.name
  display_name   = "Clientworks 2 app"
  deployment_ids = ["` + mockDeploymentID + `"]
  scopes         = ["trades:write"]
}
`,
			ExpectError: regexp.MustCompile("no scope trades:write"),
		}},
	})
}
