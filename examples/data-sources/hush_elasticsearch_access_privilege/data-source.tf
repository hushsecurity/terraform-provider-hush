data "hush_elasticsearch_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_elasticsearch_access_privilege.example.name
}

output "grant" {
  value = data.hush_elasticsearch_access_privilege.example.grant
}
