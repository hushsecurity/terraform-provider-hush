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

### Data Sources

- **[hush_deployment](data-sources/hush_deployment/)** - Read information about existing deployments
