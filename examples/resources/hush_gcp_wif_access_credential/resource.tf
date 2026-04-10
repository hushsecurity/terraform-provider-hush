# Create a GCP WIF access credential for workload identity federation
resource "hush_gcp_wif_access_credential" "example" {
  name                 = "prod-gcp-wif"
  description          = "Production GCP WIF credential"
  deployment_ids       = [hush_deployment.example.id]
  project_number       = "123456789012"
  pool_id              = "my-identity-pool"
  workload_provider_id = "my-workload-provider"
}
