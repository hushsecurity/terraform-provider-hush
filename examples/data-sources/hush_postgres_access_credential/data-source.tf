data "hush_postgres_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_postgres_access_credential.example.name
}

output "host" {
  value = data.hush_postgres_access_credential.example.host
}

output "port" {
  value = data.hush_postgres_access_credential.example.port
}

output "db_name" {
  value = data.hush_postgres_access_credential.example.db_name
}

output "username" {
  value = data.hush_postgres_access_credential.example.username
}

output "ssl_mode" {
  value = data.hush_postgres_access_credential.example.ssl_mode
}
