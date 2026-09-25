# Datadog, in its EU region, on the gateway of one deployment. The catalog
# entry lists the regions to pick from.
data "hush_mcp_catalog_entry" "datadog_app" {
  app_catalog_id = "datadog"
}

resource "hush_mcp_application" "datadog" {
  app_catalog_id = data.hush_mcp_catalog_entry.datadog_app.app_catalog_id
  display_name   = "Datadog"
  deployment_ids = [hush_deployment.gateway.id]
  url_label      = "EU"

  lifecycle {
    precondition {
      condition     = contains(data.hush_mcp_catalog_entry.datadog_app.urls[*].label, "EU")
      error_message = "Datadog no longer offers an EU address."
    }
  }
}

# GitHub needs an OAuth app you registered with it, and is limited to two
# kinds of agent.
resource "hush_mcp_application" "github" {
  app_catalog_id = "github"
  display_name   = "GitHub"
  description    = "Repositories of the platform team"
  deployment_ids = [hush_deployment.gateway.id]
  allowed_agents = ["claude-code", "cursor"]
  scopes         = ["repo", "read:org"]

  client_id                = "Iv1.0123456789abcdef"
  client_secret_wo         = var.github_oauth_client_secret
  client_secret_wo_version = "1"

  # Granted to the platform and SRE groups, or to the CTO -- through any agent
  # but one.
  assignment {
    condition {
      match {
        source   = "user"
        property = "groups"
        op       = "in"
        values   = ["grp-platform", "grp-sre"]
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
        value    = "agt-shared-ci"
      }
    }
  }
}

# Google Workspace servers bill their quota to a Google Cloud project.
resource "hush_mcp_application" "gmail" {
  app_catalog_id    = "gmail"
  display_name      = "Gmail"
  deployment_ids    = [hush_deployment.gateway.id]
  google_project_id = "acme-workspace-mcp"
  assign_all        = true

  client_id                = "0123456789-abcdef.apps.googleusercontent.com"
  client_secret_wo         = var.google_oauth_client_secret
  client_secret_wo_version = "1"
}

# An application of a custom MCP server.
resource "hush_mcp_application" "internal_tools_eu" {
  app_catalog_id = hush_custom_mcp_application.internal_tools.app_catalog_id
  display_name   = "Internal tools (EU)"
  deployment_ids = [hush_deployment.gateway.id]
  url_label      = "eu"
}
