# Create notification channels first
resource "hush_notification_channel" "email_alerts" {
  name        = "security-email"
  description = "Email notifications for security alerts"
  enabled     = true
  type        = "email"

  config {
    address = "security@example.com"
  }
}

resource "hush_notification_channel" "slack_alerts" {
  name        = "slack-security"
  description = "Slack notifications for security team"
  enabled     = true
  type        = "slack"

  config {
    integration_id = "B1234567890"
    channel        = "security-alerts"
  }
}

# Immediate alert configuration for new secrets at risk
resource "hush_notification_configuration" "immediate_alerts" {
  name        = "New Secret at Risk Alerts"
  description = "Immediate notifications when new secrets are detected at risk"
  enabled     = true

  channel_ids = [
    hush_notification_channel.email_alerts.id,
    hush_notification_channel.slack_alerts.id
  ]

  aggregation = "short"
  trigger     = "new_nhi_at_risk"
}

# Weekly digest configuration
resource "hush_notification_configuration" "weekly_digest" {
  name        = "Weekly Security Digest"
  description = "Weekly summary of all secrets at risk"
  enabled     = true

  channel_ids = [
    hush_notification_channel.email_alerts.id
  ]

  aggregation = "week"
  trigger     = "nhi_digest"
}

output "immediate_alerts_config_id" {
  value = hush_notification_configuration.immediate_alerts.id
}

output "weekly_digest_config_id" {
  value = hush_notification_configuration.weekly_digest.id
}
