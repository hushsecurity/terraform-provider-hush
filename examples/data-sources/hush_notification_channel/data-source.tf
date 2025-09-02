# Look up notification channel by ID
data "hush_notification_channel" "by_id" {
  id = "nc-123abc"
}

# Look up notification channel by name
data "hush_notification_channel" "by_name" {
  name = "security-alerts-email"
}

output "channel_by_id" {
  value = data.hush_notification_channel.by_id
}

output "channel_by_name" {
  value = data.hush_notification_channel.by_name
}
