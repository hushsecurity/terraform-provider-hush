# Create an AWS access key dynamic access credential
resource "hush_aws_access_key_access_credential" "example" {
  name                   = "prod-aws-key"
  description            = "Production AWS access key credential"
  deployment_ids         = [hush_deployment.example.id]
  access_key_id_value    = "AKIAIOSFODNN7EXAMPLE"
  secret_access_key_wo   = var.aws_secret_access_key
}
