# Let the gateway ask for consent in a Slack direct message, or through a
# browser sign-in. Declare one of these per deployment.
resource "hush_agw_consent_methods" "gateway" {
  deployment_id = hush_deployment.gateway.id
  methods       = ["push_slack", "oidc_callback"]
}
