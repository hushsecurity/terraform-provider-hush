data "hush_sendgrid_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_sendgrid_access_privilege.example.name
}

output "scopes" {
  value = data.hush_sendgrid_access_privilege.example.scopes
}
