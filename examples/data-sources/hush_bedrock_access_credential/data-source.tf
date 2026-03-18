data "hush_bedrock_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_bedrock_access_credential.example.name
}

output "region" {
  value = data.hush_bedrock_access_credential.example.region
}

output "access_key_id" {
  value = data.hush_bedrock_access_credential.example.access_key_id
}

output "has_provider_credentials" {
  value = data.hush_bedrock_access_credential.example.has_provider_credentials
}
