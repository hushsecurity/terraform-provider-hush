package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func rateLimitAppConfig(body string) string {
	return `
resource "hush_mcp_application" "limits" {
  app_catalog_id = "datadog"
  display_name   = "acc-rate-limits"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
` + body + `}
`
}

const rateLimits = `
  rate_limit {
    target_type = "group"
    target      = "write"
    max_calls   = 30
    window      = 300
  }
  rate_limit {
    target_type = "tool"
    target      = "search"
    max_calls   = 10
    window      = 60
  }
  rate_limit {
    target_type = "app"
    max_calls   = 500
    window      = 3600
    mode        = "warn"
  }
`

func TestAccResourceMCPApplication_rateLimits(t *testing.T) {
	const addr = "hush_mcp_application.limits"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-rate-limits"),
		Steps: []resource.TestStep{
			{
				Config: rateLimitAppConfig(""),
				Check:  resource.TestCheckResourceAttr(addr, "rate_limit.#", "0"),
			},
			{
				Config: rateLimitAppConfig(rateLimits),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "rate_limit.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(addr, "rate_limit.*", map[string]string{
						"target_type": "group", "target": "write", "max_calls": "30", "window": "300", "mode": "block",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(addr, "rate_limit.*", map[string]string{
						"target_type": "app", "target": "", "max_calls": "500", "mode": "warn",
					}),
				),
			},
			{
				Config:   rateLimitAppConfig(rateLimits),
				PlanOnly: true,
			},
			// A changed limit is set again; a removed block removes its limit.
			{
				Config: rateLimitAppConfig(`
  rate_limit {
    target_type = "tool"
    target      = "search"
    max_calls   = 5
    window      = 60
  }
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "rate_limit.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(addr, "rate_limit.*", map[string]string{
						"target_type": "tool", "target": "search", "max_calls": "5",
					}),
				),
			},
			{
				Config: rateLimitAppConfig(""),
				Check:  resource.TestCheckResourceAttr(addr, "rate_limit.#", "0"),
			},
		},
	})
}

func TestAccResourceMCPApplication_rateLimitsRefused(t *testing.T) {
	for name, tc := range map[string]struct{ body, err string }{
		"two blocks for one target": {`
  rate_limit {
    target_type = "group"
    target      = "write"
    max_calls   = 1
    window      = 60
  }
  rate_limit {
    target_type = "group"
    target      = "write"
    max_calls   = 2
    window      = 60
  }
`, `group "write" has more than one block`},
		"app with a target": {`
  rate_limit {
    target_type = "app"
    target      = "x"
    max_calls   = 1
    window      = 60
  }
`, `an 'app' limit takes no target`},
		"unknown group": {`
  rate_limit {
    target_type = "group"
    target      = "admin"
    max_calls   = 1
    window      = 60
  }
`, `group "admin" is not one of`},
		"unknown tool": {`
  rate_limit {
    target_type = "tool"
    target      = "nope"
    max_calls   = 1
    window      = 60
  }
`, `has no tool "nope"`},
		"window too short": {`
  rate_limit {
    target_type = "app"
    max_calls   = 1
    window      = 59
  }
`, `window to be in the range \(60 - 86400\)`},
	} {
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config:      rateLimitAppConfig(tc.body),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tc.err),
				}},
			})
		})
	}
}
