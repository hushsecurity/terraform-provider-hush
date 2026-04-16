# Create a SendGrid access privilege with mail send and template scopes
resource "hush_sendgrid_access_privilege" "example" {
  name        = "my-sendgrid-privilege"
  description = "Mail sending with template access"
  scopes      = ["mail.send", "templates.read", "templates.create"]
}
