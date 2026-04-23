# Create a Redis dynamic access credential (plain Redis, password auth)
resource "hush_redis_access_credential" "example" {
  name                = "prod-redis"
  description         = "Production Redis credential"
  deployment_ids      = [hush_deployment.example.id]
  host                = "redis.example.com"
  port                = 6379
  username            = "default"
  password_wo         = var.redis_password
  password_wo_version = "1"
  tls                 = true
  database            = 0
  engine              = "redis"
}

# Create an AWS ElastiCache dynamic access credential.
# Provisions ephemeral users via the ElastiCache CreateUser API and adds
# them to the configured user group. Omit access_key_id/secret_access_key
# to fall back to the AWS default credential chain (IRSA / instance
# profile / WIF).
resource "hush_redis_access_credential" "elasticache_example" {
  name           = "prod-elasticache"
  description    = "Production ElastiCache (Valkey) credential"
  deployment_ids = [hush_deployment.example.id]
  host           = "my-cluster.xxxxxx.0001.use1.cache.amazonaws.com"
  port           = 6379
  tls            = true
  engine         = "elasticache"
  cache_engine   = "valkey"
  region         = "us-east-1"
  user_group_id  = "my-elasticache-user-group"
}
