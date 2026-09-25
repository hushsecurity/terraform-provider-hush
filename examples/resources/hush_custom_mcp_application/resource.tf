# An internal MCP server in two regions, authenticated with an API key header.
# Create applications from it with
# app_catalog_id = hush_custom_mcp_application.internal_tools.app_catalog_id.
resource "hush_custom_mcp_application" "internal_tools" {
  name         = "internal-tools"
  display_name = "Internal tools"
  description  = "The platform team's MCP server"

  url {
    label = "us"
    url   = "https://mcp.us.internal.example.com/mcp"
  }
  url {
    label = "eu"
    url   = "https://mcp.eu.internal.example.com/mcp"
  }

  headers = {
    "X-Tenant" = "acme"
  }

  # Leave the tool blocks out to let the gateway detect the server's tools.
  tool {
    name        = "search_runbooks"
    type        = "read"
    description = "Search the runbooks"
  }
  tool {
    name        = "restart_service"
    type        = "destructive"
    description = "Restart a service"
  }

  auth {
    type              = "header"
    name              = "X-Api-Key"
    secret_wo         = var.internal_tools_api_key
    secret_wo_version = "1"
  }
}

# A server that needs an OAuth app you registered with it.
resource "hush_custom_mcp_application" "partner" {
  name         = "partner-crm"
  display_name = "Partner CRM"

  url {
    url = "https://mcp.partner.example.com/mcp"
  }

  scopes                   = ["contacts.read"]
  client_id                = "hush-gateway"
  client_secret_wo         = var.partner_client_secret
  client_secret_wo_version = "1"
}
