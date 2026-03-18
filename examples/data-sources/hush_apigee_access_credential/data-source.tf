data "hush_apigee_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_apigee_access_credential.example.name
}

output "has_provider_credentials" {
  value = data.hush_apigee_access_credential.example.has_provider_credentials
}
