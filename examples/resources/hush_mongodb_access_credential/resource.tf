# Create a MongoDB dynamic access credential
resource "hush_mongodb_access_credential" "example" {
  name           = "prod-mongodb"
  description    = "Production MongoDB database credential"
  deployment_ids = [hush_deployment.example.id]
  db_name        = "mydb"
  host           = "mongodb.example.com"
  port           = 27017
  username       = "app_user"
  password_wo    = var.mongodb_password
  auth_source    = "admin"
  tls            = true
}
