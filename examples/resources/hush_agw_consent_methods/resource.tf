# Let the gateway ask for consent in the chat itself, in a Slack direct
# message, or through a browser sign-in. A client that does not support MCP
# elicitation is offered the other methods. Declare one of these per deployment.
resource "hush_agw_consent_methods" "gateway" {
  deployment_id = hush_deployment.gateway.id
  methods       = ["mcp_elicitation", "push_slack", "oidc_callback"]
}
