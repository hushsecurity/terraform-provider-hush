# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2025-07-14

### Added

* Initial public release of the `hushsecurity/hush` Terraform provider.
* Supports authentication via client credentials (`client_id`, `client_secret`).
* Resource `hush_deployment`:

  * Create Hush deployments via Terraform
  * Supports `name`, `description`, and `env_type` fields
* Automatically fetches deployment token and password after creation.
* Supports deletion and reading of deployments.
