data "hush_gcp_wif_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_gcp_wif_access_credential.example.name
}

output "issuer_url" {
  value = data.hush_gcp_wif_access_credential.example.issuer_url
}

output "project_number" {
  value = data.hush_gcp_wif_access_credential.example.project_number
}
