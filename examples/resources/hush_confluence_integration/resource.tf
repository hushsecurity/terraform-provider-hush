# Manage a Confluence integration
resource "hush_confluence_integration" "example" {
  name       = "my-confluence"
  org_domain = "mycompany.atlassian.net"
  user       = "admin@mycompany.com"
  api_key    = var.confluence_api_key
}

# With write-only API key (recommended for production)
resource "hush_confluence_integration" "secure" {
  name              = "my-confluence-secure"
  org_domain        = "mycompany.atlassian.net"
  user              = "admin@mycompany.com"
  api_key_wo        = var.confluence_api_key
  api_key_wo_version = "v1"
}
