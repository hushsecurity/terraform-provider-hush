terraform {
  required_providers {
    hush = {
      source = "hushsecurity/hush"
    }
  }
}

provider "hush" {
  api_key_id     = var.api_key_id
  api_key_secret = var.api_key_secret
}
