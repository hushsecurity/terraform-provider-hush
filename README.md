# Terraform Provider for Hush Security

The [Hush Terraform Provider](https://registry.terraform.io/providers/hushsecurity/hush/latest) allows managing Hush sensor deployments via the Hush Security API.

## Features

* **Resource Management**: Create, read, update, and delete Hush deployments, notification channels, notification configurations, and access credentials
* **Write-Only Secrets**: Enhanced security for plaintext credentials with write-only attributes (secrets not stored in state)
* **Data Sources**: Query existing deployments, notification channels, notification configurations, and access credentials by ID or name
* **Flexible Lookup**: Support for both ID-based and name-based lookups across all resources
* **Automatic Authentication**: OAuth2 client credentials flow with automatic token refresh
* **Comprehensive Examples**: Ready-to-use examples for all supported resources and data sources
* **Auto-generated Documentation**: Complete provider documentation in the `docs/` directory

## Resources

* `hush_deployment` - Manage Hush sensor deployments
* `hush_notification_channel` - Manage notification channels (email, webhook, Slack)
* `hush_notification_configuration` - Manage notification configurations and triggers
* `hush_plaintext_access_credential` - Manage plaintext access credentials (supports write-only secrets)
* `hush_kv_access_credential` - Manage key-value access credentials

## Data Sources

* `hush_deployment` - Read existing Hush deployments by ID or name
* `hush_notification_channel` - Read existing notification channels by ID or name
* `hush_notification_configuration` - Read existing notification configurations by ID or name
* `hush_plaintext_access_credential` - Read existing plaintext access credentials by ID
* `hush_kv_access_credential` - Read existing key-value access credentials by ID

## Requirements

* Terraform >= 1.3
* Go >= 1.24 (for development)
* Hush API credentials (`api_key_id` and `api_key_secret`)

## Authentication

The provider supports authentication via:

* `api_key_id` – Your Hush API key ID
* `api_key_secret` – Your Hush API key secret
* `realm` – (Optional) Hush realm (US, EU), defaults to US

These can also be set via environment variables:

```bash
export HUSH_API_KEY_ID=your_api_key_id
export HUSH_API_KEY_SECRET=your_api_key_secret
```

## Quick Start

```hcl
terraform {
  required_providers {
    hush = {
      source = "hushsecurity/hush"
    }
  }
}

provider "hush" {
  api_key_id     = var.api_key_id
  api_key_secret = var.api_key_secret
  realm          = "US"  # or "EU"
}

resource "hush_deployment" "example" {
  name        = "my-deployment"
  description = "Example deployment"
  env_type    = "prod"
  kind        = "k8s"
}

# Create a notification channel
resource "hush_notification_channel" "email_alerts" {
  name        = "security-alerts"
  description = "Email notifications for security alerts"
  enabled     = true

  email_config {
    address = "security@example.com"
  }
}

# Create a notification configuration
resource "hush_notification_configuration" "immediate_alerts" {
  name        = "New Secret at Risk Alerts"
  description = "Immediate notifications for new secrets at risk"
  enabled     = true
  
  channel_ids = [hush_notification_channel.email_alerts.id]
  aggregation = "short"
  trigger     = "new_nhi_at_risk"
}

# Look up deployment by name
data "hush_deployment" "by_name" {
  name = "my-deployment"
}
```

## Examples

Complete examples are available in the [`examples/`](./examples/) directory:

* [`examples/resources/hush_deployment/`](./examples/resources/hush_deployment/) - Creating deployments
* [`examples/resources/hush_notification_channel/`](./examples/resources/hush_notification_channel/) - Creating notification channels
* [`examples/resources/hush_notification_configuration/`](./examples/resources/hush_notification_configuration/) - Creating notification configurations
* [`examples/data-sources/hush_deployment/`](./examples/data-sources/hush_deployment/) - Reading deployments
* [`examples/data-sources/hush_notification_channel/`](./examples/data-sources/hush_notification_channel/) - Reading notification channels
* [`examples/data-sources/hush_notification_configuration/`](./examples/data-sources/hush_notification_configuration/) - Reading notification configurations

## Documentation

Auto-generated documentation is available in the [`docs/`](./docs/) directory and on the [Terraform Registry](https://registry.terraform.io/providers/hushsecurity/hush/latest/docs).

## Development

### Building the Provider

```bash
go build -o terraform-provider-hush
```

### Running Tests

```bash
make test
```

### Generating Documentation

```bash
make docs
```

### Local Development

For local testing, use the `.terraformrc` `dev_overrides` configuration to point Terraform to your local plugin build:

```hcl
provider_installation {
  dev_overrides {
    "hushsecurity/hush" = "/path/to/your/terraform-provider-hush"
  }
  direct {}
}
```

## License

[Apache 2.0](./LICENSE)
