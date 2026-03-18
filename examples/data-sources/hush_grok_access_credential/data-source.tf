data "hush_grok_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_grok_access_credential.example.name
}

output "team_id" {
  value = data.hush_grok_access_credential.example.team_id
}
