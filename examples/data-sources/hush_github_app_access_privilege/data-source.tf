data "hush_github_app_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "permissions" {
  value = data.hush_github_app_access_privilege.example.permissions
}
