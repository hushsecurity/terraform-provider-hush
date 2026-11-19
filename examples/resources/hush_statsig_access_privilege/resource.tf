# A server secret key limited to production
resource "hush_statsig_access_privilege" "example" {
  name         = "statsig-server-production"
  description  = "Server SDK key for production"
  key_type     = "server"
  environments = ["production"]
}
