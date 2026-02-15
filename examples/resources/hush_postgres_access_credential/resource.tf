# Create a PostgreSQL dynamic access credential
resource "hush_postgres_access_credential" "example" {
  name           = "prod-postgres"
  description    = "Production PostgreSQL database credential"
  deployment_ids = [hush_deployment.example.id]
  db_name        = "mydb"
  host           = "postgres.example.com"
  port           = 5432
  ssl_mode       = "require"
  username       = "app_user"
  password_wo    = var.postgres_password
}
