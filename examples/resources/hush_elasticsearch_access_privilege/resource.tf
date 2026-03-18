# Create an Elasticsearch access privilege
resource "hush_elasticsearch_access_privilege" "example" {
  name        = "read-logs"
  description = "Read access to log indices"

  grant {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }
}
