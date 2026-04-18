# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

* **Sendgrid**: `hush_sendgrid_access_credential` and `hush_sendgrid_access_privilege` resources and data sources
* **Salesforce**: `hush_salesforce_access_credential` and `hush_salesforce_access_privilege` resources and data sources
* **Datadog**: `hush_datadog_access_credential` and `hush_datadog_access_privilege` resources and data sources

### Fixed

* **Resources `hush_twilio_access_credential`, `hush_gitlab_access_credential`**: Allow changing `deployment_ids` without forcing resource replacement

## [1.7.0] - 2026-04-13

### Added

* **GCP Workload Identity Federation**: Issue short-lived GCP credentials through WIF, without managing long-lived service account keys
  * `hush_gcp_wif_access_credential` resource and data source
  * GCP WIF delivery configuration on `hush_access_policy` for just-in-time credential delivery
* **GitLab**: `hush_gitlab_access_credential` and `hush_gitlab_access_privilege` resources and data sources

## [1.6.0] - 2026-04-10

### Added

* **AWS Workload Identity Federation**: Issue short-lived AWS credentials through WIF, without managing long-lived IAM access keys
  * `hush_aws_wif_access_credential` resource and data source
  * AWS WIF delivery configuration on `hush_access_policy` for just-in-time credential delivery

## [1.5.0] - 2026-04-10

### Added

* **Volume Delivery**: New delivery mode for `hush_access_policy` that writes credentials to a mounted volume instead of environment variables, enabling file-based consumption of secrets

## [1.4.0] - 2026-04-09

### Added

* **Snowflake**: `hush_snowflake_access_credential` and `hush_snowflake_access_privilege` resources and data sources

### Changed

* Build provider binary with the Go 1.25 toolchain (previously Go 1.24)

## [1.3.4] - 2026-03-18

### Added

* **Twilio**: `hush_twilio_access_credential` and `hush_twilio_access_privilege` resources and data sources
* **AWS Access Keys**: `hush_aws_access_key_access_credential` and `hush_aws_access_key_access_privilege` resources and data sources
* **Azure App**: `hush_azure_app_access_credential` and `hush_azure_app_access_privilege` resources and data sources
* **GCP Service Account**: `hush_gcp_sa_access_credential` and `hush_gcp_sa_access_privilege` resources and data sources
* **RabbitMQ**: `hush_rabbitmq_access_credential` and `hush_rabbitmq_access_privilege` resources and data sources

## [1.3.3] - 2026-03-14

### Added

* **Elasticsearch**: `hush_elasticsearch_access_credential` and `hush_elasticsearch_access_privilege` resources and data sources
* **Apigee**: `hush_apigee_access_credential` and `hush_apigee_access_privilege` resources and data sources
* **Bedrock**: `hush_bedrock_access_credential` resource and data source
* **Redis**: `hush_redis_access_credential` and `hush_redis_access_privilege` resources and data sources
* **Grok**: `hush_grok_access_credential` and `hush_grok_access_privilege` resources and data sources

### Fixed

* **Dynamic access credentials**: Stop sending `deployment_ids` on update to match API behavior (the field is not updatable server-side)
* **Resource `hush_access_policy`**: Persist resource ID to state when creation fails partway through, so the next apply can recover instead of orphaning the resource
* `HUSH_REALM` environment variable handling

## [1.3.2] - 2026-02-18

### Changed

* Update provider documentation to reflect the flattened `env_delivery_config` schema introduced in v1.3.1

## [1.3.1] - 2026-02-17

### Changed

* **Resource `hush_access_policy`**: Flatten `env_delivery_config` schema by removing the nested `item` block for a simpler configuration
* List documentation under the standard Terraform Registry "Resources" and "Data Sources" groupings (remove subcategory)

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

* Rename delivery item schema field `value` to `name` to match API
* Mark `api_key_secret` provider attribute as sensitive
* Remove `created_at`/`modified_at` from existing service resources

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

* **Deployment Data Source**: Add support for looking up deployments by `name` in addition to `id`

## [1.0.3] - 2025-08-13

### Changed

* **Error Handling**: Improve error handling throughout the provider for better user experience
  * Centralize error handling with new `APIError` struct providing clear, actionable error messages
  * Simplify error propagation with clean type assertion patterns
  * Improve 404 handling for graceful resource state management

### Removed

* **Logging**: Eliminate tflog imports and verbose error logging in provider layer
  * Streamline operations for cleaner, more reliable execution

## [1.0.2] - 2025-08-11

### Fixed

* **Resource `hush_deployment`**: Fix `kind` field validation to accept lowercase values
  * Accept `k8s`, `ecs`, `serverless` (lowercase) to match API requirements
  * Previously required uppercase values (`K8S`, `ECS`, `SERVERLESS`) which caused API errors
* **Resource `hush_deployment`**: Add validation for `env_type` field
  * Accept only `dev` or `prod` values as required by the API
  * Catch invalid values at plan time to prevent runtime errors

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
* **Modular Architecture**: Organize provider code into modular structure following Terraform best practices
* **Enhanced HTTP Client**: Proper error handling, token lifecycle management, and response body closure
* **Go 1.24 Support**: Built with latest Go toolchain for optimal performance and security
