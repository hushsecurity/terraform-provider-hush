resource "hush_gcp_integration" "example" {
  name        = "my-gcp-integration"
  description = "GCP integration for scanning secrets"

  # Complete the integration with service account from onboarding module
  service_account_email = module.hush_gcp_onboard.service_account_email

  features {
    name    = "secret_manager"
    enabled = true
  }

  features {
    name    = "iam"
    enabled = true
  }

  projects {
    project_id = "my-gcp-project"
    enabled    = true
  }
}

# GCP onboarding module creates service account and IAM bindings
module "hush_gcp_onboard" {
  source  = "hushsecurity/onboard/gcp"
  version = ">= 1.1.0"

  hush_org_id                = "org-xxxxxxxxxxxx" # Your Hush org ID
  gcp_organization_id        = "123456789012"     # Your GCP org ID
  service_account_project_id = "my-gcp-project"   # Project for service account
  project_ids                = ["my-gcp-project"] # Projects to scan
}
