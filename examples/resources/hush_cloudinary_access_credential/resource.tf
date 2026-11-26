# Create a Cloudinary dynamic access credential
resource "hush_cloudinary_access_credential" "example" {
  name           = "prod-cloudinary"
  description    = "Production Cloudinary credential"
  deployment_ids = [hush_deployment.example.id]
  account_id     = "abcde1fg-2345-6789-0abc-d1efg2hi3jkl"

  api_key_wo            = var.cloudinary_account_management_key
  api_key_wo_version    = "1"
  api_secret_wo         = var.cloudinary_account_management_secret
  api_secret_wo_version = "1"
}
