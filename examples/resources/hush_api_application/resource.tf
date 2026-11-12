# The billing API on the gateway of one deployment, for the finance team, asking
# the user before listing invoices.
resource "hush_api_application" "billing" {
  app_catalog_id = hush_custom_api_application.billing.app_catalog_id
  display_name   = "Billing"
  deployment_ids = [hush_deployment.gateway.id]

  tool_operation {
    name      = "list_invoices"
    operation = "user_consent"
  }

  assignment {
    condition {
      match {
        source   = "user"
        property = "groups"
        op       = "in"
        values   = ["finance"]
      }
    }
  }
}
