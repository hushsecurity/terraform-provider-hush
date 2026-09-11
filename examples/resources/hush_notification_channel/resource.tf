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
resource "hush_notification_channel" "soar" {
  name        = "soar-intake"
  description = "Structured events into the SOAR"
  enabled     = true

  webhook_config {
    url                  = "https://soar.internal:8443/api/events"
    onprem_deployment_id = "dep-abcdefghijk"
    payload_format       = "json"
    tls_verify           = false # internal endpoint on a private CA

    auth {
      type                  = "header"
      name                  = "Authorization"
      credential_wo         = var.soar_token
      credential_wo_version = "1"
    }
  }
}

# Splunk HTTP Event Collector. The host and a token are the whole destination:
# the collector path, the Authorization scheme and the event envelope are the
# API's. The token is proven against HEC when it is saved, so the channel needs
# no verification. Most Splunk Enterprise collectors sit on a private network
# on port 8088 and need the access bridge; Splunk Cloud is reached directly.
resource "hush_notification_channel" "splunk" {
  name        = "splunk-hec"
  description = "Notifications into the security index"
  enabled     = true

  splunk_config {
    url                  = "https://splunk.internal:8088"
    onprem_deployment_id = "dep-abcdefghijk"
    tls_verify           = false # HEC's self-signed certificate
    index                = "security"
    token_wo             = var.splunk_hec_token
    token_wo_version     = "1"
  }
}
