# Manage an Artifactory integration
resource "hush_artifactory_integration" "example" {
  name    = "my-artifactory"
  org_url = "https://mycompany.jfrog.io"
  token   = var.artifactory_token
}

# With write-only token (recommended for production)
resource "hush_artifactory_integration" "secure" {
  name             = "my-artifactory-secure"
  org_url          = "https://mycompany.jfrog.io"
  token_wo         = var.artifactory_token
  token_wo_version = "v1"
}

# With optional description
resource "hush_artifactory_integration" "with_description" {
  name        = "my-artifactory-with-desc"
  description = "Artifactory integration for scanning repositories"
  org_url     = "https://mycompany.jfrog.io"
  token       = var.artifactory_token
}
