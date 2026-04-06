data "hush_gitlab_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_gitlab_access_credential.example.name
}

output "base_url" {
  value = data.hush_gitlab_access_credential.example.base_url
}

output "resource_type" {
  value = data.hush_gitlab_access_credential.example.resource_type
}
