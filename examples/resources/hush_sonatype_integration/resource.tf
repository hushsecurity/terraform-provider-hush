# Manage a Sonatype integration
resource "hush_sonatype_integration" "example" {
  name    = "my-sonatype"
  org_url = "https://my-company.sonatype.com"
  user    = "admin@mycompany.com"
  api_key = var.sonatype_api_key
}

# With write-only API key (recommended for production)
resource "hush_sonatype_integration" "secure" {
  name              = "my-sonatype-secure"
  org_url           = "https://my-company.sonatype.com"
  user              = "admin@mycompany.com"
  api_key_wo        = var.sonatype_api_key
  api_key_wo_version = "v1"
}

# With optional description
resource "hush_sonatype_integration" "with_description" {
  name        = "my-sonatype-with-desc"
  description = "Sonatype integration for scanning repositories"
  org_url     = "https://my-company.sonatype.com"
  user        = "admin@mycompany.com"
  api_key     = var.sonatype_api_key
}
