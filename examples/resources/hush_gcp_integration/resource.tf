resource "hush_gcp_integration" "example" {
  name        = "my-gcp-integration"
  description = "GCP integration for scanning secrets"

  features {
    name    = "secret_manager"
    enabled = true
  }

  features {
    name    = "iam"
    enabled = true
  }

  projects {
    project_id = "my-gcp-project-123"
    enabled    = true
  }

  # After running the onboarding script, uncomment to complete the integration:
  # service_account_email = "hush-sa@my-gcp-project-123.iam.gserviceaccount.com"
}
