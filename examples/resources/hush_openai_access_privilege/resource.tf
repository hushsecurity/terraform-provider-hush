# Create an OpenAI access privilege
resource "hush_openai_access_privilege" "example" {
  name            = "restricted-access"
  description     = "Restricted access with specific model permissions"
  permission_type = "Restricted"
}
