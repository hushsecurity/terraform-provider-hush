# Manage a Bitbucket integration
resource "hush_bitbucket_integration" "example" {
  name           = "my-bitbucket"
  workspace_slug = "my-workspace"
  token          = var.bitbucket_token
}

# With write-only token (recommended for production)
resource "hush_bitbucket_integration" "secure" {
  name             = "my-bitbucket-secure"
  workspace_slug   = "my-workspace"
  token_wo         = var.bitbucket_token
  token_wo_version = "v1"
}

# With optional description
resource "hush_bitbucket_integration" "with_description" {
  name           = "my-bitbucket-with-desc"
  description    = "Bitbucket integration for scanning repositories"
  workspace_slug = "my-workspace"
  token          = var.bitbucket_token
}
