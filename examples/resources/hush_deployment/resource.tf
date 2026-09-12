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

# A deployment that trusts more than one issuer. Repeat the block per issuer --
# the provider stores them all in the API's oidc_providers field.
resource "hush_deployment" "multi_oidc" {
  name     = "multi-oidc-deployment"
  env_type = "dev"
  kind     = "k8s"

  oidc_provider {
    issuer           = "https://oidc.eks.eu-central-1.amazonaws.com/id/AAAA1111BBBB2222CCCC3333DDDD4444"
    audience         = "https://kubernetes.default.svc"
    allowed_subjects = ["system:serviceaccount:hush-security:*"]
  }

  oidc_provider {
    issuer           = "https://oidc.eks.eu-central-1.amazonaws.com/id/EEEE5555FFFF6666AAAA7777BBBB8888"
    audience         = "https://kubernetes.default.svc"
    allowed_subjects = ["system:serviceaccount:hush-security:*"]
  }
}

# A deployment covering a cluster where you also run the MCP gateway yourself.
# The hostname must be exactly the agw.hostname the hush-agw chart is installed
# with -- nothing reconciles the two.
resource "hush_deployment" "gateway" {
  name     = "gateway-deployment"
  env_type = "prod"
  kind     = "k8s"

  agw {
    hostname = "gw.example.com"
  }
}

# A gateway Hush runs. The region places it; the address is derived, so read it
# back from gateway_url rather than stating it. The region cannot be changed
# once the gateway is built -- moving one means deleting this deployment and
# creating another.
resource "hush_deployment" "hosted_gateway" {
  name     = "hosted-gateway-deployment"
  env_type = "prod"
  kind     = "hosted"

  agw {
    region = "fra"
  }
}

output "hosted_gateway_url" {
  value = hush_deployment.hosted_gateway.agw[0].gateway_url
}

output "deployment" {
  value = hush_deployment.example
}
