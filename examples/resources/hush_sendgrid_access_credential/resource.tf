# Create a SendGrid dynamic access credential
resource "hush_sendgrid_access_credential" "example" {
  name           = "prod-sendgrid"
  description    = "Production SendGrid credential"
  deployment_ids = [hush_deployment.example.id]

  api_key_wo         = var.sendgrid_api_key
  api_key_wo_version = 1
}
