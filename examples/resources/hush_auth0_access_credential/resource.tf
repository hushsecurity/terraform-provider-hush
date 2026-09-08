# Create an Auth0 dynamic access credential.
#
# The root credential is an Auth0 machine-to-machine application authorized for
# the Management API with `read:clients`, `read:client_grants` and
# `create:client_credentials`, `read:client_credentials`,
# `update:client_credentials` and `delete:client_credentials`. Hush uses it to
# rotate a Private Key JWT keypair on the application each privilege names.
#
# `domain` must be the tenant's canonical domain, shown under Settings >
# General. Set `custom_domain` as well if requests should be routed through a
# custom domain: the Management API audience is always derived from `domain`,
# while the hosts follow `custom_domain`.
#
# Each version delivers `domain`, `client_id` and `private_key`. The workload
# signs a client assertion with that key, so it needs an Auth0 SDK or an OAuth
# client that supports `private_key_jwt`.
resource "hush_auth0_access_credential" "example" {
  name                     = "prod-auth0"
  description              = "Production Auth0 management credential"
  deployment_ids           = [hush_deployment.example.id]
  domain                   = "acme.us.auth0.com"
  client_id                = var.auth0_client_id
  client_secret_wo         = var.auth0_client_secret
  client_secret_wo_version = 1
}
