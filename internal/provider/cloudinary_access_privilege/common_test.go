package cloudinary_access_privilege

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestCheckGrant(t *testing.T) {
	str := cty.StringVal
	set := func(ids ...string) cty.Value {
		vals := make([]cty.Value, len(ids))
		for i, id := range ids {
			vals[i] = str(id)
		}
		return cty.SetVal(vals)
	}
	noRole := cty.NullVal(cty.String)
	noPolicies := cty.NullVal(cty.Set(cty.String))
	upload := "cld::policy::global::folder_and_asset_management::create_asset"

	cases := []struct {
		name     string
		roleID   cty.Value
		policies cty.Value
		wantErr  string
	}{
		{"role", str("cld::role::prodenv::editor"), noPolicies, ""},
		{"policies", noRole, set(upload), ""},
		{"neither", noRole, noPolicies, oneGrant},
		{"both", str("cld::role::prodenv::editor"), set(upload), oneGrant},
		{"empty policies", noRole, cty.SetValEmpty(cty.String), "system_policy_ids cannot be empty"},
		{"empty policies with a role", str("cld::role::prodenv::editor"), cty.SetValEmpty(cty.String),
			"system_policy_ids cannot be empty"},
		// Unknown at plan: left to apply.
		{"unknown role", cty.UnknownVal(cty.String), noPolicies, ""},
		{"unknown role with policies", cty.UnknownVal(cty.String), set(upload), ""},
		{"unknown policies", noRole, cty.UnknownVal(cty.Set(cty.String)), ""},
		{"unknown policy id", noRole, cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}), ""},
		{"unknown policy id with a role", str("cld::role::prodenv::editor"),
			cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}), oneGrant},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkGrant(tc.roleID, tc.policies)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestValidatePolicyID(t *testing.T) {
	for id, wantErr := range map[string]string{
		"cld::policy::global::folder_and_asset_management::create_asset": "",
		"cld::policy::global::marketplace:manage":                        "", // single colons in the catalog
		"cld::policy::global::api_keys::manage":                          "",
		"cld::policy::content::folder::view":                             "is not a cld::policy::global:: policy",
		"cld::policy::global::*":                                         "is not a cld::policy::global:: policy",
		"cld::policy::global::Upload":                                    "is not a cld::policy::global:: policy",
		"cld::policy::global::":                                          "is not a cld::policy::global:: policy",
	} {
		_, errs := validatePolicyID(id, "system_policy_ids")
		if wantErr == "" {
			if len(errs) > 0 {
				t.Errorf("%q: unexpected errors %v", id, errs)
			}
			continue
		}
		if len(errs) == 0 || !strings.Contains(errs[0].Error(), wantErr) {
			t.Errorf("%q: errs = %v, want one containing %q", id, errs, wantErr)
		}
	}
}

func TestCloudNamePattern(t *testing.T) {
	for name, want := range map[string]bool{
		"product1":                     true,
		"Product-1":                    true,
		"p1":                           true,
		"p":                            false, // under 2 characters
		"1product":                     false,
		"pro duct":                     false,
		"pro_duct":                     false,
		"a" + strings.Repeat("b", 127): true,
		"a" + strings.Repeat("b", 128): false,
	} {
		if got := cloudNameRe.MatchString(name); got != want {
			t.Errorf("%q: match = %v, want %v", name, got, want)
		}
	}
}
