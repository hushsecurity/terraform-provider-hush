# Scan a GitLab group
resource "hush_gitlab_integration" "group_example" {
  name         = "my-gitlab-group"
  group_id     = 12345
  token        = var.gitlab_token
  visibilities = ["private", "internal"]
}

# Scan a specific GitLab project
resource "hush_gitlab_integration" "project_example" {
  name       = "my-gitlab-project"
  project_id = 67890
  token      = var.gitlab_token
}

# Self-hosted GitLab with selected repos
resource "hush_gitlab_integration" "self_hosted" {
  name           = "my-self-hosted-gitlab"
  group_id       = 11111
  token          = var.gitlab_token
  base_url       = "https://gitlab.internal.company.com"
  selected_repos = ["repo-a", "repo-b"]
}

# Output the integration ID
output "gitlab_integration_id" {
  value       = hush_gitlab_integration.group_example.id
  description = "ID of the created GitLab integration"
}
