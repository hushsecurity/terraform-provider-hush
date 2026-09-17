# Create an Azure app access privilege that provisions a new app registration
# and assigns it roles. Roles are a nested block of the app_config block, each
# naming a role and the scope it is granted at.
resource "hush_azure_app_access_privilege" "example" {
  name        = "azure-storage-access"
  description = "Azure storage access privilege"

  app_config {
    display_name = "hush-storage-reader"

    roles {
      name  = "Storage Blob Data Reader"
      scope = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg"
    }
  }
}

# Bind the privilege to an app registration managed outside Hush instead of
# provisioning one. Exactly one of app_id and app_config must be set.
resource "hush_azure_app_access_privilege" "existing_app" {
  name        = "azure-existing-app"
  description = "Privilege bound to an existing app registration"
  app_id      = "11111111-1111-1111-1111-111111111111"
}
