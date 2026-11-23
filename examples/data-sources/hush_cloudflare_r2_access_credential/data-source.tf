data "hush_cloudflare_r2_access_credential" "example" {
  id = "acr-eu12345678"
}

output "account_id" {
  value = data.hush_cloudflare_r2_access_credential.example.account_id
}
