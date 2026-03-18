data "hush_apigee_access_privilege" "example" {
  id = "apr-eu12345678"
}

output "name" {
  value = data.hush_apigee_access_privilege.example.name
}

output "developer_email" {
  value = data.hush_apigee_access_privilege.example.developer_email
}

output "project_id" {
  value = data.hush_apigee_access_privilege.example.project_id
}

output "api_products" {
  value = data.hush_apigee_access_privilege.example.api_products
}
