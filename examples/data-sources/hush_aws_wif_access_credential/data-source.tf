data "hush_aws_wif_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_aws_wif_access_credential.example.name
}

output "audience" {
  value = data.hush_aws_wif_access_credential.example.audience
}

output "issuer_url" {
  value = data.hush_aws_wif_access_credential.example.issuer_url
}
