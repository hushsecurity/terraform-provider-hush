# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or are testable even if some parts are not relevant for the documentation.

- `provider/provider.tf` example file for the provider index page
- `data-sources/<DATASOURCE NAME>/data-source.tf` example file for the named data source page  
- `resources/<RESOURCE NAME>/resource.tf` example file for the named resource page

## Getting Started

Each example directory contains:
- Configuration files demonstrating usage
- Variable definitions where applicable
- Documentation specific to that example

### Prerequisites

1. [Terraform](https://terraform.io/downloads) installed (version 1.0+)
2. Hush API credentials (API key ID and secret)
3. Access to a Hush environment

### Basic Usage

1. Navigate to the desired example directory
2. Update configuration with your credentials and desired settings
3. Run Terraform:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Available Examples

### Resources

- **[hush_deployment](resources/hush_deployment/)** - Create and manage Hush deployments
- **[hush_notification_channel](resources/hush_notification_channel/)** - Create and manage notification channels (email, webhook, Slack)
- **[hush_notification_configuration](resources/hush_notification_configuration/)** - Create and manage notification configurations and triggers
- **[hush_plaintext_access_credential](resources/hush_plaintext_access_credential/)** - Create and manage plaintext access credentials for single secret values
- **[hush_kv_access_credential](resources/hush_kv_access_credential/)** - Create and manage key-value access credentials for multiple secret pairs

### Data Sources

- **[hush_deployment](data-sources/hush_deployment/)** - Read information about existing deployments
- **[hush_notification_channel](data-sources/hush_notification_channel/)** - Read information about existing notification channels
- **[hush_notification_configuration](data-sources/hush_notification_configuration/)** - Read information about existing notification configurations
- **[hush_plaintext_access_credential](data-sources/hush_plaintext_access_credential/)** - Read information about existing plaintext access credentials
- **[hush_kv_access_credential](data-sources/hush_kv_access_credential/)** - Read information about existing key-value access credentials
