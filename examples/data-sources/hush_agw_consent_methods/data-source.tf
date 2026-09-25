data "hush_agw_consent_methods" "gateway" {
  deployment_id = hush_deployment.gateway.id
}

output "gateway_consent_methods" {
  value = data.hush_agw_consent_methods.gateway.methods
}
