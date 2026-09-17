# Create a GCP SA access privilege that provisions a new service account and
# grants it project roles.
resource "hush_gcp_sa_access_privilege" "example" {
  name        = "gcp-storage-access"
  description = "GCP storage access privilege"
  project_id  = "my-gcp-project"

  sa_config {
    display_name = "hush-storage-reader"
    roles        = ["roles/storage.objectViewer"]
  }
}

# Bind the privilege to a service account managed outside Hush instead of
# provisioning one. Exactly one of sa_email and sa_config must be set.
resource "hush_gcp_sa_access_privilege" "existing_sa" {
  name        = "gcp-existing-sa"
  description = "Privilege bound to an existing service account"
  project_id  = "my-gcp-project"
  sa_email    = "app-reader@my-gcp-project.iam.gserviceaccount.com"
}
