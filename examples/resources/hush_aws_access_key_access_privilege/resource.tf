# Create an AWS access key access privilege
resource "hush_aws_access_key_access_privilege" "example" {
  name        = "s3-read-access"
  description = "AWS S3 read access privilege"
  policies    = ["arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"]
}
