data "hush_deployment" "existing" {
  id = "dep-123abc"
}

output "deployment" {
  value = data.hush_deployment.existing
}
