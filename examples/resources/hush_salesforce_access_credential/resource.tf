# Create a Salesforce dynamic access credential
resource "hush_salesforce_access_credential" "example" {
  name           = "prod-salesforce"
  description    = "Production Salesforce credential"
  deployment_ids = [hush_deployment.example.id]
  instance_url   = "https://myorg.salesforce.com"
  client_id      = "3MVG9..."

  client_secret_wo         = var.salesforce_client_secret
  client_secret_wo_version = 1
}
