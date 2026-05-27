# Create an Azure WIF access credential for workload identity federation
resource "hush_azure_wif_access_credential" "example" {
  name           = "prod-azure-wif"
  description    = "Production Azure WIF credential"
  deployment_ids = [hush_deployment.example.id]
}
