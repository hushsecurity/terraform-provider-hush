# Create a Grok dynamic access credential
resource "hush_grok_access_credential" "example" {
  name           = "prod-grok"
  description    = "Production Grok API credential"
  deployment_ids = [hush_deployment.example.id]
  api_key_wo     = var.grok_api_key
  team_id        = "team-abc123"
}
