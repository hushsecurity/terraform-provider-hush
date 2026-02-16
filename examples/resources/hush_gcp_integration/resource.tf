# Create a GCP integration with specific projects and features
resource "hush_gcp_integration" "example" {
  name        = "production-gcp"
  description = "GCP integration for production environment"

  project {
    project_id = "my-gcp-project-001"
    enabled    = true
  }

  project {
    project_id = "my-gcp-project-002"
    enabled    = true
  }

  feature {
    name    = "iam"
    enabled = true
  }

  feature {
    name    = "secret_manager"
    enabled = true
  }
}

# Output the onboarding script to run in GCP Cloud Shell
output "onboarding_script" {
  value       = hush_gcp_integration.example.onboarding_script
  sensitive   = true
  description = "Run this script in GCP Cloud Shell to set up IAM permissions"
}

# Output the integration status
output "integration_status" {
  value       = hush_gcp_integration.example.status
  description = "Current status of the integration (pending or ok)"
}

# After running the onboarding script, complete the integration
# by adding the service_account_email from the script output:
#
# resource "hush_gcp_integration" "example" {
#   ...
#   service_account_email = "hush-<integration-id>@my-gcp-project.iam.gserviceaccount.com"
# }
