# Create a MongoDB Atlas dynamic access credential using a service account
resource "hush_mongodb_atlas_access_credential" "example" {
  name           = "prod-atlas"
  description    = "Production MongoDB Atlas database credential"
  deployment_ids = [hush_deployment.example.id]
  group_id       = "5e2211c17a3e5a48f5497de3"
  db_name        = "mydb"
  host           = "cluster0.abcde.mongodb.net"

  # Service account authentication
  client_id                = "mdb_sa_id_abc123"
  client_secret_wo         = var.atlas_client_secret
  client_secret_wo_version = "1"
}

# Alternatively, authenticate with an API key pair:
#
#   public_key             = "abcdefgh"
#   private_key_wo         = var.atlas_private_key
#   private_key_wo_version = "1"
