package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMCPCatalogEntry(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "hush_mcp_catalog_entry" "datadog" {
  app_catalog_id = "datadog"
}
data "hush_mcp_catalog_entry" "quickbooks" {
  app_catalog_id = "quickbooks"
}
data "hush_mcp_catalog_entry" "slack" {
  app_catalog_id = "slack"
}
data "hush_mcp_catalog_entry" "crowdstrike" {
  app_catalog_id = "crowdstrike"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "id", "datadog"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "display_name", "Datadog"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "urls.#", "2"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "urls.1.label", "EU"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "hosted", "false"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "tools.#", "3"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "tools.2.type", "destructive"),
					// A hosted entry has no address to pick.
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.quickbooks", "hosted", "true"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.quickbooks", "urls.#", "0"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.quickbooks", "manual_registration", "true"),
					// A single address carries no label, which reads as empty.
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.slack", "urls.0.label", ""),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.slack", "scopes.#", "2"),
					// A hosted entry with several addresses lists their labels only.
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.crowdstrike", "urls.#", "2"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.crowdstrike", "urls.1.label", "EU-1"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.crowdstrike", "urls.1.url", ""),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.crowdstrike", "auth_types.0", "client_credentials"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.crowdstrike", "auth_required", "true"),
					resource.TestCheckResourceAttr("data.hush_mcp_catalog_entry.datadog", "auth_required", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceMCPCatalogEntry_unknown(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "hush_mcp_catalog_entry" "nope" {
  app_catalog_id = "nope"
}
`,
				ExpectError: regexp.MustCompile(`mcp/nope not found`),
			},
		},
	})
}
