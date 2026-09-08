package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// Clearing must send null, not "": midgard types the field DomainStr | None
// and "" is a 422.
func TestUpdateAuth0AccessCredentialCustomDomainWireFormat(t *testing.T) {
	cases := []struct {
		name  string
		input UpdateAuth0AccessCredentialInput
		want  string
	}{
		{
			name:  "unchanged is omitted",
			input: UpdateAuth0AccessCredentialInput{},
			want:  `{}`,
		},
		{
			name: "a value is sent as a string",
			input: UpdateAuth0AccessCredentialInput{
				CustomDomain: NewNullableString("auth.acme.com"),
			},
			want: `{"custom_domain":"auth.acme.com"}`,
		},
		{
			name: "cleared is sent as null",
			input: UpdateAuth0AccessCredentialInput{
				CustomDomain: NewNullableString(""),
			},
			want: `{"custom_domain":null}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
			if strings.Contains(string(got), `"custom_domain":""`) {
				t.Error(`sent an empty string; midgard answers 422 for that`)
			}
		})
	}
}
