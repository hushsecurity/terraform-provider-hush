# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.3.0] - 2025-11-13

### Added

* **Access Credentials**: Secure secrets management for deployments
  * `hush_plaintext_access_credential` resource for managing single secret values
  * `hush_plaintext_access_credential` data source for importing existing credentials
  * `hush_kv_access_credential` resource for managing key-value secret pairs
  * `hush_kv_access_credential` data source for importing existing credentials
  * **Write-Only Secrets**: Enhanced security for plaintext credentials with `secret_wo` and `secret_wo_version` attributes
    * Secrets stored using `secret_wo` are never persisted in Terraform state
    * Version-based secret rotation with automatic resource recreation
    * Backward compatible with standard `secret` attribute for traditional state storage

## [1.2.0] - 2025-09-10

### Added

* **Notification Channels**: Support for email, webhook, and Slack notification channels
  * `hush_notification_channel` resource for managing notification delivery channels
  * `hush_notification_channel` data source for importing existing channels
  * Multiple email address support for team notifications

* **Notification Configurations**: Automated security alert management  
  * `hush_notification_configuration` resource for predefined security workflows
  * `hush_notification_configuration` data source with lookup by name, ID, or trigger type
  * Support for "New Secret At Risk" and "Secrets at Risk Digest" configurations

## [1.1.0] - 2025-08-20

### Added

* **Deployment Data Source**: Added support for looking up deployments by `name` in addition to `id`

## [1.0.3] - 2025-08-13

### Changed

* **Error Handling**: Improved error handling throughout the provider for better user experience
  * Centralized error handling with new `APIError` struct providing clear, actionable error messages
  * Simplified error propagation with clean type assertion patterns
  * Improved 404 handling for graceful resource state management

### Removed

* **Logging**: Eliminated tflog imports and verbose error logging in provider layer
  * Streamlined operations for cleaner, more reliable execution

## [1.0.2] - 2025-08-11

### Fixed

* **Resource `hush_deployment`**: Fixed `kind` field validation to accept lowercase values
  * Changed validation to accept `k8s`, `ecs`, `serverless` (lowercase) to match API requirements
  * Previously required uppercase values (`K8S`, `ECS`, `SERVERLESS`) which caused API errors
* **Resource `hush_deployment`**: Added validation for `env_type` field
  * Now validates that `env_type` accepts only `dev` or `prod` values as required by the API
  * Prevents runtime errors by catching invalid values at plan time

## [1.0.1] - 2025-08-10

### Added

* **Resource `hush_deployment`**: Add support for `kind` field
  * New required `kind` field with validation for values: `K8S`, `ECS`, `SERVERLESS`
  * Supports both create and update operations for deployment kind

## [1.0.0] - 2025-07-27

### Added

* **Initial Release**: First public release of the `hushsecurity/hush` Terraform provider
* **Authentication**: OAuth2 client credentials flow with automatic token refresh and thread-safe implementation
* **Resource `hush_deployment`**: Create, read, update, and delete Hush sensor deployments
  * Supports `name`, `description`, and `env_type` fields
  * Automatically fetches deployment credentials (token and password) after creation
  * Full CRUD operations with proper state management
* **Data Source `hush_deployment`**: Read existing Hush deployments by ID
* **Comprehensive Examples**: Complete example configurations for both resources and data sources
  * Resource examples in `examples/resources/hush_deployment/`
  * Data source examples in `examples/data-sources/hush_deployment/`
  * Each example includes main.tf, variables.tf, terraform.tfvars.example, and README.md
* **Auto-generated Documentation**: Provider documentation generated from schema definitions using tfplugindocs
* **Structured Logging**: Enhanced logging using terraform-plugin-log/tflog for better debugging and troubleshooting
* **Modular Architecture**: Organized provider code into modular structure following Terraform best practices
* **Enhanced HTTP Client**: Proper error handling, token lifecycle management, and response body closure
* **Go 1.24 Support**: Built with latest Go toolchain for optimal performance and security
