# Create a Statsig dynamic access credential from a Console API key with the
# omni_read_write and can_access_keys scopes
resource "hush_statsig_access_credential" "example" {
  name           = "prod-statsig"
  description    = "Production Statsig project"
  deployment_ids = [hush_deployment.example.id]

  api_key_wo         = var.statsig_console_api_key
  api_key_wo_version = "1"
}
