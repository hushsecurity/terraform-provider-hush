# Create an Elasticsearch dynamic access credential
resource "hush_elasticsearch_access_credential" "example" {
  name           = "prod-elasticsearch"
  description    = "Production Elasticsearch credential"
  deployment_ids = [hush_deployment.example.id]
  host           = "elasticsearch.example.com"
  port           = 9200
  username       = "elastic"
  password_wo    = var.elasticsearch_password
  tls            = true
}
