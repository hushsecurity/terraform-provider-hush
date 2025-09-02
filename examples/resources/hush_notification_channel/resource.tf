resource "hush_notification_channel" "team_emails" {
  name        = "security-team-emails"
  description = "Email notifications"
  enabled     = true

  email_config {
    address = "security-lead@example.com"
  }
  
  email_config {
    address = "security-engineer@example.com"
  }
}

output "notification_channel" {
  value = hush_notification_channel.team_emails
}
