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
