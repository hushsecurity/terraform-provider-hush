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

// Confluence Integration

type ConfluenceIntegration struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Status             string `json:"status,omitempty"`
	StatusMessage      string `json:"status_message,omitempty"`
	StatusAt           string `json:"status_at,omitempty"`
	Type               string `json:"type,omitempty"`
	OrgDomain          string `json:"org_domain"`
	OnpremDeploymentID string `json:"onprem_deployment_id,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	ModifiedAt         string `json:"modified_at,omitempty"`
	NextRescanAt       string `json:"next_rescan_at,omitempty"`
}

type CreateConfluenceIntegrationInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	OrgDomain   string `json:"org_domain"`
	User        string `json:"user"`
	ApiKey      string `json:"api_key"`
}

type UpdateConfluenceIntegrationInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ReplaceConfluenceApiKeyInput struct {
	User   string `json:"user"`
	ApiKey string `json:"api_key"`
}

type ConfluenceIntegrationListResponse struct {
	Items        []ConfluenceIntegration `json:"items"`
	PageNumber   int                     `json:"page_number"`
	Total        *int                    `json:"total"`
	PreviousPage *string                 `json:"previous_page"`
	NextPage     *string                 `json:"next_page"`
}

func CreateConfluenceIntegration(ctx context.Context, c *Client, input *CreateConfluenceIntegrationInput) (*ConfluenceIntegration, error) {
	path := fmt.Sprintf("%s/confluence", integrationsEndpoint)
	var resp ConfluenceIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetConfluenceIntegration(ctx context.Context, c *Client, id string) (*ConfluenceIntegration, error) {
	path := fmt.Sprintf("%s/%s/confluence", integrationsEndpoint, id)
	var resp ConfluenceIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetConfluenceIntegrationsByName(ctx context.Context, c *Client, name string) ([]ConfluenceIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=confluence", integrationsEndpoint, encodedName)
	var resp ConfluenceIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateConfluenceIntegration(ctx context.Context, c *Client, id string, input *UpdateConfluenceIntegrationInput) (*ConfluenceIntegration, error) {
	path := fmt.Sprintf("%s/%s/confluence", integrationsEndpoint, id)
	var resp ConfluenceIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteConfluenceIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceConfluenceApiKey(ctx context.Context, c *Client, id string, input *ReplaceConfluenceApiKeyInput) error {
	path := fmt.Sprintf("%s/%s/confluence/api_key", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}

// Jira Integration

type JiraIntegration struct {
	ID                   string `json:"id,omitempty"`
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	Status               string `json:"status,omitempty"`
	StatusMessage        string `json:"status_message,omitempty"`
	Type                 string `json:"type,omitempty"`
	OrgDomain            string `json:"org_domain"`
	OnpremDeploymentID   string `json:"onprem_deployment_id,omitempty"`
	SyncIssuesResolution *bool  `json:"sync_issues_resolution,omitempty"`
	EnableScans          *bool  `json:"enable_scans,omitempty"`
	WebhookProvisioned   bool   `json:"webhook_provisioned,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	ModifiedAt           string `json:"modified_at,omitempty"`
}

type CreateJiraIntegrationInput struct {
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	OrgDomain            string `json:"org_domain"`
	User                 string `json:"user"`
	ApiKey               string `json:"api_key"`
	SyncIssuesResolution *bool  `json:"sync_issues_resolution,omitempty"`
	EnableScans          *bool  `json:"enable_scans,omitempty"`
}

type UpdateJiraIntegrationInput struct {
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	OrgDomain            *string `json:"org_domain,omitempty"`
	SyncIssuesResolution *bool   `json:"sync_issues_resolution,omitempty"`
}

type ReplaceJiraApiKeyInput struct {
	User   string `json:"user"`
	ApiKey string `json:"api_key"`
}

type JiraIntegrationListResponse struct {
	Items        []JiraIntegration `json:"items"`
	PageNumber   int               `json:"page_number"`
	Total        *int              `json:"total"`
	PreviousPage *string           `json:"previous_page"`
	NextPage     *string           `json:"next_page"`
}

