data "hush_mongodb_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_mongodb_access_credential.example.name
}

output "host" {
  value = data.hush_mongodb_access_credential.example.host
}

output "port" {
  value = data.hush_mongodb_access_credential.example.port
}

output "db_name" {
  value = data.hush_mongodb_access_credential.example.db_name
}

output "username" {
  value = data.hush_mongodb_access_credential.example.username
}

output "auth_source" {
  value = data.hush_mongodb_access_credential.example.auth_source
}

output "tls" {
  value = data.hush_mongodb_access_credential.example.tls
}
