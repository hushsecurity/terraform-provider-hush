# Create a Bedrock dynamic access credential with explicit AWS keys
resource "hush_bedrock_access_credential" "example" {
  name           = "prod-bedrock"
  description    = "Production Bedrock credential"
  deployment_ids = [hush_deployment.example.id]
  region         = "us-east-1"
  access_key_id  = var.aws_access_key_id

  secret_access_key_wo         = var.aws_secret_access_key
  secret_access_key_wo_version = 1
}
