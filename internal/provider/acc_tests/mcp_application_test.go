package acc_tests

import (
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// mcpAppsDestroyed checks that the applications a test created, by display
// name, are gone. Tests run in parallel against one fake, so it looks at no
// others.
func mcpAppsDestroyed(displayNames ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		for id, app := range heimdall.apps {
			if slices.Contains(displayNames, app["display_name"].(string)) {
				return fmt.Errorf("application %s (%s) still exists", id, app["display_name"])
			}
		}
		return nil
	}
}

// mcpAppSent checks the client secret heimdall last received, which neither
// state nor a read can show.
func mcpAppSent(addr string, want any) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("%s not in state", addr)
		}
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		if got := heimdall.appSecrets[rs.Primary.ID]; got != want {
			return fmt.Errorf("%s: heimdall holds client secret %v, want %v", addr, got, want)
		}
		return nil
	}
}

func datadogAppConfig(body string) string {
	return `
resource "hush_mcp_application" "datadog" {
  app_catalog_id = "datadog"
  display_name   = "acc-datadog"
  url_label      = "EU"
` + body + "}\n"
}

func TestAccResourceMCPApplication(t *testing.T) {
	const addr = "hush_mcp_application.datadog"
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-datadog"),
		Steps: []resource.TestStep{
			{
				Config: datadogAppConfig(`
  description    = "first"
  deployment_ids = ["` + mockDeploymentID + `"]
  allowed_agents = ["claude-code", "cursor"]
  scopes         = ["read"]
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "name", "datadog-eu"),
					resource.TestCheckResourceAttr(addr, "url", "https://mcp.datadoghq.eu/v1/mcp"),
					resource.TestCheckResourceAttr(addr, "catalog_display_name", "Datadog"),
					resource.TestCheckResourceAttr(addr, "enabled", "true"),
					resource.TestCheckResourceAttr(addr, "allowed_agents.#", "2"),
					resource.TestCheckResourceAttr(addr, "scopes.0", "read"),
					resource.TestCheckResourceAttr(addr, "tools.#", "3"),
					// Each tool reports the operation it resolves to.
					resource.TestCheckResourceAttr(addr, "tools.0.operation", "allow"),
					resource.TestCheckResourceAttr(addr, "tools.2.operation", "block"),
				),
			},
			{
				Config: datadogAppConfig(`
  description    = "first"
  deployment_ids = ["` + mockDeploymentID + `"]
  allowed_agents = ["claude-code", "cursor"]
  scopes         = ["read"]
`),
				PlanOnly: true,
			},
			// Clearing: the description and the agent limit go back to none.
			{
				Config: datadogAppConfig(`
  deployment_ids = ["` + mockDeploymentID + `", "` + mockDeploymentID2 + `"]
  enabled        = false
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "description", ""),
					resource.TestCheckResourceAttr(addr, "allowed_agents.#", "0"),
					resource.TestCheckResourceAttr(addr, "deployment_ids.#", "2"),
					resource.TestCheckResourceAttr(addr, "enabled", "false"),
					// Left unset, scopes keep what the application holds.
					resource.TestCheckResourceAttr(addr, "scopes.0", "read"),
				),
			},
			// An import learns url_label from the name, lowercased, which the
			// configuration's "EU" must not read as a change.
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"url_label"},
			},
		},
	})
}

