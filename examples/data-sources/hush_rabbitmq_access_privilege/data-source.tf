data "hush_rabbitmq_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_rabbitmq_access_privilege.example.name
}

output "permissions" {
  value = data.hush_rabbitmq_access_privilege.example.permissions
}

output "tags" {
  value = data.hush_rabbitmq_access_privilege.example.tags
}
