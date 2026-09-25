data "hush_mcp_catalog_entry" "datadog" {
  app_catalog_id = "datadog"
}

# The labels an application's url_label may name, such as ["US1", "EU"].
output "datadog_url_labels" {
  value = data.hush_mcp_catalog_entry.datadog.urls[*].label
}

# The destructive tools, which the gateway blocks by default.
output "datadog_destructive_tools" {
  value = [for tool in data.hush_mcp_catalog_entry.datadog.tools : tool.name if tool.type == "destructive"]
}
