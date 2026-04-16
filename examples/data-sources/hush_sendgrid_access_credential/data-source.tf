data "hush_sendgrid_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_sendgrid_access_credential.example.name
}
