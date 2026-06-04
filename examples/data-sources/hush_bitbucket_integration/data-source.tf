# Lookup by ID
data "hush_bitbucket_integration" "by_id" {
  id = "int-abc123"
}

# Lookup by name
data "hush_bitbucket_integration" "by_name" {
  name = "my-bitbucket"
}

output "bitbucket_status" {
  value = data.hush_bitbucket_integration.by_name.status
}

output "bitbucket_workspace" {
  value = data.hush_bitbucket_integration.by_name.workspace_slug
}
