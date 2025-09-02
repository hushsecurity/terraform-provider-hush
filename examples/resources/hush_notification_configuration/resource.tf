# Look up predefined notification configuration
data "hush_notification_configuration" "secret_alerts" {
  trigger = "new_nhi_at_risk"
}

# Create notification channel
resource "hush_notification_channel" "email_alerts" {
  name        = "security-email"
  description = "Email notifications for security alerts"
  enabled     = true

  email_config {
    address = "security@example.com"
  }
}

# Configure predefined notification settings
resource "hush_notification_configuration" "alerts" {
  id      = data.hush_notification_configuration.secret_alerts.id
  enabled = true

  channel_ids = [
    hush_notification_channel.email_alerts.id
  ]
}

output "notification_configuration" {
  value = hush_notification_configuration.alerts
}
