package client

import (
	"context"
	"testing"
)

func TestNewClient_ValidatesParameters(t *testing.T) {
	tests := []struct {
		name        string
		keyID       string
		keySecret   string
		baseURL     string
		expectError bool
	}{
		{
			name:        "empty key ID",
			keyID:       "",
			keySecret:   "secret",
			baseURL:     "https://api.eu.dev.hush-security.com",
			expectError: true,
		},
		{
			name:        "empty key secret",
			keyID:       "key123",
			keySecret:   "",
			baseURL:     "https://api.eu.dev.hush-security.com",
			expectError: true,
		},
		{
			name:        "valid parameters with custom baseURL",
			keyID:       "key123",
			keySecret:   "secret",
			baseURL:     "https://api.eu.dev.hush-security.com",
			expectError: true, // Will fail due to invalid credentials, but validates parameter handling
		},
		{
			name:        "valid parameters with empty baseURL uses default",
			keyID:       "key123",
			keySecret:   "secret",
			baseURL:     "",
			expectError: true, // Will fail due to invalid credentials, but validates parameter handling
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(context.Background(), tt.keyID, tt.keySecret, tt.baseURL)

			// For empty credentials, we expect immediate validation errors
			if (tt.keyID == "" || tt.keySecret == "") && err == nil {
				t.Errorf("expected error for missing credentials but got none")
			}

			// For valid credentials (but fake), we expect authentication errors
			if tt.keyID != "" && tt.keySecret != "" && err == nil {
				t.Errorf("expected authentication error for fake credentials but got none")
			}
		})
	}
}
