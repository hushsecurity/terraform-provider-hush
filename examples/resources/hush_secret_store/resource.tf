# A secret store has exactly one backend configuration block, chosen from the
# supported kinds below. The backend configuration is immutable: changing any field
# in the block (or switching to a different kind) replaces the store.

# AWS Secrets Manager
resource "hush_secret_store" "aws_sm" {
  name           = "prod-aws-sm"
  description    = "Production AWS Secrets Manager store"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  aws_sm {
    prefix     = "hush"
    region     = "eu-west-1"
    kms_key_id = "arn:aws:kms:eu-west-1:123456789012:key/abcd1234-..." # optional
  }
}

# AWS SSM Parameter Store
resource "hush_secret_store" "aws_ssm" {
  name           = "prod-aws-ssm"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  aws_ssm {
    prefix = "hush"
    region = "us-east-1"
    # kms_key_id = "arn:aws:kms:us-east-1:123456789012:key/abcd1234-..." # optional
  }
}

# GCP Secret Manager
resource "hush_secret_store" "gcp_sm" {
  name           = "prod-gcp-sm"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  gcp_sm {
    prefix     = "hush"
    project_id = "my-gcp-project"
  }
}

# Kubernetes Secrets
resource "hush_secret_store" "k8s" {
  name           = "prod-k8s"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  k8s_secrets {
    prefix    = "hush"
    namespace = "hush-secrets" # optional; defaults to the access-manager namespace
  }
}

# HashiCorp Vault, on its KV v2 secrets engine
resource "hush_secret_store" "hc_vault" {
  name           = "prod-vault"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  hc_vault {
    prefix  = "hush"
    address = "https://vault.example.internal:8200"
    mount   = "secret" # optional; the KV v2 mount, "secret" when omitted

    # optional; only when the server's certificate does not chain to a
    # publicly trusted root
    # ca_cert = file("vault-ca.pem")

    # optional; a Vault Enterprise namespace
    # vault_namespace = "admin/acme"

    auth {
      # the access manager presents its pod's service-account token to this
      # role, and the cluster's TokenReview api vouches for it
      role = "hush-am"
      # method = "kubernetes" # optional; the default
      # mount  = "kubernetes" # optional; where the auth method is mounted
    }
  }
}

# The same, for a Vault that cannot reach the cluster's api server: the
# service-account token is validated against the cluster's JWKS instead.
resource "hush_secret_store" "hc_vault_jwt" {
  name           = "prod-vault-jwt"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  hc_vault {
    prefix  = "hush"
    address = "https://vault.example.internal:8200"

    auth {
      method = "jwt"
      role   = "hush-am"
    }
  }
}

# A token the deployment already holds. Nothing here names it: set
# secretStore.hcVault.auth.token or auth.tokenSecretRef in the deployment's
# hush-am chart values. Nothing renews it, so prefer "kubernetes".
resource "hush_secret_store" "hc_vault_token" {
  name           = "prod-vault-token"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  hc_vault {
    prefix  = "hush"
    address = "https://vault.example.internal:8200"

    auth {
      method = "token"
    }
  }
}

# az_key_vault - Azure Key Vault
resource "hush_secret_store" "az_key_vault" {
  name           = "prod-keyvault"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  az_key_vault {
    # at most 32 characters: a Key Vault secret name is capped at 127 and the
    # rest is taken by what the access manager appends
    prefix    = "hush"
    vault_url = "https://acme-prod.vault.azure.net"
    # cloud   = "usgov" # with a .vault.usgovcloudapi.net vault_url;
    #                   # "public" when omitted

    auth {
      # the access manager's own identity -- on AKS, workload identity, set
      # through accessManager.workloadIdentity.azure.clientId in the
      # deployment's hush-am chart values, with no secret configured anywhere
      method = "default"
    }
  }
}

# A service principal, for a cluster without workload identity. Nothing here
# names the secret: set secretStore.azKeyVault.auth.clientSecretRef, or
# auth.clientSecret, in the deployment's hush-am chart values.
resource "hush_secret_store" "az_key_vault_client_secret" {
  name           = "prod-keyvault-sp"
  deployment_ids = ["dep-xxxxxxxxxxxxxxxx"]

  az_key_vault {
    prefix    = "hush"
    vault_url = "https://acme-prod.vault.azure.net"

    auth {
      method    = "client_secret"
      tenant_id = "00000000-0000-0000-0000-000000000000"
      client_id = "00000000-0000-0000-0000-000000000000"
    }
  }
}
