# Create a RabbitMQ access privilege
resource "hush_rabbitmq_access_privilege" "example" {
  name        = "my-rabbitmq-privilege"
  description = "RabbitMQ privilege with full access to default vhost"

  permissions {
    vhost     = "/"
    configure = ""         # no permission to create/delete resources
    write     = "app\\..*" # publish only to app.* exchanges/queues
    read      = ".*"       # consume from any queue
  }

  tags = ["monitoring"]
}
