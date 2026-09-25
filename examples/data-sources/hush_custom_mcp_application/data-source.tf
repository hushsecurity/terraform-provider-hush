data "hush_custom_mcp_application" "internal_tools" {
  name = "internal-tools"
}

# The id applications are created from.
output "internal_tools_app_catalog_id" {
  value = data.hush_custom_mcp_application.internal_tools.app_catalog_id
}
