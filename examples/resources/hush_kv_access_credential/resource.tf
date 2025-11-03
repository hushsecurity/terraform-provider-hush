# Create a KV access credential
resource "hush_kv_access_credential" "example" {
  name          = "example-database-config"
  description   = "Database configuration for production environment"
  deployment_ids = ["dep-example123456789"]
  
  items {
    key   = "DB_HOST"
    value = "prod-database.example.com"
  }
  
  items {
    key   = "DB_PORT"
    value = "5432"
  }
  
  items {
    key   = "DB_USERNAME"
    value = "app_user"
  }
  
  items {
    key   = "DB_PASSWORD"
    value = "super_secret_password"
  }
}

# Output the credential information
output "credential_id" {
  value = hush_kv_access_credential.example.id
}

output "credential_type" {
  value = hush_kv_access_credential.example.type
}

output "available_keys" {
  value = hush_kv_access_credential.example.keys
}
