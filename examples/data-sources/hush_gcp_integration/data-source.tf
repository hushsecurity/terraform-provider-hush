# Look up a GCP integration by ID
data "hush_gcp_integration" "by_id" {
  id = "integ-123456"
}

# Look up a GCP integration by name
data "hush_gcp_integration" "by_name" {
  name = "production-gcp"
}

# Output retrieved values
output "integration_status" {
  value       = data.hush_gcp_integration.by_name.status
  description = "Current status of the GCP integration"
}

output "integration_projects" {
  value       = data.hush_gcp_integration.by_name.project
  description = "Projects configured in the GCP integration"
}