// The client secret is never returned; it is sent at creation, re-sent when
// its version changes, and cleared along with the client id.
func TestAccResourceMCPApplication_clientCredentials(t *testing.T) {
	const addr = "hush_mcp_application.slack"
	config := func(creds string) string {
		return `
resource "hush_mcp_application" "slack" {
  app_catalog_id = "slack"
  display_name   = "acc-slack"
  deployment_ids = ["` + mockDeploymentID + `"]
` + creds + "}\n"
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-slack"),
		Steps: []resource.TestStep{
			{
				Config: config(`
  client_id                = "cid"
  client_secret_wo         = "s1"
  client_secret_wo_version = "1"
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "client_id", "cid"),
					mcpAppSent(addr, "s1"),
				),
			},
			{
				Config: config(`
  client_id                = "cid"
  client_secret_wo         = "s2"
  client_secret_wo_version = "2"
`),
				Check: mcpAppSent(addr, "s2"),
			},
			// An entry with manual registration needs the credentials only to
			// be created, so they can be removed afterwards.
			{
				Config: config(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "client_id", ""),
					mcpAppSent(addr, nil),
				),
			},
		},
	})
}

// A gateway that has not published its key refuses to enable an application
// with a client secret. The create is made disabled, so the refusal comes on
// the patch that enables it, with the id already in state: the next apply
// replaces the tainted application rather than tripping over one it never
// learned the id of.
func TestAccResourceMCPApplication_gatewayWithoutKey(t *testing.T) {
	const keyless = "dep-acc-keyless"
	config := `
resource "hush_mcp_application" "early" {
  app_catalog_id = "slack"
  display_name   = "acc-early"
  deployment_ids = ["` + keyless + `"]
  client_id      = "cid"
  client_secret  = "s"
}
`
	setKeyless := func(missing bool) func() {
		return func() {
			heimdall.mu.Lock()
			defer heimdall.mu.Unlock()
			heimdall.keyless[keyless] = missing
		}
	}
	countApps := func(want int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			heimdall.mu.Lock()
			defer heimdall.mu.Unlock()
			n := 0
			for _, app := range heimdall.apps {
				if app["display_name"] == "acc-early" {
					n++
				}
			}
			if n != want {
				return fmt.Errorf("heimdall holds %d acc-early applications, want %d", n, want)
			}
			return nil
		}
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-early"),
		Steps: []resource.TestStep{
			{
				PreConfig:   setKeyless(true),
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s)no public key.*Apply again once the gateway is up`),
			},
			{
				PreConfig: setKeyless(false),
				Config:    config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_mcp_application.early", "enabled", "true"),
					countApps(1),
				),
			},
		},
	})
}

// An empty scopes list is sent as one, asking for no scopes, rather than read
// as unset and left to the entry's defaults.
func TestAccResourceMCPApplication_emptyScopes(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-no-scopes"),
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_mcp_application" "no_scopes" {
  app_catalog_id = "slack"
  display_name   = "acc-no-scopes"
  deployment_ids = ["` + mockDeploymentID + `"]
  client_id      = "cid"
  client_secret  = "s"
  scopes         = []
}
`,
				Check: resource.TestCheckResourceAttr("hush_mcp_application.no_scopes", "scopes.#", "0"),
			},
		},
	})
}

// Only an MCP application with a catalog id can be imported: one created
// before applications carried it leaves no entry to read it through.
func TestAccResourceMCPApplication_importLegacy(t *testing.T) {
	heimdall.mu.Lock()
	heimdall.apps["app-acc-legacy"] = map[string]any{
		"id": "app-acc-legacy", "type": "mcp", "app_catalog_id": nil, "name": "linear", "display_name": "acc-legacy",
	}
	heimdall.mu.Unlock()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_mcp_application" "legacy" {
  app_catalog_id = "linear"
  display_name   = "acc-legacy"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`,
				ResourceName:  "hush_mcp_application.legacy",
				ImportState:   true,
				ImportStateId: "app-acc-legacy",
				ExpectError:   regexp.MustCompile(`app-acc-legacy predates catalog ids`),
			},
		},
	})
}

