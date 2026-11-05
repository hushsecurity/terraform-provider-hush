# Create a GitHub App access credential. The access manager mints short-lived
# installation tokens from the App's private key.
resource "hush_github_app_access_credential" "example" {
  name            = "prod-github-app"
  description     = "Production GitHub App credential"
  deployment_ids  = [hush_deployment.example.id]
  client_id       = "Iv23liExampleClient1"
  installation_id = "12345678"

  private_key_wo         = var.github_app_private_key
  private_key_wo_version = "1"
}
