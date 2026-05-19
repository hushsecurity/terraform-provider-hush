package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const integrationsEndpoint = "/v1/integrations"

// Integration status constants
const (
	IntegrationStatusOK                  = "ok"
	IntegrationStatusDisabled            = "disabled"
	IntegrationStatusSuspended           = "suspended"
	IntegrationStatusError               = "error"
	IntegrationStatusWarning             = "warning"
	IntegrationStatusDeleted             = "deleted"
	IntegrationStatusPending             = "pending"
	IntegrationStatusPendingRegistration = "pending_registration"
)

// GitLab Integration

type GitlabIntegration struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	Status             string   `json:"status,omitempty"`
	StatusMessage      string   `json:"status_message,omitempty"`
	Type               string   `json:"type,omitempty"`
	OnpremDeploymentID string   `json:"onprem_deployment_id,omitempty"`
	GroupID            *int     `json:"group_id,omitempty"`
	ProjectID          *int     `json:"project_id,omitempty"`
	Group              string   `json:"group,omitempty"`
	Visibilities       []string `json:"visibilities,omitempty"`
	BaseURL            string   `json:"base_url,omitempty"`
	SelectedRepos      []string `json:"selected_repos,omitempty"`
	BotName            string   `json:"bot_name,omitempty"`
	EnablePRScans      *bool    `json:"enable_pr_scans,omitempty"`
	CreatedAt          string   `json:"created_at,omitempty"`
	ModifiedAt         string   `json:"modified_at,omitempty"`
}

type CreateGitlabIntegrationInput struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	OnpremDeploymentID string   `json:"onprem_deployment_id,omitempty"`
	Token              string   `json:"token"`
	GroupID            *int     `json:"group_id,omitempty"`
	ProjectID          *int     `json:"project_id,omitempty"`
	Visibilities       []string `json:"visibilities,omitempty"`
	BaseURL            string   `json:"base_url,omitempty"`
	SelectedRepos      []string `json:"selected_repos,omitempty"`
	EnablePRScans      *bool    `json:"enable_pr_scans,omitempty"`
}

type UpdateGitlabIntegrationInput struct {
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	OnpremDeploymentID *string  `json:"onprem_deployment_id,omitempty"`
	Visibilities       []string `json:"visibilities,omitempty"`
	BaseURL            *string  `json:"base_url,omitempty"`
	SelectedRepos      []string `json:"selected_repos,omitempty"`
	EnablePRScans      *bool    `json:"enable_pr_scans,omitempty"`
}

type ReplaceGitlabTokenInput struct {
	Token string `json:"token"`
}

type GitlabIntegrationListResponse struct {
	Items        []GitlabIntegration `json:"items"`
	PageNumber   int                 `json:"page_number"`
	Total        *int                `json:"total"`
	PreviousPage *string             `json:"previous_page"`
	NextPage     *string             `json:"next_page"`
}

func CreateGitlabIntegration(ctx context.Context, c *Client, input *CreateGitlabIntegrationInput) (*GitlabIntegration, error) {
	path := fmt.Sprintf("%s/gitlab", integrationsEndpoint)
	var resp GitlabIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGitlabIntegration(ctx context.Context, c *Client, id string) (*GitlabIntegration, error) {
	path := fmt.Sprintf("%s/%s/gitlab", integrationsEndpoint, id)
	var resp GitlabIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGitlabIntegrationsByName(ctx context.Context, c *Client, name string) ([]GitlabIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=gitlab", integrationsEndpoint, encodedName)
	var resp GitlabIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateGitlabIntegration(ctx context.Context, c *Client, id string, input *UpdateGitlabIntegrationInput) (*GitlabIntegration, error) {
	path := fmt.Sprintf("%s/%s/gitlab", integrationsEndpoint, id)
	var resp GitlabIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteGitlabIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceGitlabToken(ctx context.Context, c *Client, id string, input *ReplaceGitlabTokenInput) error {
	path := fmt.Sprintf("%s/%s/gitlab/token", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}
