data "hush_github_app_access_credential" "example" {
  id = "acr-eu12345678"
}

output "installation_id" {
  value = data.hush_github_app_access_credential.example.installation_id
}
