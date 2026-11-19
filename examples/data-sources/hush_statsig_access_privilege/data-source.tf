data "hush_statsig_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_statsig_access_privilege.example.name
}

output "key_type" {
  value = data.hush_statsig_access_privilege.example.key_type
}
