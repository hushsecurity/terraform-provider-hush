# Email notification channel example
resource "hush_notification_channel" "email_alerts" {
  name        = "security-alerts-email"
  description = "Email notifications for security alerts"
  enabled     = true
  type        = "email"

  config {
    address = "security@example.com"
  }
}

# Webhook notification channel example
resource "hush_notification_channel" "webhook_alerts" {
  name        = "security-webhook"
  description = "Webhook notifications for security events"
  enabled     = true
  type        = "webhook"

  config {
    url    = "https://example.com/webhook/security"
    method = "POST"
  }
}

# Slack notification channel example
resource "hush_notification_channel" "slack_alerts" {
  name        = "slack-security-channel"
  description = "Slack notifications for security team"
  enabled     = true
  type        = "slack"

  config {
    integration_id = "B1234567890"
    channel        = "security-alerts"
  }
}

output "email_channel_id" {
  value = hush_notification_channel.email_alerts.id
}

output "webhook_channel_id" {
  value = hush_notification_channel.webhook_alerts.id
}

output "slack_channel_id" {
  value = hush_notification_channel.slack_alerts.id
}
