# Lookup by ID
data "hush_infisical_integration" "by_id" {
  id = "int-abc123"
}

# Lookup by name
data "hush_infisical_integration" "by_name" {
  name = "my-infisical"
}

output "infisical_status" {
  value = data.hush_infisical_integration.by_name.status
}

output "infisical_url" {
  value = data.hush_infisical_integration.by_name.base_url
}
