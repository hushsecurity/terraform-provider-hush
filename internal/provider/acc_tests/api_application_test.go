package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func apiAppConfig(tools, app string) string {
	return `
resource "hush_custom_api_application" "src" {
  name         = "acc-api-app"
  display_name = "Acc API app"
  url          = "https://api.example.com"
` + tools + `}

resource "hush_api_application" "test" {
  app_catalog_id = hush_custom_api_application.src.app_catalog_id
  deployment_ids = ["` + mockDeploymentID + `"]
` + app + `}
`
}

const apiAppTools = `
  tool {
    name   = "list_items"
    type   = "read"
    method = "GET"
    path   = "/items"
  }
  tool {
    name   = "delete_item"
    type   = "destructive"
    method = "DELETE"
    path   = "/items/{id}"
    input_schema = jsonencode({
      type       = "object"
      properties = { path = { type = "object", properties = { id = { type = "string" } } } }
    })
  }
`

// Every tool disabled: heimdall refuses it while an application uses the app.
const apiAppToolsDisabled = `
  tool {
    name    = "list_items"
    type    = "read"
    method  = "GET"
    path    = "/items"
    enabled = false
  }
`

func TestAccResourceAPIApplication(t *testing.T) {
	const addr = "hush_api_application.test"
	step1 := apiAppConfig(apiAppTools, `
  display_name = "acc-api-app"
  tool_defaults {
    destructive = "user_consent"
  }
  tool_operation {
    name      = "list_items"
    operation = "block"
  }
  assignment {
    condition {
      match {
        source   = "user"
        property = "email"
        op       = "eq"
        value    = "dev@example.com"
      }
    }
  }
`)
	step2 := apiAppConfig(apiAppTools, `
  display_name = "acc-api-app renamed"
  description  = "second"
  enabled      = false
  assign_all   = true
`)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-api-app", "acc-api-app renamed"),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", "custom-acc-api-app"),
					resource.TestCheckResourceAttr(addr, "url", "https://api.example.com"),
					resource.TestCheckResourceAttr(addr, "enabled", "true"),
					resource.TestCheckResourceAttr(addr, "tools.#", "2"),
					resource.TestCheckResourceAttr(addr, "tools.0.operation", "block"),
					resource.TestCheckResourceAttr(addr, "tools.1.operation", "user_consent"),
					resource.TestCheckResourceAttr(addr, "assignment.#", "1"),
				),
			},
			{Config: step1, PlanOnly: true},
			{
				Config: step2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "display_name", "acc-api-app renamed"),
					resource.TestCheckResourceAttr(addr, "enabled", "false"),
					resource.TestCheckResourceAttr(addr, "assign_all", "true"),
					resource.TestCheckResourceAttr(addr, "assignment.#", "0"),
					resource.TestCheckResourceAttr(addr, "tool_operation.#", "0"),
					resource.TestCheckResourceAttr(addr, "tools.0.operation", "allow"),
				),
			},
			{Config: step2, PlanOnly: true},
			{ResourceName: addr, ImportState: true, ImportStateVerify: true},
			{
				Config:      apiAppConfig(apiAppToolsDisabled, `  display_name = "acc-api-app renamed"`+"\n"),
				ExpectError: regexp.MustCompile(`applications use custom app acc-api-app, so at least one tool\s+must be enabled`),
			},
		},
	})
}

// A tool renamed on the custom app takes its operation with it, and the
// application's block for the old name goes in the same apply.
func TestAccResourceAPIApplication_toolRename(t *testing.T) {
	const addr = "hush_api_application.renamed"
	config := func(tools, name string) string {
		return `
resource "hush_custom_api_application" "src" {
  name         = "acc-api-rename"
  display_name = "Acc API rename"
  url          = "https://api.example.com"
` + tools + `}

resource "hush_api_application" "renamed" {
  app_catalog_id = hush_custom_api_application.src.app_catalog_id
  display_name   = "acc-api-rename"
  deployment_ids = ["` + mockDeploymentID + `"]
  tool_operation {
    name      = "` + name + `"
    operation = "block"
  }
}
`
	}
	renamed := `
  tool {
    name         = "list_pets"
    type         = "read"
    method       = "GET"
    path         = "/pets"
    operation_id = "listPets"
  }
`
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-api-rename"),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					seedDetectedAPITools("acc-api-rename", map[string]any{
						"name": "listPets", "title": "", "description": "", "type": "read",
						"enabled": true, "method": "GET", "path": "/pets",
						"input_schema": map[string]any{"type": "object"},
						"content_type": "application/json", "operation_id": "listPets",
					})
				},
				Config: config("", "listPets"),
				Check:  resource.TestCheckResourceAttr(addr, "tools.0.operation", "block"),
			},
			{
				Config: config(renamed, "list_pets"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "tools.0.name", "list_pets"),
					resource.TestCheckResourceAttr(addr, "tools.0.operation", "block"),
				),
			},
			{Config: config(renamed, "list_pets"), PlanOnly: true},
		},
	})
}