func CreateJiraIntegration(ctx context.Context, c *Client, input *CreateJiraIntegrationInput) (*JiraIntegration, error) {
	path := fmt.Sprintf("%s/jira", integrationsEndpoint)
	var resp JiraIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetJiraIntegration(ctx context.Context, c *Client, id string) (*JiraIntegration, error) {
	path := fmt.Sprintf("%s/%s/jira", integrationsEndpoint, id)
	var resp JiraIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetJiraIntegrationsByName(ctx context.Context, c *Client, name string) ([]JiraIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=jira", integrationsEndpoint, encodedName)
	var resp JiraIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateJiraIntegration(ctx context.Context, c *Client, id string, input *UpdateJiraIntegrationInput) (*JiraIntegration, error) {
	path := fmt.Sprintf("%s/%s/jira", integrationsEndpoint, id)
	var resp JiraIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteJiraIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceJiraApiKey(ctx context.Context, c *Client, id string, input *ReplaceJiraApiKeyInput) error {
	path := fmt.Sprintf("%s/%s/jira/api_key", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}

// Bitbucket Integration

type BitbucketIntegration struct {
	ID                    string `json:"id,omitempty"`
	Name                  string `json:"name"`
	Description           string `json:"description,omitempty"`
	Status                string `json:"status,omitempty"`
	StatusMessage         string `json:"status_message,omitempty"`
	StatusAt              string `json:"status_at,omitempty"`
	Type                  string `json:"type,omitempty"`
	OnpremDeploymentID    string `json:"onprem_deployment_id,omitempty"`
	WorkspaceSlug         string `json:"workspace_slug"`
	CreatedAt             string `json:"created_at,omitempty"`
	ModifiedAt            string `json:"modified_at,omitempty"`
	NextRescanAt          string `json:"next_rescan_at,omitempty"`
	NextFullScanAt        string `json:"next_full_scan_at,omitempty"`
	NextPeriodicChecksAt  string `json:"next_periodic_checks_at,omitempty"`
	NextUpdateResourcesAt string `json:"next_update_resources_at,omitempty"`
}

type CreateBitbucketIntegrationInput struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OnpremDeploymentID string `json:"onprem_deployment_id,omitempty"`
	Token              string `json:"token"`
	WorkspaceSlug      string `json:"workspace_slug"`
}

type UpdateBitbucketIntegrationInput struct {
	Name               *string `json:"name,omitempty"`
	Description        *string `json:"description,omitempty"`
	OnpremDeploymentID *string `json:"onprem_deployment_id,omitempty"`
}

type ReplaceBitbucketTokenInput struct {
	Token string `json:"token"`
}

type BitbucketIntegrationListResponse struct {
	Items        []BitbucketIntegration `json:"items"`
	PageNumber   int                    `json:"page_number"`
	Total        *int                   `json:"total"`
	PreviousPage *string                `json:"previous_page"`
	NextPage     *string                `json:"next_page"`
}

func CreateBitbucketIntegration(ctx context.Context, c *Client, input *CreateBitbucketIntegrationInput) (*BitbucketIntegration, error) {
	path := fmt.Sprintf("%s/bitbucket", integrationsEndpoint)
	var resp BitbucketIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetBitbucketIntegration(ctx context.Context, c *Client, id string) (*BitbucketIntegration, error) {
	path := fmt.Sprintf("%s/%s/bitbucket", integrationsEndpoint, id)
	var resp BitbucketIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetBitbucketIntegrationsByName(ctx context.Context, c *Client, name string) ([]BitbucketIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=bitbucket", integrationsEndpoint, encodedName)
	var resp BitbucketIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateBitbucketIntegration(ctx context.Context, c *Client, id string, input *UpdateBitbucketIntegrationInput) (*BitbucketIntegration, error) {
	path := fmt.Sprintf("%s/%s/bitbucket", integrationsEndpoint, id)
	var resp BitbucketIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteBitbucketIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceBitbucketToken(ctx context.Context, c *Client, id string, input *ReplaceBitbucketTokenInput) error {
	path := fmt.Sprintf("%s/%s/bitbucket/token", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}

// Infisical Integration

type InfisicalIntegration struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Status             string `json:"status,omitempty"`
	StatusMessage      string `json:"status_message,omitempty"`
	Type               string `json:"type,omitempty"`
	OnpremDeploymentID string `json:"onprem_deployment_id,omitempty"`
	BaseURL            string `json:"base_url"`
	CreatedAt          string `json:"created_at,omitempty"`
	ModifiedAt         string `json:"modified_at,omitempty"`
}

type CreateInfisicalIntegrationInput struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OnpremDeploymentID string `json:"onprem_deployment_id,omitempty"`
	BaseURL            string `json:"base_url"`
	ClientID           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
}

type UpdateInfisicalIntegrationInput struct {
	Name               *string `json:"name,omitempty"`
	Description        *string `json:"description,omitempty"`
	OnpremDeploymentID *string `json:"onprem_deployment_id,omitempty"`
	BaseURL            *string `json:"base_url,omitempty"`
}

type ReplaceInfisicalCredentialsInput struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type InfisicalIntegrationListResponse struct {
	Items        []InfisicalIntegration `json:"items"`
	PageNumber   int                    `json:"page_number"`
	Total        *int                   `json:"total"`
	PreviousPage *string                `json:"previous_page"`
	NextPage     *string                `json:"next_page"`
}

func CreateInfisicalIntegration(ctx context.Context, c *Client, input *CreateInfisicalIntegrationInput) (*InfisicalIntegration, error) {
	path := fmt.Sprintf("%s/infisical", integrationsEndpoint)
	var resp InfisicalIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetInfisicalIntegration(ctx context.Context, c *Client, id string) (*InfisicalIntegration, error) {
	path := fmt.Sprintf("%s/%s/infisical", integrationsEndpoint, id)
	var resp InfisicalIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetInfisicalIntegrationsByName(ctx context.Context, c *Client, name string) ([]InfisicalIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=infisical", integrationsEndpoint, encodedName)
	var resp InfisicalIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateInfisicalIntegration(ctx context.Context, c *Client, id string, input *UpdateInfisicalIntegrationInput) (*InfisicalIntegration, error) {
	path := fmt.Sprintf("%s/%s/infisical", integrationsEndpoint, id)
	var resp InfisicalIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteInfisicalIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceInfisicalCredentials(ctx context.Context, c *Client, id string, input *ReplaceInfisicalCredentialsInput) error {
	path := fmt.Sprintf("%s/%s/infisical/credentials", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}

// Sonatype Integration

type SonatypeIntegration struct {
	ID                    string `json:"id,omitempty"`
	Name                  string `json:"name"`
	Description           string `json:"description,omitempty"`
	Status                string `json:"status,omitempty"`
	StatusMessage         string `json:"status_message,omitempty"`
	StatusAt              string `json:"status_at,omitempty"`
	Type                  string `json:"type,omitempty"`
	OnpremDeploymentID    string `json:"onprem_deployment_id,omitempty"`
	OrgURL                string `json:"org_url,omitempty"`
	WebhookProvisioned    bool   `json:"webhook_provisioned,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	ModifiedAt            string `json:"modified_at,omitempty"`
	NextRescanAt          string `json:"next_rescan_at,omitempty"`
	NextFullScanAt        string `json:"next_full_scan_at,omitempty"`
	NextPeriodicChecksAt  string `json:"next_periodic_checks_at,omitempty"`
	NextUpdateResourcesAt string `json:"next_update_resources_at,omitempty"`
}

type CreateSonatypeIntegrationInput struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OnpremDeploymentID string `json:"onprem_deployment_id,omitempty"`
	OrgURL             string `json:"org_url,omitempty"`
	User               string `json:"user"`
	ApiKey             string `json:"api_key"`
}

type UpdateSonatypeIntegrationInput struct {
	Name               *string `json:"name,omitempty"`
	Description        *string `json:"description,omitempty"`
	OnpremDeploymentID *string `json:"onprem_deployment_id,omitempty"`
	OrgURL             *string `json:"org_url,omitempty"`
}

type ReplaceSonatypeApiKeyInput struct {
	User   string `json:"user"`
	ApiKey string `json:"api_key"`
}

type SonatypeIntegrationListResponse struct {
	Items        []SonatypeIntegration `json:"items"`
	PageNumber   int                   `json:"page_number"`
	Total        *int                  `json:"total"`
	PreviousPage *string               `json:"previous_page"`
	NextPage     *string               `json:"next_page"`
}

func CreateSonatypeIntegration(ctx context.Context, c *Client, input *CreateSonatypeIntegrationInput) (*SonatypeIntegration, error) {
	path := fmt.Sprintf("%s/sonatype", integrationsEndpoint)
	var resp SonatypeIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetSonatypeIntegration(ctx context.Context, c *Client, id string) (*SonatypeIntegration, error) {
	path := fmt.Sprintf("%s/%s/sonatype", integrationsEndpoint, id)
	var resp SonatypeIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetSonatypeIntegrationsByName(ctx context.Context, c *Client, name string) ([]SonatypeIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=sonatype", integrationsEndpoint, encodedName)
	var resp SonatypeIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateSonatypeIntegration(ctx context.Context, c *Client, id string, input *UpdateSonatypeIntegrationInput) (*SonatypeIntegration, error) {
	path := fmt.Sprintf("%s/%s/sonatype", integrationsEndpoint, id)
	var resp SonatypeIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteSonatypeIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceSonatypeApiKey(ctx context.Context, c *Client, id string, input *ReplaceSonatypeApiKeyInput) error {
	path := fmt.Sprintf("%s/%s/sonatype/api_key", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}

// Artifactory Integration

type ArtifactoryIntegration struct {
	ID                    string `json:"id,omitempty"`
	Name                  string `json:"name"`
	Description           string `json:"description,omitempty"`
	Status                string `json:"status,omitempty"`
	StatusMessage         string `json:"status_message,omitempty"`
	StatusAt              string `json:"status_at,omitempty"`
	Type                  string `json:"type,omitempty"`
	OnpremDeploymentID    string `json:"onprem_deployment_id,omitempty"`
	OrgURL                string `json:"org_url,omitempty"`
	WebhookProvisioned    bool   `json:"webhook_provisioned,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	ModifiedAt            string `json:"modified_at,omitempty"`
	NextRescanAt          string `json:"next_rescan_at,omitempty"`
	NextFullScanAt        string `json:"next_full_scan_at,omitempty"`
	NextPeriodicChecksAt  string `json:"next_periodic_checks_at,omitempty"`
	NextUpdateResourcesAt string `json:"next_update_resources_at,omitempty"`
}

type CreateArtifactoryIntegrationInput struct {
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	OnpremDeploymentID string `json:"onprem_deployment_id,omitempty"`
	OrgURL             string `json:"org_url,omitempty"`
	Token              string `json:"token"`
}

type UpdateArtifactoryIntegrationInput struct {
	Name               *string `json:"name,omitempty"`
	Description        *string `json:"description,omitempty"`
	OnpremDeploymentID *string `json:"onprem_deployment_id,omitempty"`
	OrgURL             *string `json:"org_url,omitempty"`
}

type ReplaceArtifactoryTokenInput struct {
	Token string `json:"token"`
}

type ArtifactoryIntegrationListResponse struct {
	Items        []ArtifactoryIntegration `json:"items"`
	PageNumber   int                      `json:"page_number"`
	Total        *int                     `json:"total"`
	PreviousPage *string                  `json:"previous_page"`
	NextPage     *string                  `json:"next_page"`
}

func CreateArtifactoryIntegration(ctx context.Context, c *Client, input *CreateArtifactoryIntegrationInput) (*ArtifactoryIntegration, error) {
	path := fmt.Sprintf("%s/artifactory", integrationsEndpoint)
	var resp ArtifactoryIntegration
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetArtifactoryIntegration(ctx context.Context, c *Client, id string) (*ArtifactoryIntegration, error) {
	path := fmt.Sprintf("%s/%s/artifactory", integrationsEndpoint, id)
	var resp ArtifactoryIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetArtifactoryIntegrationsByName(ctx context.Context, c *Client, name string) ([]ArtifactoryIntegration, error) {
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s&type=artifactory", integrationsEndpoint, encodedName)
	var resp ArtifactoryIntegrationListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func UpdateArtifactoryIntegration(ctx context.Context, c *Client, id string, input *UpdateArtifactoryIntegrationInput) (*ArtifactoryIntegration, error) {
	path := fmt.Sprintf("%s/%s/artifactory", integrationsEndpoint, id)
	var resp ArtifactoryIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteArtifactoryIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func ReplaceArtifactoryToken(ctx context.Context, c *Client, id string, input *ReplaceArtifactoryTokenInput) error {
	path := fmt.Sprintf("%s/%s/artifactory/token", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodPut, path, input, nil)
}
