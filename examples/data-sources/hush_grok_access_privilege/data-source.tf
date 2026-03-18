data "hush_grok_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_grok_access_privilege.example.name
}

output "endpoints" {
  value = data.hush_grok_access_privilege.example.endpoints
}

output "models" {
  value = data.hush_grok_access_privilege.example.models
}
