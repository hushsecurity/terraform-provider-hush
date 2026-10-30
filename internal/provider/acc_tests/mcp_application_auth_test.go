package acc_tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func crowdstrikeAppConfig(auth string) string {
	return `
resource "hush_mcp_application" "falcon" {
  app_catalog_id = "crowdstrike"
  display_name   = "acc-falcon"
  url_label      = "EU-1"
  deployment_ids = ["` + mockDeploymentID + `"]
` + auth + "}\n"
}

// The entry picks the token URL from the region. The secret is never
// returned; a new one is sent in place, and removing auth replaces the app.
func TestAccResourceMCPApplication_auth(t *testing.T) {
	const addr = "hush_mcp_application.falcon"
	auth := func(version string) string {
		return `
  auth {
    type              = "client_credentials"
    client_id         = "falcon-client"
    secret_wo         = "s` + version + `"
    secret_wo_version = "` + version + `"
  }
`
	}
	var firstID string
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-falcon"),
		Steps: []resource.TestStep{
			{
				Config: crowdstrikeAppConfig(auth("1")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", "crowdstrike-eu-1"),
					resource.TestCheckResourceAttr(addr, "auth.0.type", "client_credentials"),
					resource.TestCheckResourceAttr(addr, "auth.0.client_id", "falcon-client"),
					resource.TestCheckResourceAttr(addr, "auth.0.token_url", "https://api.eu-1.crowdstrike.com/oauth2/token"),
					resource.TestCheckResourceAttr(addr, "auth.0.auth_method", "post"),
					mcpAppSent(addr, "s1"),
					resource.TestCheckResourceAttrWith(addr, "id", func(id string) error {
						firstID = id
						return nil
					}),
				),
			},
			{Config: crowdstrikeAppConfig(auth("1")), PlanOnly: true},
			{
				Config: crowdstrikeAppConfig(auth("2")),
				Check: resource.ComposeTestCheckFunc(
					mcpAppSent(addr, "s2"),
					resource.TestCheckResourceAttrWith(addr, "id", func(id string) error {
						if id != firstID {
							return fmt.Errorf("a new secret replaced %s with %s", firstID, id)
						}
						return nil
					}),
				),
			},
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"url_label", "auth.0.secret_wo_version"},
			},
		},
	})
}

// An import has no secret version; the configuration's re-sends the secret
// rather than replacing the application.
func TestAccResourceMCPApplication_authImport(t *testing.T) {
	const addr, id = "hush_mcp_application.falcon", "app-acc-falcon-import"
	heimdall.mu.Lock()
	heimdall.apps[id] = map[string]any{
		"id": id, "name": "crowdstrike-eu-1", "type": "mcp", "app_catalog_id": "crowdstrike",
		"catalog_display_name": "CrowdStrike Falcon", "display_name": "acc-falcon-import",
		"description": nil, "deployment_ids": []any{mockDeploymentID}, "allowed_agents": nil,
		"enabled": true, "assignments": []any{}, "assign_all": false, "url": nil, "hosted": true,
		"scopes": []any{}, "oauth_relay": false, "client_id": nil, "alias": nil,
		"tool_groups": []any{}, "tools": []any{},
		"auth": map[string]any{
			"type": "client_credentials", "client_id": "falcon-client", "client_secret": "****",
			"token_url": "https://api.eu-1.crowdstrike.com/oauth2/token", "auth_method": "post",
			"scopes": []any{},
		},
	}
	heimdall.mu.Unlock()
	config := strings.Replace(crowdstrikeAppConfig(`
  auth {
    type              = "client_credentials"
    client_id         = "falcon-client"
    secret_wo         = "s1"
    secret_wo_version = "1"
  }
`), "acc-falcon", "acc-falcon-import", 1)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-falcon-import"),
		Steps: []resource.TestStep{
			{
				Config:             config,
				ResourceName:       addr,
				ImportState:        true,
				ImportStateId:      id,
				ImportStatePersist: true,
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", id),
					resource.TestCheckResourceAttr(addr, "auth.0.secret_wo_version", "1"),
				),
			},
		},
	})
}

func TestAccResourceMCPApplication_authPlanErrors(t *testing.T) {
	steps := []struct{ config, err string }{
		{crowdstrikeAppConfig(""), `"crowdstrike" takes no OAuth, so an auth block is required`},
		{crowdstrikeAppConfig(`
  auth {
    type   = "bearer"
    secret = "t"
  }
`), `"crowdstrike" takes type client_credentials, not "bearer"`},
		{crowdstrikeAppConfig(`
  auth {
    type   = "client_credentials"
    secret = "t"
  }
`), `client_id is required with type "client_credentials"`},
		{crowdstrikeAppConfig(`
  auth {
    type      = "client_credentials"
    client_id = "c"
  }
`), `one of secret or secret_wo is required`},
		{datadogAppConfig(`
  deployment_ids = ["` + mockDeploymentID + `"]
  auth {
    type   = "bearer"
    secret = "t"
  }
`), `"datadog" takes no auth block`},
	}
	var testSteps []resource.TestStep
	for _, step := range steps {
		testSteps = append(testSteps, resource.TestStep{
			Config:      step.config,
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(regexp.QuoteMeta(step.err)),
		})
	}
	resource.ParallelTest(t, resource.TestCase{ProviderFactories: providerFactories, Steps: testSteps})
}
