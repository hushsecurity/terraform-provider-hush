# Create a Redis dynamic access credential
resource "hush_redis_access_credential" "example" {
  name           = "prod-redis"
  description    = "Production Redis credential"
  deployment_ids = [hush_deployment.example.id]
  host           = "redis.example.com"
  port           = 6379
  username       = "default"
  password_wo    = var.redis_password
  tls            = true
  database       = 0
}
