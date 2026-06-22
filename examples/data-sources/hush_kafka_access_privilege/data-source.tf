data "hush_kafka_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_kafka_access_privilege.example.name
}

output "acls" {
  value = data.hush_kafka_access_privilege.example.acls
}
