# Deliver the credentials as CLOUDINARY_URL, which Cloudinary's SDKs read
resource "hush_access_policy" "cloudinary_example" {
  name                 = "prod-cloudinary-policy"
  access_credential_id = hush_cloudinary_access_credential.example.id
  access_privilege_ids = [hush_cloudinary_access_privilege.example.id]
  deployment_ids       = [hush_deployment.example.id]

  attestation_criteria {
    type  = "k8s:ns"
    value = "production"
  }

  env_delivery_config {
    name = "CLOUDINARY_URL"
    key  = "cloudinary_url"
  }
}
