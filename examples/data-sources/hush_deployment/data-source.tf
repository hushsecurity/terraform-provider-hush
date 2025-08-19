# Look up deployment by ID
data "hush_deployment" "by_id" {
  id = "dep-123abc"
}

# Look up deployment by name
data "hush_deployment" "by_name" {
  name = "my-web-service"
}
