package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestUpdateDeploymentInput_OidcProviderMarshaling verifies the three update
// states of oidc_provider: omitted when unchanged, explicit null when removed,
// and the config object when set.
func TestUpdateDeploymentInput_OidcProviderMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateDeploymentInput
		contains string
		omits    bool
	}{
		{
			name:  "unchanged omits the field",
			input: UpdateDeploymentInput{},
			omits: true,
		},
		{
			name:     "removal sends explicit null",
			input:    UpdateDeploymentInput{OidcProvider: NewOidcProviderUpdate(nil)},
			contains: `"oidc_provider":null`,
		},
		{
			name: "set sends the config object",
			input: UpdateDeploymentInput{OidcProvider: NewOidcProviderUpdate(&OidcConfig{
				Issuer:          "https://issuer.example.com",
				Audience:        "hush",
				AllowedSubjects: []string{"system:serviceaccount:hush:*"},
			})},
			contains: `"oidc_provider":{"issuer":"https://issuer.example.com","audience":"hush","allowed_subjects":["system:serviceaccount:hush:*"]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			got := string(b)
			if tc.omits {
				if strings.Contains(got, "oidc_provider") {
					t.Fatalf("expected oidc_provider to be omitted, got %s", got)
				}
				return
			}
			if !strings.Contains(got, tc.contains) {
				t.Fatalf("expected %s to contain %s", got, tc.contains)
			}
		})
	}
}
