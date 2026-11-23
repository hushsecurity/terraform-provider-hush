package cloudflare_r2_access_privilege

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestCheckBuckets(t *testing.T) {
	str := cty.StringVal
	set := func(names ...string) cty.Value {
		vals := make([]cty.Value, len(names))
		for i, n := range names {
			vals[i] = str(n)
		}
		return cty.SetVal(vals)
	}
	noBuckets := cty.NullVal(cty.Set(cty.String))
	noAll := cty.NullVal(cty.Bool)

	cases := []struct {
		name       string
		permission cty.Value
		buckets    cty.Value
		allBuckets cty.Value
		wantErr    string
	}{
		{"object with buckets", str("object_read_write"), set("photos"), noAll, ""},
		{"object with all_buckets", str("object_read_only"), noBuckets, cty.True, ""},
		{"object with buckets and all_buckets false", str("object_read_only"), set("photos"), cty.False, ""},
		{"admin alone", str("admin_read_write"), noBuckets, noAll, ""},
		{"admin with all_buckets false", str("admin_read_only"), noBuckets, cty.False, ""},
		{"object with neither", str("object_read_write"), noBuckets, noAll, "needs buckets, or all_buckets = true"},
		{"object with all_buckets false", str("object_read_write"), noBuckets, cty.False, "needs buckets, or all_buckets = true"},
		{"object with both", str("object_read_write"), set("photos"), cty.True, "not both"},
		{"empty buckets", str("object_read_write"), cty.SetValEmpty(cty.String), noAll, "buckets cannot be empty"},
		{"admin with buckets", str("admin_read_write"), set("photos"), noAll, "account wide and takes no buckets"},
		{"admin with all_buckets", str("admin_read_only"), noBuckets, cty.True, "account wide and takes no buckets"},
		// Unknown at plan: left to apply.
		{"unknown permission", cty.UnknownVal(cty.String), noBuckets, noAll, ""},
		{"unknown buckets", str("object_read_write"), cty.UnknownVal(cty.Set(cty.String)), noAll, ""},
		{"unknown all_buckets", str("object_read_write"), noBuckets, cty.UnknownVal(cty.Bool), ""},
		{
			"unknown bucket name",
			str("object_read_write"), cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}), noAll, "",
		},
		{
			"unknown bucket name on admin",
			str("admin_read_write"), cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)}), noAll,
			"account wide and takes no buckets",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkBuckets(tc.permission, tc.buckets, tc.allBuckets)
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

func TestBucketNames(t *testing.T) {
	for name, want := range map[string]bool{
		"photos":    true,
		"a1-b2":     true,
		"abc":       true,
		"ab":        false, // under 3 characters
		"Photos":    false,
		"my_bucket": false,
		"-photos":   false,
		"photos-":   false,
		"photos*":   false, // would widen the token
		"*":         false,
	} {
		if got := bucketRe.MatchString(name); got != want {
			t.Errorf("%q: match = %v, want %v", name, got, want)
		}
	}
	if long := strings.Repeat("a", 63); !bucketRe.MatchString(long) {
		t.Error("a 63-character name was refused")
	}
	if tooLong := strings.Repeat("a", 64); bucketRe.MatchString(tooLong) {
		t.Error("a 64-character name was accepted")
	}
}
