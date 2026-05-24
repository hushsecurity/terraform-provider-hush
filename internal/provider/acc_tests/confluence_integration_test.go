package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hushsecurity/terraform-provider-hush/internal/testutil"
)

func init() {
	registerMockSetup(func(ms *testutil.MockServer) {
		ms.OnOperation("integrations", testutil.OpCreate, func(op testutil.Operation, obj map[string]any) *testutil.HookError {
			if _, has := obj["org_domain"]; has {
				if _, hasUser := obj["user"]; hasUser {
					if _, hasSync := obj["sync_issues_resolution"]; hasSync {
						obj["type"] = "jira"
					} else {
						obj["type"] = "confluence"
					}
				}
			}
			return nil
		})
	})
}

func TestAccResourceConfluenceIntegration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("confluence_integration", "v1/integrations"),
		Steps: []resource.TestStep{
			{
				Config: confluenceIntegrationStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_confluence_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_confluence_integration.test", "name", "test-confluence",
					),
					resource.TestCheckResourceAttr(
						"hush_confluence_integration.test", "org_domain", "testcompany.atlassian.net",
					),
					resource.TestCheckResourceAttr(
						"hush_confluence_integration.test", "user", "admin@testcompany.com",
					),
					resource.TestCheckResourceAttr(
						"hush_confluence_integration.test", "status", "ok",
					),
				),
			},
			{
				Config: confluenceIntegrationStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_confluence_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_confluence_integration.test", "name", "test-confluence-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_confluence_integration.test", "description", "updated description",
					),
				),
			},
			{
				ResourceName:            "hush_confluence_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key", "user"},
			},
		},
	})
}

func TestAccDataSourceConfluenceIntegration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: confluenceIntegrationDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_confluence_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_confluence_integration.test", "name", "test-confluence-ds",
					),
				),
			},
		},
	})
}

const confluenceIntegrationStep1 = `
resource "hush_confluence_integration" "test" {
  name       = "test-confluence"
  org_domain = "testcompany.atlassian.net"
  user       = "admin@testcompany.com"
  api_key    = "test-api-key-12345"
}
`

const confluenceIntegrationStep2 = `
resource "hush_confluence_integration" "test" {
  name        = "test-confluence-updated"
  description = "updated description"
  org_domain  = "testcompany.atlassian.net"
  user        = "admin@testcompany.com"
  api_key     = "test-api-key-12345"
}
`

const confluenceIntegrationDataSource = `
resource "hush_confluence_integration" "ds_source" {
  name       = "test-confluence-ds"
  org_domain = "testcompany.atlassian.net"
  user       = "admin@testcompany.com"
  api_key    = "test-api-key-ds"
}

data "hush_confluence_integration" "test" {
  id = hush_confluence_integration.ds_source.id
}
`
