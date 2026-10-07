# An agent running on AWS (e.g. Bedrock AgentCore) that can also act for the
# users a front end signs in for it.
resource "hush_service_agent" "summarizer" {
  name           = "summarizer"
  acts_for_users = true

  aws_iam_role {
    issuer   = "https://a1c0f603-652b-4692-937c-822f21592bdc.tokens.sts.global.api.aws"
    role_arn = "arn:aws:iam::123456789012:role/summarizer-runtime"
  }
}

# An agent running in Kubernetes, limited to one cluster's service account.
resource "hush_service_agent" "nightly_report" {
  name = "nightly-report"

  kubernetes {
    issuer          = "https://oidc.eks.eu-central-1.amazonaws.com/id/2881A7E8F2BEE882D793CDE80D900F5F"
    namespace       = "agents"
    service_account = "nightly-report"
  }
}

# An Azure AI Foundry agent, signing in with its Entra agent identity. The
# app registration (audience) must issue v2 tokens and be dedicated to Hush.
resource "hush_service_agent" "foundry_helper" {
  name = "foundry-helper"

  azure {
    tenant_id            = "8ea310af-fd38-43b1-8ff6-17df102dc19a"
    service_principal_id = "f845fb9c-55e8-4fc3-b03c-ca3fde64cdcf"
    audience             = "609677e2-6afc-4fc1-a44b-0e5e5d662b3e"
  }
}

# A front end signs users in for the summarizer with this resource.
output "summarizer_resource" {
  value = hush_service_agent.summarizer.resource
}
