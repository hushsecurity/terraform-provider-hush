# Look up a predefined notification configuration
data "hush_notification_configuration" "digest" {
  trigger = "nhi_digest"
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

# Manage a predefined configuration: config_id, then what it may change
resource "hush_notification_configuration" "alerts" {
  config_id   = data.hush_notification_configuration.digest.id
  enabled     = true
  aggregation = "week"

  channel_ids = [
    hush_notification_channel.email_alerts.id
  ]
}

output "notification_configuration" {
  value = hush_notification_configuration.alerts
}

# Without config_id the resource creates the configuration: here, the agent
# gateway's activity log streamed to Datadog Logs from one deployment.
resource "hush_notification_channel" "datadog" {
  name = "datadog-logs"

  datadog_config {
    site               = "datadoghq.eu"
    api_key_wo         = var.datadog_api_key
    api_key_wo_version = "1"
  }
}

resource "hush_notification_configuration" "activity" {
  name        = "gateway-activity"
  trigger     = "agw_activity"
  enabled     = true
  channel_ids = [hush_notification_channel.datadog.id]

  dispatch_filters {
    name   = "deployment"
    op     = "eq"
    values = ["dep-abcdefghijk"]
  }
}
