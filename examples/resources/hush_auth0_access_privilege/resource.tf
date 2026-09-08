# Grant a workload access to one or more of your Auth0-protected APIs.
#
# `audience` is the API identifier as registered in Auth0, matched by exact
# string equality. `scope` is optional: an API that does not use RBAC still
# needs the grant to exist before its client can obtain a token.
#
# The audience is not delivered to the workload, because a privilege may name
# several APIs while a token request takes exactly one. The application passes
# the audience it wants on each call.
resource "hush_auth0_access_privilege" "example" {
  name        = "orders-read-write"
  description = "Read and write orders"

  grants {
    audience = "https://api.acme.com"
    scope    = ["read:orders", "write:orders"]
  }
}
