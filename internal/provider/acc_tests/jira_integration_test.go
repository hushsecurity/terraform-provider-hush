package acc_tests

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceJiraIntegration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		CheckDestroy:      validateResourceDestroyed("jira_integration", "v1/integrations"),
		Steps: []resource.TestStep{
			{
				Config: jiraIntegrationStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_jira_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "name", "test-jira",
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "org_domain", "testcompany.atlassian.net",
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "user", "admin@testcompany.com",
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "sync_issues_resolution", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "enable_scans", "true",
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "status", "ok",
					),
				),
			},
			{
				Config: jiraIntegrationStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"hush_jira_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "name", "test-jira-updated",
					),
					resource.TestCheckResourceAttr(
						"hush_jira_integration.test", "sync_issues_resolution", "false",
					),
				),
			},
			{
				ResourceName:            "hush_jira_integration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key", "user"},
			},
		},
	})
}

func TestAccDataSourceJiraIntegration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: jiraIntegrationDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.hush_jira_integration.test", "id", regexp.MustCompile(`^int-.+$`),
					),
					resource.TestCheckResourceAttr(
						"data.hush_jira_integration.test", "name", "test-jira-ds",
					),
				),
			},
		},
	})
}

const jiraIntegrationStep1 = `
resource "hush_jira_integration" "test" {
  name                   = "test-jira"
  org_domain             = "testcompany.atlassian.net"
  user                   = "admin@testcompany.com"
  api_key                = "test-api-key-jira"
  sync_issues_resolution = true
  enable_scans           = true
}
`

const jiraIntegrationStep2 = `
resource "hush_jira_integration" "test" {
  name                   = "test-jira-updated"
  org_domain             = "testcompany.atlassian.net"
  user                   = "admin@testcompany.com"
  api_key                = "test-api-key-jira"
  sync_issues_resolution = false
  enable_scans           = true
}
`

const jiraIntegrationDataSource = `
resource "hush_jira_integration" "ds_source" {
  name                   = "test-jira-ds"
  org_domain             = "testcompany.atlassian.net"
  user                   = "admin@testcompany.com"
  api_key                = "test-api-key-jira-ds"
  sync_issues_resolution = true
  enable_scans           = true
}

data "hush_jira_integration" "test" {
  id = hush_jira_integration.ds_source.id
}
`
