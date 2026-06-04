# Manage an Infisical integration
resource "hush_infisical_integration" "example" {
  name          = "my-infisical"
  base_url      = "https://app.infisical.com"
  client_id     = var.infisical_client_id
  client_secret = var.infisical_client_secret
}

# With write-only client secret (recommended for production)
resource "hush_infisical_integration" "secure" {
  name                     = "my-infisical-secure"
  base_url                 = "https://app.infisical.com"
  client_id                = var.infisical_client_id
  client_secret_wo         = var.infisical_client_secret
  client_secret_wo_version = "v1"
}

# With optional description
resource "hush_infisical_integration" "with_description" {
  name          = "my-infisical-with-desc"
  description   = "Infisical integration for secrets monitoring"
  base_url      = "https://app.infisical.com"
  client_id     = var.infisical_client_id
  client_secret = var.infisical_client_secret
}
