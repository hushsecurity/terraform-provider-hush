# Issued per organization rather than per deployment, so there is nothing to
# look them up by.
data "hush_image_pull_credentials" "this" {}

output "hush_registry" {
  value = data.hush_image_pull_credentials.this.registry
}
