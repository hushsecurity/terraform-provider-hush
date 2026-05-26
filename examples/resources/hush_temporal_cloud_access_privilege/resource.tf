# Create a Temporal Cloud access privilege
resource "hush_temporal_cloud_access_privilege" "example" {
  name        = "prod-namespace-read"
  description = "Read access to production namespaces"

  grants {
    namespace  = "prod.acct1"
    permission = "read"
  }

  grants {
    namespace  = "staging.acct1"
    permission = "write"
  }
}
