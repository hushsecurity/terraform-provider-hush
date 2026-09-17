# Create a PostgreSQL access privilege.
# With all_in_schema, object_names lists schemas rather than individual objects.
resource "hush_postgres_access_privilege" "example" {
  name        = "app-read-write"
  description = "Read/write access to application tables"

  grants {
    privileges    = ["SELECT", "INSERT", "UPDATE"]
    object_type   = "TABLE"
    object_names  = ["public"]
    all_in_schema = true
  }

  grants {
    privileges    = ["USAGE"]
    object_type   = "SEQUENCE"
    object_names  = ["public"]
    all_in_schema = true
  }
}
