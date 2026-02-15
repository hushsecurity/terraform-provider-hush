# Create a PostgreSQL access privilege
resource "hush_postgres_access_privilege" "example" {
  name        = "app-read-write"
  description = "Read/write access to application tables"

  grants {
    privileges = ["SELECT", "INSERT", "UPDATE"]
    object_type = "table"
    all_in_schema = "public"
  }

  grants {
    privileges = ["USAGE"]
    object_type = "sequence"
    all_in_schema = "public"
  }
}
