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

// TestUpdateDeploymentInput_OidcProvidersMarshaling verifies the three update
// states of oidc_providers. Removal is an explicit null rather than an empty
// list, because the API rejects [].
func TestUpdateDeploymentInput_OidcProvidersMarshaling(t *testing.T) {
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
			input:    UpdateDeploymentInput{OidcProviders: NewOidcProvidersUpdate(nil)},
			contains: `"oidc_providers":null`,
		},
		{
			name:     "an empty list is a removal, not an empty array",
			input:    UpdateDeploymentInput{OidcProviders: NewOidcProvidersUpdate([]OidcConfig{})},
			contains: `"oidc_providers":null`,
		},
		{
			name: "set sends every entry in order",
			input: UpdateDeploymentInput{OidcProviders: NewOidcProvidersUpdate([]OidcConfig{
				{
					Issuer:          "https://blue.example.com",
					Audience:        "hush",
					AllowedSubjects: []string{"system:serviceaccount:hush-security:*"},
				},
				{
					Issuer:   "https://green.example.com",
					Audience: "hush",
				},
			})},
			contains: `"oidc_providers":[{"issuer":"https://blue.example.com","audience":"hush","allowed_subjects":["system:serviceaccount:hush-security:*"]},{"issuer":"https://green.example.com","audience":"hush"}]`,
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
				if strings.Contains(got, "oidc_providers") {
					t.Fatalf("expected oidc_providers to be omitted, got %s", got)
				}
				return
			}
			if !strings.Contains(got, tc.contains) {
				t.Fatalf("expected %s to contain %s", got, tc.contains)
			}
		})
	}
}

// The API refuses a change that would leave a deployment holding both fields,
// so an update that writes the list has to clear the singular one in the same
// request: one value and one explicit null, never two values.
func TestUpdateDeploymentInput_WritesListAndClearsSingular(t *testing.T) {
	input := UpdateDeploymentInput{
		OidcProviders: NewOidcProvidersUpdate([]OidcConfig{
			{Issuer: "https://blue.example.com", Audience: "hush"},
		}),
		OidcProvider: NewOidcProviderUpdate(nil),
	}
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(b)
	if !strings.Contains(got, `"oidc_provider":null`) {
		t.Fatalf("expected the singular field cleared in %s", got)
	}
	if !strings.Contains(got, `"oidc_providers":[{"issuer":"https://blue.example.com"`) {
		t.Fatalf("expected the list written in %s", got)
	}
}

// A create never sends the singular field at all, so the pair cannot reach the
// API from that path either.
func TestCreateDeploymentInput_SendsOnlyTheList(t *testing.T) {
	b, err := json.Marshal(CreateDeploymentInput{
		Name:          "d",
		Kind:          "k8s",
		OidcProviders: []OidcConfig{{Issuer: "https://a.example.com", Audience: "hush"}},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(b)
	if strings.Contains(got, `"oidc_provider"`) {
		t.Fatalf("expected no oidc_provider in %s", got)
	}
	if !strings.Contains(got, `"oidc_providers":[`) {
		t.Fatalf("expected oidc_providers in %s", got)
	}
}

// TestUpdateDeploymentInput_AgwMarshaling verifies the three update states of
// agw: omitted when unchanged, explicit null when the gateway is
// decommissioned, and the object when a hostname is set or moved.
func TestUpdateDeploymentInput_AgwMarshaling(t *testing.T) {
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
			input:    UpdateDeploymentInput{Agw: NewAgwUpdate(nil)},
			contains: `"agw":null`,
		},
		{
			name: "set sends the object",
			input: UpdateDeploymentInput{
				Agw: NewAgwUpdate(&Agw{Hostname: "gw.example.com"}),
			},
			contains: `"agw":{"hostname":"gw.example.com"}`,
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
				if strings.Contains(got, "agw") {
					t.Fatalf("expected agw to be omitted, got %s", got)
				}
				return
			}
			if !strings.Contains(got, tc.contains) {
				t.Fatalf("expected %s to contain %s", got, tc.contains)
			}
		})
	}
}

// A create sends the agw object only when the deployment has a gateway: an
// empty object would be measured against the on-prem rules and refused for a
// missing hostname.
func TestCreateDeploymentInput_AgwOmitted(t *testing.T) {
	b, err := json.Marshal(CreateDeploymentInput{Name: "dep", Kind: "k8s"})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(b), "agw") {
		t.Fatalf("expected agw to be omitted, got %s", b)
	}
}

// The request model forbids unknown keys and refuses the member belonging to
// the other kind -- on a hosted deployment even a null hostname -- so whichever
// one is unset must not appear at all. gateway_url is derived and has no field
// on the way in, so it never appears either.
func TestAgwMarshalsOnlyTheMembersInUse(t *testing.T) {
	tests := []struct {
		name  string
		agw   Agw
		want  string
		omits []string
	}{
		{
			name:  "customer-run sends the hostname alone",
			agw:   Agw{Hostname: "gw.example.com"},
			want:  `{"hostname":"gw.example.com"}`,
			omits: []string{"region", "gateway_url"},
		},
		{
			name:  "hosted sends the region alone",
			agw:   Agw{Region: "fra"},
			want:  `{"region":"fra"}`,
			omits: []string{"hostname", "gateway_url"},
		},
		{
			name:  "a read-back url is never sent",
			agw:   Agw{Region: "fra", GatewayURL: "https://dep-1.agw.fra.example.com/mcp"},
			want:  `{"region":"fra"}`,
			omits: []string{"gateway_url"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.agw)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			got := string(b)
			if got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
			for _, key := range tc.omits {
				if strings.Contains(got, key) {
					t.Fatalf("expected %s to be omitted, got %s", key, got)
				}
			}
		})
	}
}
