# Create a RabbitMQ dynamic access credential
resource "hush_rabbitmq_access_credential" "example" {
  name            = "prod-rabbitmq"
  description     = "Production RabbitMQ credential"
  deployment_ids  = [hush_deployment.example.id]
  host            = "rabbitmq.example.com"
  port            = 5672
  management_port = 15672
  username        = "admin"
  password_wo     = var.rabbitmq_password
  vhost           = "/"
  tls             = true
}
