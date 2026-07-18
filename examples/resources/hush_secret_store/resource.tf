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
