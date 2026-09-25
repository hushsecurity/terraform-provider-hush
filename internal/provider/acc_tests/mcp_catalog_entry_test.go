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
