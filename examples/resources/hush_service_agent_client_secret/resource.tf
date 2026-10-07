# A secret for a workload with no identity to sign in with. To rotate, add a
# second secret, move the workload to it, then remove this one.
resource "hush_service_agent_client_secret" "ci" {
  agent_id = hush_service_agent.summarizer.id
}

output "ci_secret" {
  value     = hush_service_agent_client_secret.ci.secret
  sensitive = true
}
