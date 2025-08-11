resource "hush_deployment" "example" {
  name        = "example-deployment"
  description = "Example deployment for testing"
  env_type    = "dev"
  kind        = "k8s"
}

output "deployment" {
  value = hush_deployment.example
}
