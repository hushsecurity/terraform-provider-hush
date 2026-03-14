# Create a GCP SA access privilege
resource "hush_gcp_sa_access_privilege" "example" {
  name        = "gcp-storage-access"
  description = "GCP storage access privilege"
  roles       = ["roles/storage.objectViewer"]
}
