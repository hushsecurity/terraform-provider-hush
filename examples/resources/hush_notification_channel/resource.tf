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

# A webhook delivered into a customer network through an on-prem deployment's
# access bridge, carrying the credential the endpoint requires. credential_wo
# is write-only: it never reaches Terraform state, and re-sending it is driven
# by credential_wo_version.
resource "hush_notification_channel" "splunk_hec" {
  name        = "splunk-hec"
  description = "Structured events into Splunk HEC"
  enabled     = true

  webhook_config {
    url                  = "https://splunk.internal:8088/services/collector"
    onprem_deployment_id = "dep-abcdefghijk"
    payload_format       = "json"
    tls_verify           = false # internal endpoint on a private CA

    auth {
      type                  = "header"
      name                  = "Authorization"
      credential_wo         = var.splunk_hec_token
      credential_wo_version = "1"
    }
  }
}