// Only an MCP application can be imported.
func TestAccResourceMCPApplication_importOtherType(t *testing.T) {
	heimdall.mu.Lock()
	heimdall.apps["app-acc-db"] = map[string]any{
		"id": "app-acc-db", "type": "db", "app_catalog_id": "postgres", "display_name": "acc-db",
	}
	heimdall.mu.Unlock()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "hush_mcp_application" "db" {
  app_catalog_id = "postgres"
  display_name   = "acc-db"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`,
				ResourceName:  "hush_mcp_application.db",
				ImportState:   true,
				ImportStateId: "app-acc-db",
				ExpectError:   regexp.MustCompile(`cannot manage it: application app-acc-db is a db application`),
			},
		},
	})
}

// The Google and QuickBooks entries are read and patched through routes of
// their own, which carry settings the generic ones refuse.
func TestAccResourceMCPApplication_entrySettings(t *testing.T) {
	registerFakeCatalogEntry("gmail", false)
	config := func(project, company string, sandbox bool) string {
		return fmt.Sprintf(`
resource "hush_mcp_application" "gmail" {
  app_catalog_id    = "gmail"
  display_name      = "acc-gmail"
  deployment_ids    = [%[4]q]
  google_project_id = %[1]q
}

resource "hush_mcp_application" "quickbooks" {
  app_catalog_id = "quickbooks"
  display_name   = "acc-quickbooks"
  deployment_ids = [%[4]q]
  client_id      = "cid"
  client_secret  = "s"

  quickbooks {
    company_id = %[2]q
    sandbox    = %[3]t
  }
}
`, project, company, sandbox, mockDeploymentID)
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-gmail", "acc-quickbooks"),
		Steps: []resource.TestStep{
			{
				Config: config("acc-project", "1234", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_mcp_application.gmail", "google_project_id", "acc-project"),
					resource.TestCheckResourceAttr("hush_mcp_application.quickbooks", "quickbooks.0.company_id", "1234"),
					resource.TestCheckResourceAttr("hush_mcp_application.quickbooks", "quickbooks.0.sandbox", "true"),
					resource.TestCheckResourceAttr("hush_mcp_application.quickbooks", "hosted", "true"),
					resource.TestCheckResourceAttr("hush_mcp_application.quickbooks", "url", ""),
				),
			},
			{Config: config("acc-project", "1234", true), PlanOnly: true},
			{
				Config: config("acc-project-2", "5678", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_mcp_application.gmail", "google_project_id", "acc-project-2"),
					resource.TestCheckResourceAttr("hush_mcp_application.quickbooks", "quickbooks.0.company_id", "5678"),
					resource.TestCheckResourceAttr("hush_mcp_application.quickbooks", "quickbooks.0.sandbox", "false"),
				),
			},
			{
				ResourceName:      "hush_mcp_application.gmail",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// An application created from a custom app, and the custom app refusing to
// be destroyed while an application still uses it.
func TestAccResourceMCPApplication_fromCustomApp(t *testing.T) {
	const custom = `
resource "hush_custom_mcp_application" "c" {
  name         = "acc-from-custom"
  display_name = "acc custom source"
  url {
    url = "https://mcp.example.com/mcp"
  }
  tool {
    name = "search"
    type = "read"
  }
}
`
	// An application this configuration does not manage, using the custom app.
	seedUser := func() {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		heimdall.apps["app-acc-unmanaged"] = map[string]any{
			"display_name": "acc-unmanaged", "app_catalog_id": "custom-acc-from-custom",
		}
	}
	dropUser := func() {
		heimdall.mu.Lock()
		defer heimdall.mu.Unlock()
		delete(heimdall.apps, "app-acc-unmanaged")
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-from-custom"),
		Steps: []resource.TestStep{
			{
				Config: custom + `
resource "hush_mcp_application" "custom" {
  app_catalog_id = hush_custom_mcp_application.c.app_catalog_id
  display_name   = "acc-from-custom"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("hush_mcp_application.custom", "app_catalog_id", "custom-acc-from-custom"),
					resource.TestCheckResourceAttr("hush_mcp_application.custom", "name", "custom-acc-from-custom"),
					resource.TestCheckResourceAttr("hush_mcp_application.custom", "tools.0.name", "search"),
				),
			},
			{
				PreConfig:   seedUser,
				Config:      custom,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`custom app acc-from-custom is still used by applications`),
			},
			{
				PreConfig: dropUser,
				Config:    custom,
			},
		},
	})
}

