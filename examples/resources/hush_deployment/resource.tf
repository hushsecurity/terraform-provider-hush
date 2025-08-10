resource "hush_deployment" "example" {
  name        = "example-deployment"
  description = "Example deployment for testing"
  env_type    = "dev"
  kind        = "K8S"
}

output "deployment" {
  value = hush_deployment.example
}
