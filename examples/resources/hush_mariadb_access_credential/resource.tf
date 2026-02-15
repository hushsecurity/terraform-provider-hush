# Create a MariaDB dynamic access credential
resource "hush_mariadb_access_credential" "example" {
  name           = "prod-mariadb"
  description    = "Production MariaDB database credential"
  deployment_ids = [hush_deployment.example.id]
  db_name        = "mydb"
  host           = "mariadb.example.com"
  port           = 3306
  ssl_mode       = "required"
  username       = "app_user"
  password_wo    = var.mariadb_password
}
