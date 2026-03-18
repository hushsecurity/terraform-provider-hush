# Look up deployment by ID
data "hush_deployment" "by_id" {
  id = "dep-123abc"
}

# Look up deployment by name
data "hush_deployment" "by_name" {
  name = "my-web-service"
}

output "env_type" {
  value = data.hush_deployment.by_id.env_type
}

output "kind" {
  value = data.hush_deployment.by_id.kind
}

output "status" {
  value = data.hush_deployment.by_id.status
}
