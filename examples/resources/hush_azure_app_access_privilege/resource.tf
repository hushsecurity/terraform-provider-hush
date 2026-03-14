# Create an Azure app access privilege
resource "hush_azure_app_access_privilege" "example" {
  name        = "azure-storage-access"
  description = "Azure storage access privilege"
  roles       = ["Storage Blob Data Reader"]
}
