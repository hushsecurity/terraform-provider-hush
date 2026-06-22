# Create a Kafka dynamic access credential for a self-managed (native) cluster
resource "hush_kafka_access_credential" "native" {
  name              = "prod-kafka"
  description       = "Production Kafka cluster credential"
  deployment_ids    = [hush_deployment.example.id]
  engine            = "native"
  bootstrap_servers = "broker1:9092,broker2:9092"
  username          = "admin"
  sasl_mechanism    = "SCRAM-SHA-512"
  tls               = true
  password_wo       = var.kafka_password
}

# Create a Kafka dynamic access credential for an Aiven-managed service
resource "hush_kafka_access_credential" "aiven" {
  name           = "prod-kafka-aiven"
  description     = "Production Kafka credential on Aiven"
  deployment_ids = [hush_deployment.example.id]
  engine         = "aiven"
  project        = "my-aiven-project"
  service_name   = "my-kafka-service"
  token_wo       = var.aiven_token
}
