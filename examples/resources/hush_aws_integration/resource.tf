# Single account AWS integration with onboarding module
resource "hush_aws_integration" "example" {
  name        = "my-aws-integration"
  description = "AWS integration using IAM role"

  # Role ARN from onboarding module
  role_arn = module.hush_aws_onboard.role_arn
}

# AWS onboarding module creates IAM role with required permissions
module "hush_aws_onboard" {
  source  = "hushsecurity/onboard/aws"
  version = ">= 1.4.0"

  hush_org_id = "org-xxxxxxxxxxxx" # Your Hush org ID

  # Optional: restrict to specific regions
  # allowed_regions = ["us-east-1", "us-west-2"]
}
