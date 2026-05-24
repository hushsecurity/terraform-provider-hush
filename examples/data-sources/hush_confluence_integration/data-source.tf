# Lookup by ID
data "hush_confluence_integration" "by_id" {
  id = "int-abc123"
}

# Lookup by name
data "hush_confluence_integration" "by_name" {
  name = "my-confluence"
}

output "confluence_status" {
  value = data.hush_confluence_integration.by_name.status
}
