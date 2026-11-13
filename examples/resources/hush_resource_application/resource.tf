resource "hush_resource_application" "advisors" {
  resource       = hush_resource_server.clientworks.name
  display_name   = "Clientworks - Advisors"
  deployment_ids = [hush_deployment.gateway.id]
  scopes         = ["accounts:read"]

  # The advisors, and the agents that may act for them.
  assignment {
    condition {
      match {
        source   = "user"
        property = "groups"
        op       = "in"
        values   = ["grp-advisors"]
      }
    }
  }
  assignment {
    condition {
      match {
        source   = "agent"
        property = "id"
        op       = "in"
        values   = [hush_service_agent.summarizer.id]
      }
    }
  }
}
