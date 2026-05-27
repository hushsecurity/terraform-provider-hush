# Create a MongoDB Atlas access privilege
resource "hush_mongodb_atlas_access_privilege" "example" {
  name        = "app-read-write"
  description = "Read/write access to application collections"

  grants {
    privileges     = ["FIND", "INSERT", "UPDATE", "REMOVE"]
    resource_type  = "collection"
    resource_names = ["users", "orders"]
  }

  grants {
    privileges    = ["all"]
    resource_type = "database"
  }
}
