# Create a GCP SA dynamic access credential
resource "hush_gcp_sa_access_credential" "example" {
  name                   = "prod-gcp-sa"
  description            = "Production GCP SA credential"
  deployment_ids         = [hush_deployment.example.id]
  service_account_key_wo = var.gcp_service_account_key
}
