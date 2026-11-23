# Read and write the objects of two buckets
resource "hush_cloudflare_r2_access_privilege" "example" {
  name        = "r2-photos-read-write"
  description = "Read and write photos and exports"
  permission  = "object_read_write"
  buckets     = ["photos", "exports"]
}

# Read the objects of every bucket of the account, including later ones
resource "hush_cloudflare_r2_access_privilege" "all_buckets" {
  name        = "r2-read-all"
  permission  = "object_read_only"
  all_buckets = true
}

# Read the objects of a bucket in the EU jurisdiction
resource "hush_cloudflare_r2_access_privilege" "eu" {
  name         = "r2-eu-reports-read"
  permission   = "object_read_only"
  buckets      = ["reports"]
  jurisdiction = "eu"
}
