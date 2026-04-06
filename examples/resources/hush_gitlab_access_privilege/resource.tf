# Create a GitLab access privilege with developer permissions
resource "hush_gitlab_access_privilege" "example" {
  name         = "my-gitlab-privilege"
  description  = "Developer access with full API scope"
  access_level = "Developer"
  scopes       = ["api"]
}
