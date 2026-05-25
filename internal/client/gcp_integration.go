package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GCP Integration

type GCPProjectInput struct {
	ProjectID string `json:"project_id"`
	Enabled   bool   `json:"enabled"`
}

type GCPProject struct {
	ProjectID      string `json:"project_id"`
	DisplayName    string `json:"display_name,omitempty"`
	State          string `json:"state,omitempty"`
	StateMessage   string `json:"state_message,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	Enabled        bool   `json:"enabled"`
}

type GCPFeatureInput struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type GCPFeature struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	State        string `json:"state,omitempty"`
	StateMessage string `json:"state_message,omitempty"`
}

type GCPIntegration struct {
	ID                    string       `json:"id,omitempty"`
	Name                  string       `json:"name"`
	Description           string       `json:"description,omitempty"`
	Status                string       `json:"status,omitempty"`
	StatusMessage         string       `json:"status_message,omitempty"`
	StatusAt              string       `json:"status_at,omitempty"`
	Type                  string       `json:"type,omitempty"`
	OnpremDeploymentID    string       `json:"onprem_deployment_id,omitempty"`
	ServiceAccountEmail   string       `json:"service_account_email,omitempty"`
	Projects              []GCPProject `json:"projects,omitempty"`
	Features              []GCPFeature `json:"features,omitempty"`
	CreatedAt             string       `json:"created_at,omitempty"`
	ModifiedAt            string       `json:"modified_at,omitempty"`
	NextRescanAt          string       `json:"next_rescan_at,omitempty"`
	NextFullScanAt        string       `json:"next_full_scan_at,omitempty"`
	NextPeriodicChecksAt  string       `json:"next_periodic_checks_at,omitempty"`
	NextUpdateResourcesAt string       `json:"next_update_resources_at,omitempty"`
}

type CreateGCPIntegrationInput struct {
	Name               string            `json:"name"`
	Description        string            `json:"description,omitempty"`
	OnpremDeploymentID string            `json:"onprem_deployment_id,omitempty"`
	Projects           []GCPProjectInput `json:"projects,omitempty"`
	Features           []GCPFeatureInput `json:"features,omitempty"`
}

type UpdateGCPIntegrationInput struct {
	Name               *string           `json:"name,omitempty"`
	Description        *string           `json:"description,omitempty"`
	OnpremDeploymentID *string           `json:"onprem_deployment_id,omitempty"`
	Projects           []GCPProjectInput `json:"projects,omitempty"`
	Features           []GCPFeatureInput `json:"features,omitempty"`
}

type CompleteGCPIntegrationInput struct {
	ServiceAccountEmail string `json:"service_account_email"`
}

type GCPIntegrationListResponse struct {
	Items        []GCPIntegration `json:"items"`
	PageNumber   int              `json:"page_number"`
	Total        *int             `json:"total"`
	PreviousPage *string          `json:"previous_page"`
	NextPage     *string          `json:"next_page"`
}

func CreateGCPIntegration(ctx context.Context, c *Client, input *CreateGCPIntegrationInput) (*GCPIntegration, error) {
	path := fmt.Sprintf("%s/gcp", integrationsEndpoint)
	var resp GCPIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGCPIntegration(ctx context.Context, c *Client, id string) (*GCPIntegration, error) {
	path := fmt.Sprintf("%s/%s/gcp", integrationsEndpoint, id)
	var resp GCPIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGCPIntegrationsByName(ctx context.Context, c *Client, name string) ([]GCPIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=gcp", integrationsEndpoint, encodedName)
	var resp GCPIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateGCPIntegration(ctx context.Context, c *Client, id string, input *UpdateGCPIntegrationInput) (*GCPIntegration, error) {
	path := fmt.Sprintf("%s/%s/gcp", integrationsEndpoint, id)
	var resp GCPIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteGCPIntegration(ctx context.Context, c *Client, id string) error {
	// GCP uses type-specific delete endpoint
	path := fmt.Sprintf("%s/%s/gcp", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func CompleteGCPIntegration(ctx context.Context, c *Client, id string, input *CompleteGCPIntegrationInput) (*GCPIntegration, error) {
	path := fmt.Sprintf("%s/%s/gcp", integrationsEndpoint, id)
	var resp GCPIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
