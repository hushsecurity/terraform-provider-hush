# Create an Apigee access privilege with an existing app
resource "hush_apigee_access_privilege" "example" {
  name            = "my-apigee-privilege"
  description     = "Apigee privilege with existing app"
  developer_email = "developer@example.com"
  project_id      = "my-gcp-project"
  api_products    = ["my-api-product"]
  app_name        = "my-existing-app"
}
