# Data source to retrieve an existing plaintext access credential
data "hush_plaintext_access_credential" "example" {
  id = "acr_plaintext123456789"
}

# Output credential information
output "credential_name" {
  value = data.hush_plaintext_access_credential.example.name
}

output "credential_description" {
  value = data.hush_plaintext_access_credential.example.description
}

output "deployment_ids" {
  value = data.hush_plaintext_access_credential.example.deployment_ids
}

output "credential_type" {
  value = data.hush_plaintext_access_credential.example.type
}
