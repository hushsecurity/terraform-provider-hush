package auth0_access_credential

import (
	"strings"
	"testing"
)

// A custom domain here 403s at apply.
func TestDomainMustBeCanonical(t *testing.T) {
	validate := ResourceSchema()["domain"].ValidateFunc
	if validate == nil {
		t.Fatal("domain has no ValidateFunc")
	}

	accepted := []string{
		"acme.us.auth0.com",
		"acme.auth0.com",
		"dev-abc123.eu.auth0.com",
	}
	for _, d := range accepted {
		if _, errs := validate(d, "domain"); len(errs) > 0 {
			t.Errorf("%q should be accepted, got %v", d, errs)
		}
	}

	rejected := []string{
		"auth.acme.com",               // a custom domain
		"evil-auth0.com.attacker.net", // a lookalike suffix
		"acme.auth0.com.attacker.net", // suffix in the middle
		"!!.auth0.com",                // right suffix, not a domain
		"",
	}
	for _, d := range rejected {
		if _, errs := validate(d, "domain"); len(errs) == 0 {
			t.Errorf("%q should be rejected", d)
		}
	}
}

func TestDomainRejectionNamesTheFix(t *testing.T) {
	validate := ResourceSchema()["domain"].ValidateFunc
	_, errs := validate("auth.acme.com", "domain")
	if len(errs) == 0 {
		t.Fatal("expected a rejection")
	}
	msg := errs[0].Error()
	if !strings.Contains(msg, "custom_domain") {
		t.Errorf("the error should point at custom_domain, got: %s", msg)
	}
}

// midgard types custom_domain DomainStr, so a malformed one is a 422 at apply.
func TestCustomDomainMustBeADomain(t *testing.T) {
	validate := ResourceSchema()["custom_domain"].ValidateFunc
	if validate == nil {
		t.Fatal("custom_domain has no ValidateFunc")
	}
	if _, errs := validate("auth.acme.com", "custom_domain"); len(errs) > 0 {
		t.Errorf("a real domain should be accepted, got %v", errs)
	}
	for _, d := range []string{"not a domain", "auth.acme.com/path", ""} {
		if _, errs := validate(d, "custom_domain"); len(errs) == 0 {
			t.Errorf("%q should be rejected", d)
		}
	}
}

// The pairing that guards against a perpetual diff on rotation.
func TestClientSecretWriteOnlyPairing(t *testing.T) {
	s := ResourceSchema()

	if !s["client_secret"].Sensitive {
		t.Error("client_secret must be sensitive")
	}
	if !s["client_secret_wo"].WriteOnly {
		t.Error("client_secret_wo must be write-only")
	}
	for _, f := range []string{"client_secret", "client_secret_wo"} {
		if len(s[f].ExactlyOneOf) != 2 {
			t.Errorf("%s must declare ExactlyOneOf with both forms", f)
		}
	}
	if len(s["client_secret_wo"].RequiredWith) == 0 {
		t.Error("client_secret_wo must require client_secret_wo_version")
	}
	if len(s["client_secret_wo_version"].RequiredWith) == 0 {
		t.Error("client_secret_wo_version must require client_secret_wo")
	}
}

// custom_domain is optional; domain and client_id are not.
func TestRequiredFields(t *testing.T) {
	s := ResourceSchema()
	for _, f := range []string{"domain", "client_id"} {
		if !s[f].Required {
			t.Errorf("%s must be required", f)
		}
	}
	if s["custom_domain"].Required {
		t.Error("custom_domain must be optional")
	}
	if !s["custom_domain"].Optional {
		t.Error("custom_domain must be optional")
	}
}
