resource "hush_deployment" "example" {
  name        = "example-deployment"
  description = "Example deployment for testing"
  env_type    = "dev"
  kind        = "k8s"
}

# Deployment with OIDC configuration for passwordless token exchange. The agent
# presents a signed OIDC token (e.g. a Kubernetes service account token) instead
# of the deployment password.
resource "hush_deployment" "oidc" {
  name     = "oidc-deployment"
  env_type = "prod"
  kind     = "k8s"

  oidc_provider {
    issuer           = "https://oidc.eks.us-east-1.amazonaws.com/id/D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9"
    audience         = "https://kubernetes.default.svc"
    allowed_subjects = ["system:serviceaccount:hush-security:*"]
  }
}

output "deployment" {
  value = hush_deployment.example
}
