# A REST API with two tools written by hand. To use the tools detected from
# spec_url instead, create the app without tool blocks, detect and review them
# in the console, then import it to copy them here, operation_id included.
# Create applications from it with
# app_catalog_id = hush_custom_api_application.billing.app_catalog_id.
resource "hush_custom_api_application" "billing" {
  name         = "billing"
  display_name = "Billing API"
  url          = "https://billing.internal.example.com"
  spec_url     = "https://billing.internal.example.com/openapi.json"

  auth {
    type              = "bearer"
    secret_wo         = var.billing_api_token
    secret_wo_version = "1"
  }

  tool {
    name        = "list_invoices"
    title       = "List invoices"
    description = "List the invoices of a customer"
    type        = "read"
    method      = "GET"
    path        = "/v1/invoices"
    input_schema = jsonencode({
      type = "object"
      properties = {
        query = {
          type       = "object"
          properties = { customer_id = { type = "string" } }
          required   = ["customer_id"]
        }
      }
    })
  }
  tool {
    name    = "void_invoice"
    title   = "Void invoice"
    type    = "destructive"
    method  = "POST"
    path    = "/v1/invoices/{id}/void"
    enabled = false
    input_schema = jsonencode({
      type = "object"
      properties = {
        path = { type = "object", properties = { id = { type = "string" } } }
      }
    })
  }
}
