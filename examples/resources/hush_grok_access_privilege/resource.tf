# Create a Grok access privilege with specific endpoints and models
resource "hush_grok_access_privilege" "example" {
  name        = "chat-only"
  description = "Access to Chat endpoint with grok-2 model"
  endpoints   = ["Chat"]
  models      = ["grok-2"]
}
