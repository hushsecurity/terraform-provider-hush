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

# A migration user that hands the objects it creates to an owner role with
# ALTER ... OWNER TO app_owner. The root credential's user needs ADMIN OPTION
# on app_owner (GRANT app_owner TO <root> WITH ADMIN OPTION).
resource "hush_postgres_access_privilege" "migrations" {
  name      = "app-migrations"
  member_of = ["app_owner"]

  grants {
    privileges   = ["USAGE", "CREATE"]
    object_type  = "SCHEMA"
    object_names = ["public"]
  }
}
