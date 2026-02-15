# Create a MongoDB access privilege
resource "hush_mongodb_access_privilege" "example" {
  name        = "app-read-write"
  description = "Read/write access to application collections"

  grants {
    privileges    = ["find", "insert", "update"]
    resource_type = "collection"
    resource_names = ["users", "orders"]
  }
}
