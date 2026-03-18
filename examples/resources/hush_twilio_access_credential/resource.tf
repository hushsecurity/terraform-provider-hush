# Create a Twilio dynamic access credential
resource "hush_twilio_access_credential" "example" {
  name           = "prod-twilio"
  description    = "Production Twilio credential"
  deployment_ids = [hush_deployment.example.id]
  account_sid    = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  api_key_sid    = var.twilio_api_key_sid

  api_key_secret_wo         = var.twilio_api_key_secret
  api_key_secret_wo_version = 1
}
