# Look up notification configuration by ID
data "hush_notification_configuration" "by_id" {
  id = "nconf-123abc"
}

# Look up notification configuration by name
data "hush_notification_configuration" "by_name" {
  name = "Weekly Security Digest"
}

output "config_by_id" {
  value = data.hush_notification_configuration.by_id
}

output "config_by_name" {
  value = data.hush_notification_configuration.by_name
}
