data "hush_service_agent" "summarizer" {
  name = "summarizer"
}

output "summarizer_id" {
  value = data.hush_service_agent.summarizer.id
}
