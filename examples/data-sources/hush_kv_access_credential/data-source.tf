# Data source to retrieve an existing KV access credential
data "hush_kv_access_credential" "example" {
  id = "acr_kv123456789"
}

# Output credential information
output "credential_name" {
  value = data.hush_kv_access_credential.example.name
}

output "credential_description" {
  value = data.hush_kv_access_credential.example.description
}

output "deployment_ids" {
  value = data.hush_kv_access_credential.example.deployment_ids
}

output "available_keys" {
  value = data.hush_kv_access_credential.example.keys
}

output "credential_type" {
  value = data.hush_kv_access_credential.example.type
}
