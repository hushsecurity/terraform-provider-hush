data "hush_rabbitmq_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_rabbitmq_access_credential.example.name
}

output "host" {
  value = data.hush_rabbitmq_access_credential.example.host
}

output "port" {
  value = data.hush_rabbitmq_access_credential.example.port
}

output "management_port" {
  value = data.hush_rabbitmq_access_credential.example.management_port
}

output "username" {
  value = data.hush_rabbitmq_access_credential.example.username
}

output "vhost" {
  value = data.hush_rabbitmq_access_credential.example.vhost
}

output "tls" {
  value = data.hush_rabbitmq_access_credential.example.tls
}
