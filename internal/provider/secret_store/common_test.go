package secret_store

import "testing"

// The acceptance tests cover these through a plan, but need the fixtures
// "make fetch-mock-fixtures" pulls. These are pure functions, so the boundary
// rows are cheaper here. Expectations mirror midgard's validate_prefix.
func TestPrefixValidation(t *testing.T) {
	for _, tc := range []struct {
		kind     string
		validate func(any, string) ([]string, []error)
		prefix   string
		valid    bool
	}{
		// each kind's own punctuation
		{"aws_sm", awsSMPrefixValidation, "acme", true},
		{"aws_sm", awsSMPrefixValidation, "acme-prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme_prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme.prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme+prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme=prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme@prod", true},
		{"aws_sm", awsSMPrefixValidation, "acme/prod", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme_prod", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme.prod", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme/prod/secrets", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme+prod", false},
		{"aws_ssm", awsSSMPrefixValidation, "acme@prod", false},
		{"gcp_sm", gcpSMPrefixValidation, "acme_prod", true},
		{"gcp_sm", gcpSMPrefixValidation, "acme.prod", false},
		{"gcp_sm", gcpSMPrefixValidation, "acme/prod", false},
		{"k8s_secrets", k8sPrefixValidation, "acme.prod", true},
		{"k8s_secrets", k8sPrefixValidation, "acme_prod", false},
		{"k8s_secrets", k8sPrefixValidation, "acme/prod", false},

		// adjacency: allowed where the backend attaches no meaning to it, refused
		// for k8s_secrets, whose remote key is a DNS subdomain
		{"aws_sm", awsSMPrefixValidation, "ac__me", true},
		{"aws_sm", awsSMPrefixValidation, "ac--me", true},
		{"aws_sm", awsSMPrefixValidation, "ac_.me", true},
		{"aws_ssm", awsSSMPrefixValidation, "ac__me", true},
		{"gcp_sm", gcpSMPrefixValidation, "ac__me", true},
		{"k8s_secrets", k8sPrefixValidation, "ac..me", false},
		{"k8s_secrets", k8sPrefixValidation, "ac.-me", false},

		// shape: punctuation never leads or trails
		{"aws_sm", awsSMPrefixValidation, "/acme", false},
		{"aws_sm", awsSMPrefixValidation, "acme/", false},
		{"aws_sm", awsSMPrefixValidation, "acme//prod", false},
		{"gcp_sm", gcpSMPrefixValidation, "-acme", false},
		{"gcp_sm", gcpSMPrefixValidation, "acme-", false},
		{"gcp_sm", gcpSMPrefixValidation, "ac--me", false}, // "-" is its separator
		{"k8s_secrets", k8sPrefixValidation, "", false},

		// uppercase: every layer above mufasa is lowercase-only
		{"aws_sm", awsSMPrefixValidation, "Acme", false},
		{"aws_ssm", awsSSMPrefixValidation, "ACME", false},

		// reserved case-insensitively, and only as a prefix
		{"aws_ssm", awsSSMPrefixValidation, "aws", false},
		{"aws_ssm", awsSSMPrefixValidation, "aws-corp", false},
		{"aws_ssm", awsSSMPrefixValidation, "ssmx", false},
		{"aws_ssm", awsSSMPrefixValidation, "awt", true},
		{"aws_ssm", awsSSMPrefixValidation, "acme-aws", true},
		{"aws_sm", awsSMPrefixValidation, "aws-corp", true},
		{"gcp_sm", gcpSMPrefixValidation, "ssmx", true},

		// the depth budget, at and past the boundary
		{"aws_ssm", awsSSMPrefixValidation, "a/b/c/d/e/f/g/h/i", true},
		{"aws_ssm", awsSSMPrefixValidation, "a/b/c/d/e/f/g/h/i/j", false},
		{"aws_sm", awsSMPrefixValidation, "a/b/c/d/e/f/g/h/i/j", true},
	} {
		_, errs := tc.validate(tc.prefix, "prefix")
		if got := len(errs) == 0; got != tc.valid {
			t.Errorf("%s: validate(%q) valid = %v, want %v (%v)",
				tc.kind, tc.prefix, got, tc.valid, errs)
		}
	}
}
