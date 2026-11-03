# Create a Pinot dynamic access credential. host and port name the controller;
# the cluster must keep its users in ZooKeeper (ZkBasicAuthAccessControlFactory).
resource "hush_pinot_access_credential" "example" {
  name                = "prod-pinot"
  description         = "Production Pinot credential"
  deployment_ids      = [hush_deployment.example.id]
  host                = "pinot-controller.example.com"
  port                = 9000
  username            = "admin"
  password_wo         = var.pinot_password
  password_wo_version = "1"
  broker_url          = "http://pinot-broker.example.com:8099"
}
