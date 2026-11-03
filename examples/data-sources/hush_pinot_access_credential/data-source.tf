data "hush_pinot_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_pinot_access_credential.example.name
}

output "host" {
  value = data.hush_pinot_access_credential.example.host
}

output "broker_url" {
  value = data.hush_pinot_access_credential.example.broker_url
}
