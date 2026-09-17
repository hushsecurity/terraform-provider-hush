# Create an Apigee dynamic access credential
resource "hush_apigee_access_credential" "example" {
  name           = "prod-apigee"
  description    = "Production Apigee credential"
  deployment_ids = [hush_deployment.example.id]

  service_account_key_wo         = var.apigee_service_account_key
  service_account_key_wo_version = "1"
}
