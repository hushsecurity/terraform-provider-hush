# Create a Jira integration
resource "hush_jira_integration" "example" {
  name       = "my-jira"
  org_domain = "mycompany.atlassian.net"
  user       = "admin@mycompany.com"
  api_key    = var.jira_api_key
}

# With optional settings
resource "hush_jira_integration" "full" {
  name                   = "my-jira-full"
  description            = "Production Jira integration"
  org_domain             = "mycompany.atlassian.net"
  user                   = "admin@mycompany.com"
  api_key                = var.jira_api_key
  sync_issues_resolution = true
  enable_scans           = true
}
