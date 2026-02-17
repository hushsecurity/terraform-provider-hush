# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.3.0] - 2026-02-17

### Added

* **Static Credential Providers**: Long-lived credentials assigned to workloads via access policies, for systems that do not support dynamic credential issuance
  * `hush_plaintext_access_credential` resource and data source for managing single secret values
  * `hush_kv_access_credential` resource and data source for managing key-value secret pairs

* **Dynamic Credential Providers**: Short-lived credentials issued on demand and activated exclusively through access policies, enabling just-in-time access with automatic expiration and safe revocation
  * `hush_postgres_access_credential` resource and data source for PostgreSQL
  * `hush_mongodb_access_credential` resource and data source for MongoDB
  * `hush_mysql_access_credential` resource and data source for MySQL
  * `hush_mariadb_access_credential` resource and data source for MariaDB
  * `hush_openai_access_credential` resource and data source for OpenAI
  * `hush_gemini_access_credential` resource and data source for Google Gemini

* **Access Privileges**: Reusable permission sets that define what actions are granted, decoupled from workloads and credential providers
  * `hush_postgres_access_privilege` resource with column-level and schema-wide grant support
  * `hush_mongodb_access_privilege` resource with database and collection-level grants
  * `hush_mysql_access_privilege` resource with database and table-level grants
  * `hush_openai_access_privilege` resource with role-based and restricted permission support

* **Access Policies**: Define when access is allowed and how it is issued by combining attestation conditions, a credential provider, and a privilege — evaluated automatically at runtime
  * `hush_access_policy` resource for binding credentials, privileges, and deployments
  * Kubernetes attestation criteria support
  * Environment variable delivery configuration with key mapping and template-based delivery

### Changed

* Renamed delivery item schema field `value` to `name` to match API
* Marked `api_key_secret` provider attribute as sensitive
* Removed `created_at`/`modified_at` from existing service resources

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
