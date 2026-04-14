# Create a Salesforce access privilege with API and refresh token scopes
resource "hush_salesforce_access_privilege" "example" {
  name        = "my-salesforce-privilege"
  description = "API access with refresh token"
  run_as_user = "admin@myorg.salesforce.com"
  scopes      = ["Api", "RefreshToken"]
}