// What heimdall refuses at apply, refused at plan time from the catalog.
func TestAccResourceMCPApplication_planChecks(t *testing.T) {
	for name, tc := range map[string]struct{ body, expect string }{
		"unknown entry": {
			body:   `app_catalog_id = "nope"`,
			expect: `the catalog has no MCP entry "nope"`,
		},
		"label missing among several": {
			body:   `app_catalog_id = "datadog"`,
			expect: `"datadog" has several addresses; pick one of US1, EU`,
		},
		"label unknown": {
			body:   "app_catalog_id = \"datadog\"\n  url_label = \"MARS\"",
			expect: `"MARS" is not one of datadog's labels: US1, EU`,
		},
		"label on a single address": {
			body:   "app_catalog_id = \"slack\"\n  url_label = \"x\"\n  client_id = \"c\"\n  client_secret = \"s\"",
			expect: `"slack" has a single address and takes no label`,
		},
		"manual registration without credentials": {
			body:   `app_catalog_id = "slack"`,
			expect: `"slack" needs an OAuth app you registered with it`,
		},
		// heimdall wants both halves, so a client id alone is refused too.
		"manual registration without a secret": {
			body:   "app_catalog_id = \"slack\"\n  client_id = \"c\"",
			expect: `"slack" needs an OAuth app you registered with it`,
		},
		"google entry without a project": {
			body:   `app_catalog_id = "gdrive"`,
			expect: `google_project_id: required for "gdrive"`,
		},
		"project on another entry": {
			body:   "app_catalog_id = \"datadog\"\n  url_label = \"EU\"\n  google_project_id = \"acc-project\"",
			expect: `google_project_id: only valid for the Google Workspace entries`,
		},
		"quickbooks without a company": {
			body:   "app_catalog_id = \"quickbooks\"\n  client_id = \"c\"\n  client_secret = \"s\"",
			expect: `a quickbooks block is required for "quickbooks"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{{
					Config: `
resource "hush_mcp_application" "bad" {
  display_name   = "acc-bad"
  deployment_ids = ["` + mockDeploymentID + `"]
  ` + tc.body + `
}
`,
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(tc.expect),
				}},
			})
		})
	}
}

// The data source finds an application by display name or by id, and reports
// it as the resource does, less the client secret.
func TestAccDataSourceMCPApplication(t *testing.T) {
	const config = `
resource "hush_mcp_application" "src" {
  app_catalog_id    = "gmail"
  display_name      = "acc-ds-gmail"
  deployment_ids    = ["` + mockDeploymentID + `"]
  google_project_id = "acc-project"
  assign_all        = true

  tool_operation {
    name      = "search"
    operation = "block"
  }
}

data "hush_mcp_application" "by_name" {
  display_name = hush_mcp_application.src.display_name
}

data "hush_mcp_application" "by_id" {
  id = hush_mcp_application.src.id
}
`
	registerFakeCatalogEntry("gmail", false)
	heimdall.mu.Lock()
	heimdall.catalog["gmail"]["tools"] = []map[string]any{
		{"name": "search", "type": "read", "description": "Search", "operation": nil},
	}
	heimdall.mu.Unlock()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-ds-gmail"),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.hush_mcp_application.by_name", "id", "hush_mcp_application.src", "id"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_name", "app_catalog_id", "gmail"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_name", "google_project_id", "acc-project"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_name", "assign_all", "true"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_name", "tool_operation.#", "1"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_name", "tools.0.operation", "block"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_id", "display_name", "acc-ds-gmail"),
					resource.TestCheckResourceAttr("data.hush_mcp_application.by_id", "tool_defaults.0.write", "user_consent"),
				),
			},
			{
				Config: `
data "hush_mcp_application" "none" {
  display_name = "acc-no-such-app"
}
`,
				ExpectError: regexp.MustCompile(`no MCP application found with display_name: acc-no-such-app`),
			},
		},
	})
}

// Display names are unique in heimdall; were two ever to match, the data
// source refuses to pick one rather than report either.
func TestAccDataSourceMCPApplication_ambiguousName(t *testing.T) {
	heimdall.mu.Lock()
	for _, id := range []string{"app-acc-twin-1", "app-acc-twin-2"} {
		heimdall.apps[id] = map[string]any{
			"id": id, "type": "mcp", "name": "slack", "app_catalog_id": "slack", "display_name": "acc-twin",
		}
	}
	heimdall.mu.Unlock()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{{
			Config: `
data "hush_mcp_application" "twin" {
  display_name = "acc-twin"
}
`,
			ExpectError: regexp.MustCompile(`multiple MCP applications found with display_name: acc-twin`),
		}},
	})
}

// An empty allowed_agents is no limit, as leaving it out is: it is what the
// configuration an import generates writes for an application without one,
// and heimdall refuses an empty list, so none is ever sent.
func TestAccResourceMCPApplication_emptyAllowedAgents(t *testing.T) {
	const addr = "hush_mcp_application.any_agent"
	config := func(agents string) string {
		return `
resource "hush_mcp_application" "any_agent" {
  app_catalog_id = "datadog"
  display_name   = "acc-any-agent"
  url_label      = "US1"
  deployment_ids = ["` + mockDeploymentID + `"]
  allowed_agents = ` + agents + `
}
`
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-any-agent"),
		Steps: []resource.TestStep{
			{
				Config: config(`[]`),
				Check:  resource.TestCheckResourceAttr(addr, "allowed_agents.#", "0"),
			},
			{
				Config: config(`["cursor"]`),
				Check:  resource.TestCheckResourceAttr(addr, "allowed_agents.#", "1"),
			},
			// Back to empty lifts the limit: heimdall is sent null.
			{
				Config: config(`[]`),
				Check: func(s *terraform.State) error {
					heimdall.mu.Lock()
					defer heimdall.mu.Unlock()
					if got := heimdall.apps[s.RootModule().Resources[addr].Primary.ID]["allowed_agents"]; got != nil {
						return fmt.Errorf("heimdall holds allowed_agents %v, want null", got)
					}
					return nil
				},
			},
		},
	})
}

// An application of a custom app that names no OAuth client inherits the
// custom app's, and follows it when it changes. What a read finds is then
// inherited, not drift: were it planned as a change to empty, the apply would
// send a null that strips the credentials.
func TestAccResourceMCPApplication_inheritsCustomAppClient(t *testing.T) {
	const addr = "hush_mcp_application.inherits"
	config := func(clientID, secret string) string {
		return `
resource "hush_custom_mcp_application" "src" {
  name          = "acc-inherit-src"
  display_name  = "acc inherit source"
  client_id     = "` + clientID + `"
  client_secret = "` + secret + `"
  url {
    url = "https://mcp.example.com/mcp"
  }
}

resource "hush_mcp_application" "inherits" {
  app_catalog_id = hush_custom_mcp_application.src.app_catalog_id
  display_name   = "acc-inherits"
  deployment_ids = ["` + mockDeploymentID + `"]
}
`
	}
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      mcpAppsDestroyed("acc-inherits"),
		Steps: []resource.TestStep{
			{
				Config: config("cid-1", "s-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "client_id", "cid-1"),
					mcpAppSent(addr, "s-1"),
				),
			},
			{Config: config("cid-1", "s-1"), PlanOnly: true},
			// The custom app's new client reaches the application, which
			// plans nothing and keeps the credentials.
			{
				Config: config("cid-2", "s-2"),
				Check: resource.ComposeTestCheckFunc(
					mcpAppSent(addr, "s-2"),
				),
			},
			{
				Config: config("cid-2", "s-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "client_id", "cid-2"),
					mcpAppSent(addr, "s-2"),
				),
			},
			{Config: config("cid-2", "s-2"), PlanOnly: true},
		},
	})
}
