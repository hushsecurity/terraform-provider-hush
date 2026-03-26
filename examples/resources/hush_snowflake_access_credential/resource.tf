# Create a Snowflake dynamic access credential with password authentication
resource "hush_snowflake_access_credential" "example" {
  name           = "prod-snowflake"
  description    = "Production Snowflake database credential"
  deployment_ids = [hush_deployment.example.id]
  account        = "MYORG-MYACCOUNT"
  warehouse      = "COMPUTE_WH"
  database       = "PROD_DB"
  schema         = "PUBLIC"
  username       = "admin_user"
  auth_method    = "password"
  password_wo    = var.snowflake_password
}

# Create a Snowflake dynamic access credential with key-pair authentication
resource "hush_snowflake_access_credential" "keypair" {
  name           = "prod-snowflake-keypair"
  description    = "Production Snowflake credential using key-pair auth"
  deployment_ids = [hush_deployment.example.id]
  account        = "MYORG-MYACCOUNT"
  warehouse      = "COMPUTE_WH"
  database       = "PROD_DB"
  schema         = "PUBLIC"
  role           = "SYSADMIN"
  username       = "admin_user"
  auth_method    = "key-pair"
  private_key_wo = var.snowflake_private_key
}
