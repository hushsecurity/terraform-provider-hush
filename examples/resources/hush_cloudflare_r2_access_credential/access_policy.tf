# Deliver the S3 credentials under the names every AWS SDK reads
resource "hush_access_policy" "cloudflare_r2_example" {
  name                 = "prod-r2-policy"
  access_credential_id = hush_cloudflare_r2_access_credential.example.id
  access_privilege_ids = [hush_cloudflare_r2_access_privilege.example.id]
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  env_delivery_config {
    name = "AWS_ACCESS_KEY_ID"
    key  = "access_key_id"
  }

  env_delivery_config {
    name = "AWS_SECRET_ACCESS_KEY"
    key  = "secret_access_key"
  }

  env_delivery_config {
    name = "AWS_ENDPOINT_URL_S3"
    key  = "endpoint"
  }
}
