# Terraform Provider for Hush Security

The [Hush Terraform Provider](https://registry.terraform.io/providers/hushsecurity/hush/latest) allows managing Hush sensor deployments via the Hush Security API.

## Requirements

* Terraform >= 1.3
* Hush API credentials (`client_id` and `client_secret`)

## Inputs

* `client_id` – API key ID
* `client_secret` – API key secret

These can also be set via environment variables:

```bash
export HUSH_API_KEY_ID=...
export HUSH_API_KEY_SECRET=...
```

## Development

For local testing, use the `.terraformrc` `dev_overrides` configuration to point Terraform to your local plugin build.

## License

[Apache 2.0](./LICENSE)
