data "hush_agent_authorization_server" "gateway" {
  deployment_id = hush_deployment.gateway.id
}

# The Audience an agent passes to sts:GetWebIdentityToken
output "agent_token_audience" {
  value = data.hush_agent_authorization_server.gateway.issuer
}
