data "hush_cloudinary_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "cloud_name" {
  value = data.hush_cloudinary_access_privilege.example.cloud_name
}

output "role_id" {
  value = data.hush_cloudinary_access_privilege.example.role_id
}
