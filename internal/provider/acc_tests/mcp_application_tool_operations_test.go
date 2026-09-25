package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func toolOpsAppConfig(body string) string {
	return `
resource "hush_mcp_application" "tools" {
  app_catalog_id = "datadog"
  display_name   = "acc-tools"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
` + body + `}

resource "terraform_data" "operations" {
  input = hush_mcp_application.tools.tools[*].operation
}
`
}

// The fake catalog's datadog tools are search (read), create (write) and
// drop (destructive).
//
// A terraform_data reads the tools' operations: were tools left known across
// an edit that changes them, it would be planned with the previous ones, and
// the apply would fail as an inconsistent final plan.
func TestAccResourceMCPApplication_toolOperations(t *testing.T) {
	const addr = "hush_mcp_application.tools"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-tools"),
		Steps: []resource.TestStep{
			// Nothing written: the class defaults show, and no tool has its own.
			{
				Config: toolOpsAppConfig(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "tool_defaults.0.read", "allow"),
					resource.TestCheckResourceAttr(addr, "tool_defaults.0.write", "user_consent"),
					resource.TestCheckResourceAttr(addr, "tool_defaults.0.destructive", "block"),
					resource.TestCheckResourceAttr(addr, "tool_operation.#", "0"),
				),
			},
			// One class changed, the others left to what the application has.
			{
				Config: toolOpsAppConfig(`
  tool_defaults {
    write = "block"
  }
  tool_operation {
    name      = "drop"
    operation = "user_consent"
  }
  tool_operation {
    name      = "search"
    operation = "block"
  }
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "tool_defaults.0.read", "allow"),
					resource.TestCheckResourceAttr(addr, "tool_defaults.0.write", "block"),
					resource.TestCheckResourceAttr(addr, "tool_operation.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(addr, "tool_operation.*", map[string]string{
						"name": "drop", "operation": "user_consent",
					}),
					// tools shows each tool's resolved operation.
					resource.TestCheckResourceAttr(addr, "tools.0.operation", "block"),
					resource.TestCheckResourceAttr(addr, "tools.1.operation", "block"),
					resource.TestCheckResourceAttr(addr, "tools.2.operation", "user_consent"),
					resource.TestCheckResourceAttr("terraform_data.operations", "output.1", "block"),
				),
			},
			{
				Config: toolOpsAppConfig(`
  tool_defaults {
    write = "block"
  }
  tool_operation {
    name      = "drop"
    operation = "user_consent"
  }
  tool_operation {
    name      = "search"
    operation = "block"
  }
`),
				PlanOnly: true,
			},
			// Dropping a block returns the tool to its class; dropping the
			// defaults block leaves the classes as they are.
			{
				Config: toolOpsAppConfig(`
  tool_operation {
    name      = "drop"
    operation = "allow"
  }
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "tool_defaults.0.write", "block"),
					resource.TestCheckResourceAttr(addr, "tool_operation.#", "1"),
					resource.TestCheckResourceAttr(addr, "tools.0.operation", "allow"),
					resource.TestCheckResourceAttr(addr, "tools.2.operation", "allow"),
				),
			},
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"url_label"},
			},
		},
	})
}

func TestAccResourceMCPApplication_toolOperationChecks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-tools-bad"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_mcp_application" "bad" {
  app_catalog_id = "datadog"
  display_name   = "acc-tools-bad"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
  tool_operation {
    name      = "search"
    operation = "allow"
  }
  tool_operation {
    name      = "search"
    operation = "block"
  }
}
`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`tool_operation: tool "search" has more than one block`),
			},
			{
				Config: `
resource "hush_mcp_application" "bad" {
  app_catalog_id = "datadog"
  display_name   = "acc-tools-bad"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
  tool_operation {
    name      = "no_such_tool"
    operation = "allow"
  }
}
`,
				// On create the entry is read, so a typo is refused at plan
				// time, before anything exists to be tainted.
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`tool_operation: "datadog" has no tool "no_such_tool"`),
			},
			{
				Config: `
resource "hush_mcp_application" "bad" {
  app_catalog_id = "datadog"
  display_name   = "acc-tools-bad"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`,
			},
			// An update reads no entry, so the refusal is heimdall's.
			{
				Config: `
resource "hush_mcp_application" "bad" {
  app_catalog_id = "datadog"
  display_name   = "acc-tools-bad"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
  tool_operation {
    name      = "no_such_tool"
    operation = "allow"
  }
}
`,
				ExpectError: regexp.MustCompile(`tool_operation: the application has no such tool`),
			},
		},
	})
}
