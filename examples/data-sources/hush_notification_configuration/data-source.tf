# Look up notification configuration by ID
data "hush_notification_configuration" "by_id" {
  id = "ntf-123abc"
}

# Look up notification configuration by trigger
data "hush_notification_configuration" "by_trigger" {
  trigger = "new_nhi_at_risk"
}

# Look up notification configuration by name
data "hush_notification_configuration" "by_name" {
  name = "Weekly Security Digest"
}

output "config_by_id" {
  value = data.hush_notification_configuration.by_id
}

output "config_by_trigger" {
  value = data.hush_notification_configuration.by_trigger
}

output "config_by_name" {
  value = data.hush_notification_configuration.by_name
}
