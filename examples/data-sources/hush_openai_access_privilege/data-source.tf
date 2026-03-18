data "hush_openai_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_openai_access_privilege.example.name
}

output "permission_type" {
  value = data.hush_openai_access_privilege.example.permission_type
}

output "permissions" {
  value = data.hush_openai_access_privilege.example.permissions
}
