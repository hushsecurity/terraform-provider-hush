# Create a Gemini dynamic access credential
resource "hush_gemini_access_credential" "example" {
  name                    = "prod-gemini"
  description             = "Production Gemini API credential"
  deployment_ids          = [hush_deployment.example.id]
  project_id              = "my-gcp-project"
  service_account_key_wo  = var.gemini_service_account_key
}
