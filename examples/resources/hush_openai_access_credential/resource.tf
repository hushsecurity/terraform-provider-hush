# Create an OpenAI dynamic access credential
resource "hush_openai_access_credential" "example" {
  name               = "prod-openai"
  description        = "Production OpenAI API credential"
  deployment_ids     = [hush_deployment.example.id]
  api_key_wo         = var.openai_api_key
  api_key_wo_version = "1"
  project_id         = "proj_abc123"
}
