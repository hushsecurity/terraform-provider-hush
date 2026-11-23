data "hush_cloudflare_r2_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "permission" {
  value = data.hush_cloudflare_r2_access_privilege.example.permission
}

output "buckets" {
  value = data.hush_cloudflare_r2_access_privilege.example.buckets
}
