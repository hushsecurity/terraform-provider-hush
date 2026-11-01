# Narrow tokens to read access on two repositories
resource "hush_github_app_access_privilege" "example" {
  name         = "read-app-repos"
  description  = "Read the app repositories"
  repositories = ["frontend", "backend"]
  permissions = {
    contents = "read"
    metadata = "read"
  }
}
