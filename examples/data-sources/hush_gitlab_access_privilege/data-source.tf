data "hush_gitlab_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_gitlab_access_privilege.example.name
}

output "access_level" {
  value = data.hush_gitlab_access_privilege.example.access_level
}

output "scopes" {
  value = data.hush_gitlab_access_privilege.example.scopes
}
