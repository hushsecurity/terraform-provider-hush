# Create a Temporal Cloud dynamic access credential
resource "hush_temporal_cloud_access_credential" "example" {
  name               = "prod-temporal-cloud"
  description        = "Production Temporal Cloud API credential"
  deployment_ids     = [hush_deployment.example.id]
  api_key_wo         = var.temporal_cloud_api_key
  api_key_wo_version = 1
}
