# Lookup by ID
data "hush_jira_integration" "by_id" {
  id = "int-123456"
}

# Lookup by name
data "hush_jira_integration" "by_name" {
  name = "my-jira"
}
