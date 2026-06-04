# Lookup by ID
data "hush_artifactory_integration" "by_id" {
  id = "int-abc123"
}

# Lookup by name
data "hush_artifactory_integration" "by_name" {
  name = "my-artifactory"
}

output "artifactory_status" {
  value = data.hush_artifactory_integration.by_name.status
}

output "artifactory_url" {
  value = data.hush_artifactory_integration.by_name.org_url
}
