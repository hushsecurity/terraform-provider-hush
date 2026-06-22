data "hush_kafka_access_credential" "example" {
  id = "acr-eu12345678"
}

output "name" {
  value = data.hush_kafka_access_credential.example.name
}

output "engine" {
  value = data.hush_kafka_access_credential.example.engine
}

output "bootstrap_servers" {
  value = data.hush_kafka_access_credential.example.bootstrap_servers
}

output "sasl_mechanism" {
  value = data.hush_kafka_access_credential.example.sasl_mechanism
}
