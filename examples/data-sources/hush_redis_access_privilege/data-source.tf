data "hush_redis_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_redis_access_privilege.example.name
}

output "grants" {
  value = data.hush_redis_access_privilege.example.grants
}

output "keys" {
  value = data.hush_redis_access_privilege.example.keys
}

output "channels" {
  value = data.hush_redis_access_privilege.example.channels
}
