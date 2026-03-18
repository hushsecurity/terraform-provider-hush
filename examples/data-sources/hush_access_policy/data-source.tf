data "hush_access_policy" "example" {
  id = "apl-eu12345678"
}

output "name" {
  value = data.hush_access_policy.example.name
}

output "enabled" {
  value = data.hush_access_policy.example.enabled
}

output "access_credential_id" {
  value = data.hush_access_policy.example.access_credential_id
}

output "access_privilege_ids" {
  value = data.hush_access_policy.example.access_privilege_ids
}

output "status" {
  value = data.hush_access_policy.example.status
}
