package acc_tests

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func assignmentAppConfig(body string) string {
	return `
resource "hush_mcp_application" "assigned" {
  app_catalog_id = "datadog"
  display_name   = "acc-assigned"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
` + body + "}\n"
}

// mcpAppAssignments checks the assignments heimdall holds, as JSON.
func mcpAppAssignments(addr, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		got, _ := json.Marshal(heimdall.apps[s.RootModule().Resources[addr].Primary.ID]["assignments"])
		if string(got) != want {
			return fmt.Errorf("%s: heimdall holds assignments %s, want %s", addr, got, want)
		}
		return nil
	}
}

const twoRules = `
  assignment {
    condition {
      match {
        source   = "user"
        property = "groups"
        op       = "in"
        values   = ["grp-eng", "grp-sre"]
      }
      match {
        source   = "user"
        property = "email"
        op       = "eq"
        value    = "cto@example.com"
      }
    }
    condition {
      match {
        source   = "agent"
        property = "id"
        op       = "neq"
        value    = "agt-blocked"
      }
    }
  }
  assignment {
    condition {
      match {
        source   = "user"
        property = "department"
        op       = "eq"
        value    = "finance"
      }
    }
  }
`

func TestAccResourceMCPApplication_assignments(t *testing.T) {
	const addr = "hush_mcp_application.assigned"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-assigned"),
		Steps: []resource.TestStep{
			// Written at creation, which takes none: they are patched in.
			{
				Config: assignmentAppConfig(twoRules),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "assign_all", "false"),
					resource.TestCheckResourceAttr(addr, "assignment.#", "2"),
					resource.TestCheckResourceAttr(addr, "assignment.0.condition.0.match.0.values.1", "grp-sre"),
					resource.TestCheckResourceAttr(addr, "assignment.0.condition.0.match.1.value", "cto@example.com"),
					resource.TestCheckResourceAttr(addr, "assignment.0.condition.1.match.0.source", "agent"),
					mcpAppAssignments(addr, `[{"all_of":[{"any_of":[`+
						`{"op":"in","property":"groups","source":"user","value":["grp-eng","grp-sre"]},`+
						`{"op":"eq","property":"email","source":"user","value":"cto@example.com"}]},`+
						`{"any_of":[{"op":"neq","property":"id","source":"agent","value":"agt-blocked"}]}]},`+
						`{"all_of":[{"any_of":[{"op":"eq","property":"department","source":"user","value":"finance"}]}]}]`),
				),
			},
			{Config: assignmentAppConfig(twoRules), PlanOnly: true},
			{
				Config: assignmentAppConfig("  assign_all = true\n"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "assign_all", "true"),
					resource.TestCheckResourceAttr(addr, "assignment.#", "0"),
					mcpAppAssignments(addr, `[]`),
				),
			},
			{
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
				// Recovered lowercased from the name; see the base test.
				ImportStateVerifyIgnore: []string{"url_label"},
			},
		},
	})
}

func TestAccResourceMCPApplication_assignmentChecks(t *testing.T) {
	match := func(body string) string {
		return "  assignment {\n    condition {\n      match {\n" + body + "\n      }\n    }\n  }\n"
	}
	for name, tc := range map[string]struct{ body, expect string }{
		"in with value": {
			body:   `source = "user"` + "\n" + `property = "groups"` + "\n" + `op = "in"` + "\n" + `value = "grp"`,
			expect: `assignment\[0\].condition\[0\].match\[0\]: op "in" compares against values, not value`,
		},
		"nin without values": {
			body:   `source = "user"` + "\n" + `property = "groups"` + "\n" + `op = "nin"`,
			expect: `op "nin" needs at least one entry in values`,
		},
		"eq with values": {
			body:   `source = "user"` + "\n" + `property = "email"` + "\n" + `op = "eq"` + "\n" + `values = ["a"]`,
			expect: `op "eq" compares against value, not values`,
		},
		"agent property other than id": {
			body:   `source = "agent"` + "\n" + `property = "type"` + "\n" + `op = "eq"` + "\n" + `value = "x"`,
			expect: `an agent match can only compare property "id"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config:      assignmentAppConfig(match(tc.body)),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tc.expect),
				}},
			})
		})
	}
}
