# Create a Redis access privilege
resource "hush_redis_access_privilege" "example" {
  name        = "read-only"
  description = "Read-only Redis access"

  grants {
    type   = "category"
    action = "include"
    name   = "read"
  }

  grants {
    type   = "category"
    action = "exclude"
    name   = "write"
  }

  keys     = ["*"]
  channels = ["*"]
}
