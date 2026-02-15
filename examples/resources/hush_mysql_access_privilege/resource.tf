# Create a MySQL access privilege
resource "hush_mysql_access_privilege" "example" {
  name        = "app-read-write"
  description = "Read/write access to application tables"

  grants {
    privileges    = ["SELECT", "INSERT", "UPDATE"]
    resource_type = "table"
    resource_names = ["users", "orders"]
  }
}
