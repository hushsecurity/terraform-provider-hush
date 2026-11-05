package github_app_access_privilege

import (
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestCheckPermissions(t *testing.T) {
	cases := []struct {
		name    string
		perms   cty.Value
		wantErr string
	}{
		{"null", cty.NullVal(cty.Map(cty.String)), ""},
		{"wholly unknown", cty.UnknownVal(cty.Map(cty.String)), ""},
		{"valid", cty.MapVal(map[string]cty.Value{"contents": cty.StringVal("read")}), ""},
		{
			"unknown level is skipped",
			cty.MapVal(map[string]cty.Value{"contents": cty.UnknownVal(cty.String)}),
			"",
		},
		{"empty", cty.MapValEmpty(cty.String), `name at least one permission`},
		{
			"bad level",
			cty.MapVal(map[string]cty.Value{"contents": cty.StringVal("superuser")}),
			`level must be one of read, write, admin, not "superuser"`,
		},
		{
			"bad name with unknown level",
			cty.MapVal(map[string]cty.Value{"Contents-Read": cty.UnknownVal(cty.String)}),
			`not a GitHub App permission name`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkErr(t, checkPermissions(tc.perms), tc.wantErr)
		})
	}
}

func TestCheckRepositories(t *testing.T) {
	cases := []struct {
		name    string
		repos   cty.Value
		wantErr string
	}{
		{"null", cty.NullVal(cty.List(cty.String)), ""},
		{"wholly unknown", cty.UnknownVal(cty.List(cty.String)), ""},
		{"empty", cty.ListValEmpty(cty.String), ""},
		{"distinct", cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}), ""},
		{
			"unknown elements are skipped",
			cty.ListVal([]cty.Value{cty.UnknownVal(cty.String), cty.UnknownVal(cty.String), cty.StringVal("a")}),
			"",
		},
		{
			"duplicate",
			cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b"), cty.StringVal("a")}),
			`repository "a": listed more than once`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkErr(t, checkRepositories(tc.repos), tc.wantErr)
		})
	}
}

func checkErr(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil || !regexp.MustCompile(regexp.QuoteMeta(want)).MatchString(err.Error()) {
		t.Fatalf("expected error matching %q, got %v", want, err)
	}
}
