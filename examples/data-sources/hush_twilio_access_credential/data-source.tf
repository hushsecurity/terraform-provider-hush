data "hush_twilio_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_twilio_access_credential.example.name
}

output "account_sid" {
  value = data.hush_twilio_access_credential.example.account_sid
}

output "api_key_sid" {
  value = data.hush_twilio_access_credential.example.api_key_sid
}
