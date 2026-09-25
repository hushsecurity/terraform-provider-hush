# An application managed in the console, looked up by the name it is shown
# under.
data "hush_mcp_application" "github" {
  display_name = "GitHub"
}

output "github_blocked_tools" {
  value = [for tool in data.hush_mcp_application.github.tools : tool.name if tool.operation == "block"]
}
