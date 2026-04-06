# Create a GitLab dynamic access credential
resource "hush_gitlab_access_credential" "example" {
  name           = "prod-gitlab"
  description    = "Production GitLab credential"
  deployment_ids = [hush_deployment.example.id]
  resource_type  = "project"
  resource_id    = "12345"

  token_wo         = var.gitlab_token
  token_wo_version = 1
}
