resource "hush_resource_server" "clientworks" {
  name         = "clientworks"
  display_name = "Clientworks API"
  audience     = "https://api.example.com/clientworks"

  scope {
    name        = "accounts:read"
    description = "Read accounts"
  }
  scope {
    name        = "trades:write"
    description = "Place trades"
  }

  token_lifetime = 900

  token_claim {
    name      = "teams"
    attribute = "groups"
  }
}
