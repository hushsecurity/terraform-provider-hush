data "hush_snowflake_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_snowflake_access_credential.example.name
}

output "account" {
  value = data.hush_snowflake_access_credential.example.account
}

output "warehouse" {
  value = data.hush_snowflake_access_credential.example.warehouse
}

output "database" {
  value = data.hush_snowflake_access_credential.example.database
}

output "username" {
  value = data.hush_snowflake_access_credential.example.username
}

output "auth_method" {
  value = data.hush_snowflake_access_credential.example.auth_method
}
