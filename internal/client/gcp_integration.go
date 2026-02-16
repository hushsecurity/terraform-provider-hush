package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const gcpIntegrationsEndpoint = "/v1/integrations/gcp"

// GCPProject represents a GCP project in an integration
type GCPProject struct {
	ProjectID      string  `json:"project_id"`
	Enabled        bool    `json:"enabled"`
	DisplayName    *string `json:"display_name,omitempty"`
	State          string  `json:"state,omitempty"`
	OrganizationID *string `json:"organization_id,omitempty"`
}

// GCPFeature represents a feature in a GCP integration
type GCPFeature struct {
	Name         string  `json:"name"`
	Enabled      bool    `json:"enabled"`
	State        string  `json:"state,omitempty"`
	StateMessage *string `json:"state_message,omitempty"`
}

// GCPIntegration represents a GCP integration response
type GCPIntegration struct {
	ID                  string       `json:"id,omitempty"`
	Name                string       `json:"name"`
	Description         string       `json:"description,omitempty"`
	Status              string       `json:"status,omitempty"`
	Type                string       `json:"type,omitempty"`
	ServiceAccountEmail *string      `json:"service_account_email,omitempty"`
	Projects            []GCPProject `json:"projects,omitempty"`
	Features            []GCPFeature `json:"features,omitempty"`
}

// CreateGCPIntegrationInput represents the input for creating a GCP integration
type CreateGCPIntegrationInput struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Projects    []GCPProjectInput `json:"projects,omitempty"`
	Features    []GCPFeatureInput `json:"features,omitempty"`
}

// GCPProjectInput represents a project in create/update input
type GCPProjectInput struct {
	ProjectID string `json:"project_id"`
	Enabled   bool   `json:"enabled"`
}

// GCPFeatureInput represents a feature in create/update input
type GCPFeatureInput struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// CompleteGCPIntegrationInput represents the input for completing a GCP integration
type CompleteGCPIntegrationInput struct {
	ServiceAccountEmail string `json:"service_account_email"`
}

// UpdateGCPIntegrationInput represents the input for updating a GCP integration
type UpdateGCPIntegrationInput struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Projects    *[]GCPProjectInput `json:"projects,omitempty"`
	Features    *[]GCPFeatureInput `json:"features,omitempty"`
}

// GCPIntegrationListResponse represents the response from listing GCP integrations
type GCPIntegrationListResponse struct {
	Items      []GCPIntegration `json:"items"`
	NextCursor *string          `json:"next_cursor"`
	HasMore    bool             `json:"has_more"`
}

// CreateGCPIntegration creates a new GCP integration (status=pending)
func CreateGCPIntegration(ctx context.Context, c *Client, input *CreateGCPIntegrationInput) (*GCPIntegration, error) {
	var resp GCPIntegration
	if err := c.doRequest(ctx, http.MethodPost, gcpIntegrationsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CompleteGCPIntegration completes a pending GCP integration by providing the service account email
func CompleteGCPIntegration(ctx context.Context, c *Client, id string, input *CompleteGCPIntegrationInput) (*GCPIntegration, error) {
	path := fmt.Sprintf("/v1/integrations/%s/gcp", id)
	var resp GCPIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetGCPIntegration retrieves a GCP integration by ID
func GetGCPIntegration(ctx context.Context, c *Client, id string) (*GCPIntegration, error) {
	path := fmt.Sprintf("/v1/integrations/%s/gcp", id)
	var integ GCPIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &integ); err != nil {
		return nil, err
	}
	return &integ, nil
}

// UpdateGCPIntegration updates a GCP integration
func UpdateGCPIntegration(ctx context.Context, c *Client, id string, input *UpdateGCPIntegrationInput) (*GCPIntegration, error) {
	path := fmt.Sprintf("/v1/integrations/%s/gcp", id)
	var result GCPIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGCPIntegration deletes a GCP integration
func DeleteGCPIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("/v1/integrations/%s/gcp", id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetGCPIntegrationOnboardingScript retrieves the onboarding script for a GCP integration
func GetGCPIntegrationOnboardingScript(ctx context.Context, c *Client, id string) (string, error) {
	path := fmt.Sprintf("/v1/integrations/%s/gcp/onboarding_script", id)
	var script string
	// The onboarding script endpoint returns plain text, not JSON.
	// We use doRequest with nil result and handle the response manually.
	// Actually, we need a special approach since doRequest expects JSON.
	// Use a raw string response wrapper.
	if err := c.doRawRequest(ctx, http.MethodGet, path, nil, &script); err != nil {
		return "", err
	}
	return script, nil
}

// GetGCPIntegrationsByName retrieves GCP integrations by name
// Note: The API may return partial matches, so we filter for exact name matches
func GetGCPIntegrationsByName(ctx context.Context, c *Client, name string) ([]GCPIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("/v1/integrations?type=gcp&name=%s", encodedName)

	var resp GCPIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	// Filter for exact name matches since API may return partial matches
	var exactMatches []GCPIntegration
	for _, integ := range resp.Items {
		if integ.Name == name {
			exactMatches = append(exactMatches, integ)
		}
	}

	return exactMatches, nil
}
