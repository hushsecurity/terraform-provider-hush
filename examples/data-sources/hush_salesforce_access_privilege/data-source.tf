data "hush_salesforce_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_salesforce_access_privilege.example.name
}

output "run_as_user" {
  value = data.hush_salesforce_access_privilege.example.run_as_user
}

output "scopes" {
  value = data.hush_salesforce_access_privilege.example.scopes
}
