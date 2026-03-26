# Create a Snowflake access privilege with read-only access
resource "hush_snowflake_access_privilege" "example" {
  name        = "app-read-only"
  description = "Read-only access to application tables"

  grants {
    privileges    = ["USAGE"]
    resource_type = "database"
  }

  grants {
    privileges    = ["USAGE"]
    resource_type = "schema"
  }

  grants {
    privileges    = ["SELECT"]
    resource_type = "table"
  }

  grants {
    privileges    = ["USAGE"]
    resource_type = "warehouse"
  }
}
