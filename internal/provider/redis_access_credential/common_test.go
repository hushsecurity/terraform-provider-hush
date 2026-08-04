package redis_access_credential

import "testing"

// Every expectation below was taken from midgard's own validator
// (TypeAdapter(AzureResourceGroup) under pydantic 2.13.4), not from a reading
// of the pattern. Azure permits Unicode in resource group names, and the
// non-ASCII cases are the ones a bare Go \w would reject.
func TestResourceGroupRegex(t *testing.T) {
	for _, tc := range []struct {
		name  string
		valid bool
	}{
		{"prod-rg", true},
		{"rg_1", true},
		{"(rg)", true},
		{"-rg-", true},
		{"rg.1", true},
		{"grupo-señal", true},
		{"リソース-グループ", true},
		{"משאבים", true},
		{"rǵ", true},
		// Marks, connectors and joiners are word characters to the regex crate.
		{"संसाधन", true},
		{"rg⁀", true},
		{"rg\u200d1", true},
		{"rg\u200c1", true},
		// Nl is alphabetic, other numbers are not.
		{"Ⅷ", true},
		{"rg½", false},
		{"rg.", false},
		{"rg space", false},
		{"rg!", false},
		{"", false},
	} {
		if got := resourceGroupRegex.MatchString(tc.name); got != tc.valid {
			t.Errorf("resourceGroupRegex.MatchString(%q) = %v, want %v", tc.name, got, tc.valid)
		}
	}
}

// Lowercase-only, since midgard serializes these with str(UUID).
func TestUUIDRegex(t *testing.T) {
	for _, tc := range []struct {
		name  string
		valid bool
	}{
		{"1111aaaa-1111-1111-1111-111111111111", true},
		{"1111AAAA-1111-1111-1111-111111111111", false},
		{"1111aaaa-1111-1111-1111-11111111111", false},
		{"not-a-uuid", false},
	} {
		if got := uuidRegex.MatchString(tc.name); got != tc.valid {
			t.Errorf("uuidRegex.MatchString(%q) = %v, want %v", tc.name, got, tc.valid)
		}
	}
}
