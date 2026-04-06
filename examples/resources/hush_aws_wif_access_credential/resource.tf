# Create an AWS WIF access credential for workload identity federation
resource "hush_aws_wif_access_credential" "example" {
  name           = "prod-aws-wif"
  description    = "Production AWS WIF credential"
  deployment_ids = [hush_deployment.example.id]
}
