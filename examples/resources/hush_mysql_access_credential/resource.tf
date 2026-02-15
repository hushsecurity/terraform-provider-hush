# Create a MySQL dynamic access credential
resource "hush_mysql_access_credential" "example" {
  name           = "prod-mysql"
  description    = "Production MySQL database credential"
  deployment_ids = [hush_deployment.example.id]
  db_name        = "mydb"
  host           = "mysql.example.com"
  port           = 3306
  ssl_mode       = "required"
  username       = "app_user"
  password_wo    = var.mysql_password
}
