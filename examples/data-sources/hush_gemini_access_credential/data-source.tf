data "hush_gemini_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_gemini_access_credential.example.name
}

output "project_id" {
  value = data.hush_gemini_access_credential.example.project_id
}
