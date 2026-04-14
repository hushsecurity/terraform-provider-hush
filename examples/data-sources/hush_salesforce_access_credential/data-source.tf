data "hush_salesforce_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_salesforce_access_credential.example.name
}

output "instance_url" {
  value = data.hush_salesforce_access_credential.example.instance_url
}

output "client_id" {
  value = data.hush_salesforce_access_credential.example.client_id
}
