# Lookup by ID
data "hush_gitlab_integration" "by_id" {
  id = "int-123456"
}

# Lookup by name
data "hush_gitlab_integration" "by_name" {
  name = "my-gitlab-group"
}

# Output retrieved values
output "gitlab_integration_status" {
  value       = data.hush_gitlab_integration.by_name.status
  description = "Status of the GitLab integration"
}
