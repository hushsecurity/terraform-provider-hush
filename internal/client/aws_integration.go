package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AWS Integration

type AWSFeature struct {
	Name           string   `json:"name"`
	State          string   `json:"state"`
	StateMessage   string   `json:"state_message,omitempty"`
	AllowedRegions []string `json:"allowed_regions,omitempty"`
}

type AWSIntegration struct {
	ID                    string       `json:"id,omitempty"`
	Name                  string       `json:"name"`
	Description           string       `json:"description,omitempty"`
	Status                string       `json:"status,omitempty"`
	StatusMessage         string       `json:"status_message,omitempty"`
	StatusAt              string       `json:"status_at,omitempty"`
	Type                  string       `json:"type,omitempty"`
	OnpremDeploymentID    string       `json:"onprem_deployment_id,omitempty"`
	RoleArn               string       `json:"role_arn,omitempty"`
	CfStacksetArn         string       `json:"cf_stackset_arn,omitempty"`
	CfStackID             string       `json:"cf_stack_id,omitempty"`
	UniqueSuffix          string       `json:"unique_suffix,omitempty"`
	AccountIDs            []string     `json:"account_ids,omitempty"`
	CreatedBy             string       `json:"created_by,omitempty"`
	Features              []AWSFeature `json:"features,omitempty"`
	Version               string       `json:"version,omitempty"`
	CreatedAt             string       `json:"created_at,omitempty"`
	ModifiedAt            string       `json:"modified_at,omitempty"`
	NextRescanAt          string       `json:"next_rescan_at,omitempty"`
	NextFullScanAt        string       `json:"next_full_scan_at,omitempty"`
	NextPeriodicChecksAt  string       `json:"next_periodic_checks_at,omitempty"`
	NextUpdateResourcesAt string       `json:"next_update_resources_at,omitempty"`
}

type CreateAWSIntegrationInput struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	RoleArn       string `json:"role_arn,omitempty"`
	CfStacksetArn string `json:"cf_stackset_arn,omitempty"`
	UniqueSuffix  string `json:"unique_suffix,omitempty"`
}

type UpdateAWSIntegrationInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AWSIntegrationListResponse struct {
	Items        []AWSIntegration `json:"items"`
	PageNumber   int              `json:"page_number"`
	Total        *int             `json:"total"`
	PreviousPage *string          `json:"previous_page"`
	NextPage     *string          `json:"next_page"`
}

func CreateAWSIntegration(ctx context.Context, c *Client, input *CreateAWSIntegrationInput) (*AWSIntegration, error) {
	path := fmt.Sprintf("%s/aws", integrationsEndpoint)

	maxRetries := 6
	baseDelay := 10 * time.Second

	var lastErr error
	for attempt := range maxRetries {
		var resp AWSIntegration
		if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "Failed to assume") || strings.Contains(errMsg, "Failed to get IAM role") {
				lastErr = err
				delay := baseDelay * time.Duration(1<<uint(attempt))
				if delay > 60*time.Second {
					delay = 60 * time.Second
				}
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled while waiting for IAM propagation: %w", ctx.Err())
				case <-time.After(delay):
					continue
				}
			}
			return nil, err
		}
		return &resp, nil
	}
	return nil, fmt.Errorf("failed after %d retries waiting for IAM propagation: %w", maxRetries, lastErr)
}

func GetAWSIntegration(ctx context.Context, c *Client, id string) (*AWSIntegration, error) {
	path := fmt.Sprintf("%s/%s/aws", integrationsEndpoint, id)
	var resp AWSIntegration
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAWSIntegrationsByName(ctx context.Context, c *Client, name string) ([]AWSIntegration, error) {
	base := fmt.Sprintf("%s?name=%s&type=aws", integrationsEndpoint, url.QueryEscape(name))
	return collectPages(func(cursor string) ([]AWSIntegration, *string, error) {
		var resp AWSIntegrationListResponse
		if err := c.doRequest(ctx, http.MethodGet, withCursor(base, cursor), nil, &resp); err != nil {
			return nil, nil, err
		}
		return resp.Items, resp.NextPage, nil
	})
}

func UpdateAWSIntegration(ctx context.Context, c *Client, id string, input *UpdateAWSIntegrationInput) (*AWSIntegration, error) {
	path := fmt.Sprintf("%s/%s/aws", integrationsEndpoint, id)
	var resp AWSIntegration
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteAWSIntegration(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", integrationsEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
