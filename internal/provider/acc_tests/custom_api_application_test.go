package acc_tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func customAPIAppDestroyed(name string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		if _, ok := heimdall.customApps[name]; ok {
			return fmt.Errorf("custom app %s still exists", name)
		}
		return nil
	}
}

const customAPIAppStep1 = `
resource "hush_custom_api_application" "test" {
  name         = "acc-api"
  display_name = "Acc API"
  description  = "first"
  url          = "https://api.example.com/v1"
  spec_url     = "https://api.example.com/openapi.json"
  headers      = { "X-Tenant" = "acme" }

  auth {
    type              = "header"
    name              = "X-Api-Key"
    secret_wo         = "key-1"
    secret_wo_version = "1"
  }

  tool {
    name   = "list_items"
    title  = "List items"
    type   = "read"
    method = "GET"
    path   = "/items"
  }
  tool {
    name        = "get_item"
    description = "Fetch one item"
    type        = "read"
    method      = "GET"
    path        = "/items/{id}"
    input_schema = jsonencode({
      type = "object"
      properties = {
        path = { type = "object", properties = { id = { type = "string" } } }
      }
    })
  }
}
`

// The url moves while spec_url stays, so the tools want a new detect. One
// tool is disabled and its schema rewritten in another key order.
const customAPIAppStep2 = `
resource "hush_custom_api_application" "test" {
  name         = "acc-api"
  display_name = "Acc API renamed"
  url          = "https://api2.example.com/v1"
  spec_url     = "https://api.example.com/openapi.json"

  auth {
    type   = "bearer"
    secret = "tok-1"
  }

  tool {
    name    = "list_items"
    title   = "List items"
    type    = "read"
    method  = "GET"
    path    = "/items"
    enabled = false
  }
  tool {
    name         = "get_item"
    description  = "Fetch one item"
    type         = "read"
    method       = "GET"
    path         = "/items/{id}"
    input_schema = "{\"properties\": {\"path\": {\"properties\": {\"id\": {\"type\": \"string\"}}, \"type\": \"object\"}}, \"type\": \"object\"}"
  }
}
`

func TestAccResourceCustomAPIApplication(t *testing.T) {
	const addr = "hush_custom_api_application.test"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      customAPIAppDestroyed("acc-api"),
		Steps: []resource.TestStep{
			{
				Config: customAPIAppStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "acc-api"),
					resource.TestCheckResourceAttr(addr, "app_catalog_id", "custom-acc-api"),
					resource.TestCheckResourceAttr(addr, "url", "https://api.example.com/v1"),
					resource.TestCheckResourceAttr(addr, "detect_required", "false"),
					resource.TestCheckResourceAttr(addr, "headers.X-Tenant", "acme"),
					resource.TestCheckResourceAttr(addr, "auth.0.name", "X-Api-Key"),
					resource.TestCheckResourceAttr(addr, "tool.#", "2"),
					resource.TestCheckResourceAttr(addr, "tool.0.enabled", "true"),
					resource.TestCheckResourceAttr(addr, "tool.0.input_schema", `{"type":"object"}`),
					resource.TestCheckResourceAttr(addr, "tool.0.content_type", "application/json"),
					resource.TestCheckResourceAttr(addr, "tool.1.operation_id", ""),
					customAppSent("acc-api", "auth", "key-1"),
				),
			},
			{Config: customAPIAppStep1, PlanOnly: true},
			{
				Config: customAPIAppStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "display_name", "Acc API renamed"),
					resource.TestCheckResourceAttr(addr, "description", ""),
					resource.TestCheckResourceAttr(addr, "headers.%", "0"),
					resource.TestCheckResourceAttr(addr, "detect_required", "true"),
					resource.TestCheckResourceAttr(addr, "auth.0.type", "bearer"),
					resource.TestCheckResourceAttr(addr, "tool.0.enabled", "false"),
					customAppSent("acc-api", "auth", "tok-1"),
				),
			},
			{Config: customAPIAppStep2, PlanOnly: true},
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth.0.secret"},
			},
		},
	})
}

// Tools detected in the console are left alone while no tool block is
// written, and an import reads them, operation_id included, to copy into the
// configuration.
func TestAccResourceCustomAPIApplication_detectedTools(t *testing.T) {
	const addr = "hush_custom_api_application.detected"
	config := `
resource "hush_custom_api_application" "detected" {
  name         = "acc-api-detected"
  display_name = "Acc API detected"
  url          = "https://api.example.com"
  spec_url     = "https://api.example.com/openapi.json"
}
`
	written := `
resource "hush_custom_api_application" "detected" {
  name         = "acc-api-detected"
  display_name = "Acc API detected"
  url          = "https://api.example.com"
  spec_url     = "https://api.example.com/openapi.json"

  tool {
    name         = "list_pets"
    title        = "List pets"
    type         = "read"
    method       = "GET"
    path         = "/pets"
    operation_id = "listPets"
  }
}
`
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      customAPIAppDestroyed("acc-api-detected"),
		Steps: []resource.TestStep{
			{Config: config, Check: resource.TestCheckResourceAttr(addr, "tool.#", "0")},
			{
				PreConfig: func() {
					seedDetectedAPITools("acc-api-detected", map[string]any{
						"name": "listPets", "title": "", "description": "List all pets", "type": "read",
						"enabled": true, "method": "GET", "path": "/pets",
						"input_schema": map[string]any{"type": "object"},
						"content_type": "application/json", "operation_id": "listPets",
					})
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "tool.#", "1"),
					resource.TestCheckResourceAttr(addr, "tool.0.operation_id", "listPets"),
				),
			},
			{ResourceName: addr, ImportState: true, ImportStateVerify: true},
			// Renamed by its operation_id.
			{
				Config: written,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "tool.0.name", "list_pets"),
					resource.TestCheckResourceAttr(addr, "tool.0.operation_id", "listPets"),
				),
			},
			{Config: written, PlanOnly: true},
		},
	})
}

