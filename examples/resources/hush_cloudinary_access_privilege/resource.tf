# Keys with Cloudinary's built-in Editor role in the product1 environment
resource "hush_cloudinary_access_privilege" "example" {
  name        = "cloudinary-product1-editor"
  description = "Upload, edit and delete assets in product1"
  cloud_name  = "product1"
  role_id     = "cld::role::prodenv::editor"
}

# Keys that can only upload, through a role the access manager creates
resource "hush_cloudinary_access_privilege" "upload_only" {
  name       = "cloudinary-product1-upload"
  cloud_name = "product1"
  system_policy_ids = [
    "cld::policy::global::folder_and_asset_management::create_asset",
    "cld::policy::global::upload_presets::manage",
  ]
}
