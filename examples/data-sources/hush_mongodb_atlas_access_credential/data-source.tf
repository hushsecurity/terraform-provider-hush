data "hush_mongodb_atlas_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_mongodb_atlas_access_credential.example.name
}

output "host" {
  value = data.hush_mongodb_atlas_access_credential.example.host
}

output "group_id" {
  value = data.hush_mongodb_atlas_access_credential.example.group_id
}

output "db_name" {
  value = data.hush_mongodb_atlas_access_credential.example.db_name
}