func TestAccResourceCustomAPIApplication_planChecks(t *testing.T) {
	for name, tc := range map[string]struct{ body, expect string }{
		"input schema not json": {
			body:   `  tool {` + "\n" + `    name = "t"` + "\n" + `    type = "read"` + "\n" + `    method = "GET"` + "\n" + `    path = "/t"` + "\n" + `    input_schema = "{"` + "\n  }\n",
			expect: `"tool.0.input_schema" contains an invalid JSON`,
		},
		"unknown method": {
			body:   `  tool {` + "\n" + `    name = "t"` + "\n" + `    type = "read"` + "\n" + `    method = "FETCH"` + "\n" + `    path = "/t"` + "\n  }\n",
			expect: `expected tool.0.method to be one of`,
		},
		"path with a query": {
			body:   `  tool {` + "\n" + `    name = "t"` + "\n" + `    type = "read"` + "\n" + `    method = "GET"` + "\n" + `    path = "/t?x=1"` + "\n  }\n",
			expect: `must start with '/' and have no query`,
		},
		"operation_id on create": {
			body:   `  tool {` + "\n" + `    name = "t"` + "\n" + `    type = "read"` + "\n" + `    method = "GET"` + "\n" + `    path = "/t"` + "\n" + `    operation_id = "getT"` + "\n  }\n",
			expect: `operation_id "getT" can only name a tool already detected`,
		},
		"tool listed twice": {
			body:   apiTool("t", "") + apiTool("t", ""),
			expect: `tool\[1\]: name "t" is listed more than once`,
		},
		"title over 64 characters": {
			body:   apiTool("t", strings.Repeat("é", 65)),
			expect: `must be at most 64 characters`,
		},
		"reserved header": {
			body:   `  headers = { "Host" = "x" }` + "\n",
			expect: `"Host" is set by the gateway`,
		},
		"header auth duplicating a header": {
			body: `  headers = { "x-api-key" = "1" }
  auth {
    type   = "header"
    name   = "X-Api-Key"
    secret = "k"
  }
`,
			expect: `header "X-Api-Key" is already set in headers`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config: `
resource "hush_custom_api_application" "bad" {
  name         = "acc-api-bad"
  display_name = "Bad"
  url          = "https://api.example.com"
` + tc.body + "}\n",
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tc.expect),
				}},
			})
		})
	}
}

func TestAccDataSourceCustomAPIApplication(t *testing.T) {
	const config = `
resource "hush_custom_api_application" "src" {
  name         = "acc-api-ds"
  display_name = "Acc API ds"
  url          = "https://api.example.com"

  auth {
    type     = "basic"
    username = "bob"
    secret   = "pw"
  }
  tool {
    name   = "ping"
    type   = "read"
    method = "GET"
    path   = "/ping"
  }
}

data "hush_custom_api_application" "read" {
  name = hush_custom_api_application.src.name
}
`
	const addr = "data.hush_custom_api_application.read"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "acc-api-ds"),
					resource.TestCheckResourceAttr(addr, "app_catalog_id", "custom-acc-api-ds"),
					resource.TestCheckResourceAttr(addr, "url", "https://api.example.com"),
					resource.TestCheckResourceAttr(addr, "auth.0.username", "bob"),
					resource.TestCheckNoResourceAttr(addr, "auth.0.secret"),
					resource.TestCheckResourceAttr(addr, "tool.0.name", "ping"),
					resource.TestCheckResourceAttr(addr, "tool.0.method", "GET"),
				),
			},
			{
				Config: `
data "hush_custom_api_application" "none" {
  name = "acc-no-such-api"
}
`,
				ExpectError: regexp.MustCompile(`no custom API app found with name: acc-no-such-api`),
			},
		},
	})
}

func apiTool(name, title string) string {
	return "  tool {\n    name = \"" + name + "\"\n    title = \"" + title +
		"\"\n    type = \"read\"\n    method = \"GET\"\n    path = \"/t\"\n  }\n"
}

// heimdall's max_length counts characters, so 64 two-byte ones fit.
func TestAccResourceCustomAPIApplication_titleCharacters(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "hush_custom_api_application" "title" {
  name         = "acc-api-title"
  display_name = "Title"
  url          = "https://api.example.com"
` + apiTool("t", strings.Repeat("é", 64)) + "}\n",
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		}},
	})
}
