package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestUpdateInput_SecretStoreIDMarshaling verifies the three update states of
// secret_store_id: omitted when unchanged, explicit null when detached, and the
// id when re-pointed. midgard rejects "" (the field must be a valid sst- id or
// null), so detaching must send null rather than an empty string.
func TestUpdateInput_SecretStoreIDMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdatePlaintextAccessCredentialInput
		contains string
		omits    bool
	}{
		{
			name:  "unchanged omits the field",
			input: UpdatePlaintextAccessCredentialInput{},
			omits: true,
		},
		{
			name:     "detach sends explicit null",
			input:    UpdatePlaintextAccessCredentialInput{SecretStoreID: NewSecretStoreIDUpdate("")},
			contains: `"secret_store_id":null`,
		},
		{
			name:     "set sends the id",
			input:    UpdatePlaintextAccessCredentialInput{SecretStoreID: NewSecretStoreIDUpdate("sst-abc123")},
			contains: `"secret_store_id":"sst-abc123"`,
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
				if strings.Contains(got, "secret_store_id") {
					t.Fatalf("expected secret_store_id to be omitted, got %s", got)
				}
				return
			}
			if !strings.Contains(got, tc.contains) {
				t.Fatalf("expected %s to contain %s", got, tc.contains)
			}
		})
	}
}

// TestUpdateRedisInput_AzureAppCredentialMarshaling verifies the same three
// states for the azure_managed_redis app credentials: omitted when unchanged,
// explicit null when the pair is dropped (falling back to the access-manager's
// default Azure credentials), and the value when set. midgard constrains both
// halves to a non-empty string, so dropping them must send null.
func TestUpdateRedisInput_AzureAppCredentialMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateRedisAccessCredentialInput
		contains []string
		omits    bool
	}{
		{
			name:  "unchanged omits both fields",
			input: UpdateRedisAccessCredentialInput{},
			omits: true,
		},
		{
			name: "dropped pair sends explicit nulls",
			input: UpdateRedisAccessCredentialInput{
				ClientID:     NewNullableString(""),
				ClientSecret: NewNullableString(""),
			},
			contains: []string{`"client_id":null`, `"client_secret":null`},
		},
		{
			name: "set sends the values",
			input: UpdateRedisAccessCredentialInput{
				ClientID:     NewNullableString("11111111-1111-1111-1111-111111111111"),
				ClientSecret: NewNullableString("secret"),
			},
			contains: []string{
				`"client_id":"11111111-1111-1111-1111-111111111111"`,
				`"client_secret":"secret"`,
			},
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
				for _, field := range []string{"client_id", "client_secret"} {
					if strings.Contains(got, field) {
						t.Fatalf("expected %s to be omitted, got %s", field, got)
					}
				}
				return
			}
			for _, want := range tc.contains {
				if !strings.Contains(got, want) {
					t.Fatalf("expected %s to contain %s", got, want)
				}
			}
		})
	}
}
