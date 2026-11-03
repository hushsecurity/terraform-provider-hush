# Create a Pinot access privilege: queries on two tables through the broker
resource "hush_pinot_access_privilege" "example" {
  name        = "query-orders"
  description = "Query the orders tables"

  grant {
    tables = ["orders", "order_items"]
  }
}
