data "hush_custom_api_application" "billing" {
  name = "billing"
}

output "billing_tool_names" {
  value = data.hush_custom_api_application.billing.tool[*].name
}
