# Create a plaintext access credential with standard secret storage
resource "hush_plaintext_access_credential" "example" {
  name           = "example-api-key"
  description    = "API key for external service integration"
  deployment_ids = ["dep-example123456789"]
  secret         = "sk-1234567890abcdef"
}

# Create a plaintext access credential with write-only secret (enhanced security)
# The secret value will not be stored in Terraform state
resource "hush_plaintext_access_credential" "example_write_only" {
  name              = "example-api-key-secure"
  description       = "API key with enhanced security (not stored in state)"
  deployment_ids    = ["dep-example123456789"]
  secret_wo         = "sk-1234567890abcdef"
  secret_wo_version = "v1" # Change this value when rotating the secret
}

# Output the credential ID for reference
output "credential_id" {
  value = hush_plaintext_access_credential.example.id
}

output "credential_type" {
  value = hush_plaintext_access_credential.example.type
}
