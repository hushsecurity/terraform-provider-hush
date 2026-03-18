# Create a Twilio access privilege with restricted permissions
resource "hush_twilio_access_privilege" "example" {
  name            = "my-twilio-privilege"
  description     = "Twilio privilege with specific messaging permissions"
  permission_type = "Restricted"
  permissions = [
    "/twilio/messaging/messages/create",
    "/twilio/messaging/messages/read",
    "/twilio/messaging/messages/list",
  ]
}
