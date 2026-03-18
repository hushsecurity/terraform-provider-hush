data "hush_redis_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_redis_access_credential.example.name
}

output "host" {
  value = data.hush_redis_access_credential.example.host
}

output "port" {
  value = data.hush_redis_access_credential.example.port
}

output "username" {
  value = data.hush_redis_access_credential.example.username
}

output "database" {
  value = data.hush_redis_access_credential.example.database
}

output "tls" {
  value = data.hush_redis_access_credential.example.tls
}
