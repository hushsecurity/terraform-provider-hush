# Create an Azure app dynamic access credential
resource "hush_azure_app_access_credential" "example" {
  name              = "prod-azure-app"
  description       = "Production Azure app credential"
  deployment_ids    = [hush_deployment.example.id]
  tenant_id         = "12345678-1234-1234-1234-123456789012"
  client_id         = "87654321-4321-4321-4321-210987654321"
  client_secret_wo  = var.azure_client_secret
}
