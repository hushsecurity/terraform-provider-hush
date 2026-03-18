data "hush_postgres_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_postgres_access_privilege.example.name
}

output "grants" {
  value = data.hush_postgres_access_privilege.example.grants
}
