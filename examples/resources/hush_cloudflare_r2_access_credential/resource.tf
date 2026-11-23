# Create a Cloudflare R2 dynamic access credential
resource "hush_cloudflare_r2_access_credential" "example" {
  name           = "prod-r2"
  description    = "Production Cloudflare R2 credential"
  deployment_ids = [hush_deployment.example.id]
  account_id     = "0123456789abcdef0123456789abcdef"

  api_token_wo         = var.cloudflare_root_api_token
  api_token_wo_version = "1"
}
