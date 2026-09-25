package acc_tests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// customAppSent checks the secret heimdall last received for a custom app,
// which neither state nor any read can show.
func customAppSent(name, key string, want any) resource.TestCheckFunc {
	return func(*terraform.State) error {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		if got := heimdall.customSecrets[name][key]; got != want {
			return fmt.Errorf("%s %s: heimdall holds %v, want %v", name, key, got, want)
		}
		return nil
	}
}

func customAppDestroyed(*terraform.State) error {
	heimdall.mu.Lock()
	defer heimdall.mu.Unlock()
	for name := range heimdall.customApps {
		if name == "acc-custom" {
			return fmt.Errorf("custom app %s still exists", name)
		}
	}
	return nil
}

const customAppStep1 = `
resource "hush_custom_mcp_application" "test" {
  name         = "acc-custom"
  display_name = "Acc custom"
  description  = "first"

  url {
    label = "us"
    url   = "https://mcp.us.example.com/mcp"
  }
  url {
    label = "eu"
    url   = "https://mcp.eu.example.com/mcp"
  }

  scopes  = ["read"]
  headers = { "X-Tenant" = "acme" }

  tool {
    name        = "search"
    type        = "read"
    description = "Search things"
  }

  auth {
    type              = "header"
    name              = "X-Api-Key"
    secret_wo         = "key-1"
    secret_wo_version = "1"
  }
}
`

// No tool block: the tools stay as heimdall holds them. The header
// credential goes and OAuth client credentials come in.
const customAppStep2 = `
resource "hush_custom_mcp_application" "test" {
  name         = "acc-custom"
  display_name = "Acc custom renamed"

  url {
    url = "https://mcp.example.com/mcp"
  }

  client_id                = "cid-1"
  client_secret_wo         = "csecret-1"
  client_secret_wo_version = "1"
}
`

const customAppStep3 = `
resource "hush_custom_mcp_application" "test" {
  name         = "acc-custom"
  display_name = "Acc custom renamed"

  url {
    url = "https://mcp.example.com/mcp"
  }

  auth {
    type     = "basic"
    username = "bob"
    secret   = "pw-1"
  }
}
`

func TestAccResourceCustomMCPApplication(t *testing.T) {
	const addr = "hush_custom_mcp_application.test"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      customAppDestroyed,
		Steps: []resource.TestStep{
			{
				Config: customAppStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "acc-custom"),
					resource.TestCheckResourceAttr(addr, "app_catalog_id", "custom-acc-custom"),
					resource.TestCheckResourceAttr(addr, "url.#", "2"),
					resource.TestCheckResourceAttr(addr, "url.1.label", "eu"),
					resource.TestCheckResourceAttr(addr, "headers.X-Tenant", "acme"),
					resource.TestCheckResourceAttr(addr, "tool.0.name", "search"),
					resource.TestCheckResourceAttr(addr, "auth.0.name", "X-Api-Key"),
					resource.TestCheckResourceAttr(addr, "auth.0.secret_wo_version", "1"),
					customAppSent("acc-custom", "auth", "key-1"),
				),
			},
			{Config: customAppStep1, PlanOnly: true},
			{
				Config: customAppStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "display_name", "Acc custom renamed"),
					resource.TestCheckResourceAttr(addr, "description", ""),
					resource.TestCheckResourceAttr(addr, "url.#", "1"),
					resource.TestCheckResourceAttr(addr, "url.0.label", ""),
					resource.TestCheckResourceAttr(addr, "scopes.#", "0"),
					resource.TestCheckResourceAttr(addr, "headers.%", "0"),
					resource.TestCheckResourceAttr(addr, "tool.#", "1"),
					resource.TestCheckResourceAttr(addr, "auth.#", "0"),
					resource.TestCheckResourceAttr(addr, "client_id", "cid-1"),
					customAppSent("acc-custom", "client_secret", "csecret-1"),
					customAppSent("acc-custom", "auth", nil),
				),
			},
			{Config: customAppStep2, PlanOnly: true},
			// Dropping client_id drops its secret with it.
			{
				Config: customAppStep3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "client_id", ""),
					resource.TestCheckResourceAttr(addr, "auth.0.type", "basic"),
					resource.TestCheckResourceAttr(addr, "auth.0.username", "bob"),
					customAppSent("acc-custom", "client_secret", nil),
					customAppSent("acc-custom", "auth", "pw-1"),
				),
			},
			{Config: customAppStep3, PlanOnly: true},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
				// The API returns no secret, so neither it nor its version can come
				// back; the next apply re-sends it.
				ImportStateVerifyIgnore: []string{"auth.0.secret", "client_secret_wo_version"},
			},
		},
	})
}

// What heimdall refuses and the schema alone cannot, refused at plan time.
func TestAccResourceCustomMCPApplication_planChecks(t *testing.T) {
	const url = `
  url {
    url = "https://mcp.example.com/mcp"
  }
`
	for name, tc := range map[string]struct{ body, expect string }{
		"unlabelled among several urls": {
			body:   url + `  url {` + "\n" + `    label = "eu"` + "\n" + `    url = "https://mcp.eu.example.com/mcp"` + "\n  }\n",
			expect: `url\[0\]: a label is required when there is more than one url`,
		},
		"reserved header": {
			body:   url + `  headers = { "X-Hush-Thing" = "1" }` + "\n",
			expect: `"X-Hush-Thing" is reserved`,
		},
		"bearer next to oauth": {
			body: url + `  client_id = "cid"
  auth {
    type   = "bearer"
    secret = "t"
  }
`,
			expect: `type "bearer" sets the Authorization header`,
		},
		"auth without a secret": {
			body: url + `  auth {
    type = "bearer"
  }
`,
			expect: `auth: one of secret or secret_wo is required`,
		},
		"reserved auth header": {
			body: url + `  auth {
    type   = "header"
    name   = "X-Hush-Token"
    secret = "k"
  }
`,
			expect: `auth: "X-Hush-Token" is reserved`,
		},
		"basic without a username": {
			body: url + `  auth {
    type   = "basic"
    secret = "pw"
  }
`,
			expect: `auth: username is required with type "basic"`,
		},
		"header auth duplicating a header": {
			body: url + `  headers = { "x-api-key" = "1" }
  auth {
    type   = "header"
    name   = "X-Api-Key"
    secret = "k"
  }
`,
			expect: `header "X-Api-Key" is already set in headers`,
		},
		"client secret without a client id": {
			body:   url + `  client_secret = "s"` + "\n",
			expect: `all of .client_id,client_secret. must be specified`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config: `
resource "hush_custom_mcp_application" "bad" {
  name         = "acc-bad"
  display_name = "Bad"
` + tc.body + "}\n",
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tc.expect),
				}},
			})
		})
	}
}
