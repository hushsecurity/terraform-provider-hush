# Lookup by ID
data "hush_sonatype_integration" "by_id" {
  id = "int-abc123"
}

# Lookup by name
data "hush_sonatype_integration" "by_name" {
  name = "my-sonatype"
}

output "sonatype_status" {
  value = data.hush_sonatype_integration.by_name.status
}

output "sonatype_url" {
  value = data.hush_sonatype_integration.by_name.org_url
}
