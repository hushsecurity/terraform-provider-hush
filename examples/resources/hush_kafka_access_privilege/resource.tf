# Create a Kafka access privilege granting read access to all topics and groups
resource "hush_kafka_access_privilege" "example" {
  name        = "my-kafka-consumer"
  description = "Read access to all topics and consumer groups"

  acls {
    resource_type   = "Topic"
    resource_name   = "*"
    pattern_type    = "LITERAL"
    operation       = "Read"
    permission_type = "ALLOW"
  }

  acls {
    resource_type   = "Group"
    resource_name   = "app-"
    pattern_type    = "PREFIXED"
    operation       = "Read"
    permission_type = "ALLOW"
    host            = "*"
  }
}
