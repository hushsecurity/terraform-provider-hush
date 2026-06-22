package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const deploymentsEndpoint = "/v1/deployments"

// OidcConfig configures passwordless deployment token exchange via OIDC.
type OidcConfig struct {
	Issuer          string   `json:"issuer"`
	Audience        string   `json:"audience"`
	AllowedSubjects []string `json:"allowed_subjects,omitempty"`
}

type Deployment struct {
	ID           string      `json:"id,omitempty"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	EnvType      string      `json:"env_type"`
	Status       string      `json:"status,omitempty"`
	Kind         string      `json:"kind,omitempty"`
	OidcProvider *OidcConfig `json:"oidc_provider,omitempty"`
}

// CreateDeploymentInput represents the input for creating a deployment
type CreateDeploymentInput struct {
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	EnvType      string      `json:"env_type"`
	Kind         string      `json:"kind,omitempty"`
	OidcProvider *OidcConfig `json:"oidc_provider,omitempty"`
}

// UpdateDeploymentInput represents the input for updating a deployment. Each
// scalar is a pointer with omitempty so only changed fields are sent. The
// oidc_provider field needs three states -- omitted (unchanged), set, and
// explicit null (removed) -- which omitempty alone cannot express, so it uses
// the oidcProviderUpdate wrapper.
type UpdateDeploymentInput struct {
	Name         *string             `json:"name,omitempty"`
	Description  *string             `json:"description,omitempty"`
	EnvType      *string             `json:"env_type,omitempty"`
	Kind         *string             `json:"kind,omitempty"`
	OidcProvider *oidcProviderUpdate `json:"oidc_provider,omitempty"`
}

// oidcProviderUpdate marshals to null when Config is nil (removal) and to the
// config object otherwise. A nil wrapper on the input is omitted (no change).
type oidcProviderUpdate struct{ Config *OidcConfig }

func (o oidcProviderUpdate) MarshalJSON() ([]byte, error) {
	if o.Config == nil {
		return []byte("null"), nil
	}
	return json.Marshal(o.Config)
}

// NewOidcProviderUpdate wraps an OIDC config (possibly nil for removal) for an
// update request, forcing the oidc_provider field to be sent.
func NewOidcProviderUpdate(config *OidcConfig) *oidcProviderUpdate {
	return &oidcProviderUpdate{Config: config}
}

// DeploymentCredentialsResponse embeds Deployment and adds credentials
type DeploymentCredentialsResponse struct {
	Deployment
	Token           string `json:"token"`
	Password        string `json:"password"`
	ImagePullSecret string `json:"image_pull_secret"`
}

func CreateDeployment(ctx context.Context, c *Client, input *CreateDeploymentInput) (*Deployment, error) {
	var resp DeploymentCredentialsResponse
	if err := c.doRequest(ctx, http.MethodPost, deploymentsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp.Deployment, nil
}

// CreateDeploymentWithCredentials creates a deployment and returns the full credentials response
func CreateDeploymentWithCredentials(ctx context.Context, c *Client, input *CreateDeploymentInput) (*DeploymentCredentialsResponse, error) {
	var resp DeploymentCredentialsResponse
	if err := c.doRequest(ctx, http.MethodPost, deploymentsEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetDeployment(ctx context.Context, c *Client, id string) (*Deployment, error) {
	path := fmt.Sprintf("%s/%s", deploymentsEndpoint, id)
	var dep Deployment
	err := c.doRequest(ctx, http.MethodGet, path, nil, &dep)
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func UpdateDeployment(ctx context.Context, c *Client, id string, input *UpdateDeploymentInput) (*Deployment, error) {
	path := fmt.Sprintf("%s/%s", deploymentsEndpoint, id)
	var result Deployment
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteDeployment(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", deploymentsEndpoint, id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// DeploymentListResponse represents the response from listing deployments
type DeploymentListResponse struct {
	Items      []Deployment `json:"items"`
	NextCursor *string      `json:"next_cursor"`
	HasMore    bool         `json:"has_more"`
}

// GetDeploymentsByName retrieves deployments by name
func GetDeploymentsByName(ctx context.Context, c *Client, name string) ([]Deployment, error) {
	// URL encode the name parameter to handle special characters
	encodedName := url.QueryEscape(name)
	path := fmt.Sprintf("%s?name=%s", deploymentsEndpoint, encodedName)

	var resp DeploymentListResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// AccessBridgeStatus represents the response from the access_bridge endpoint
type AccessBridgeStatus struct {
	Status string `json:"status"`
}

const AccessBridgeStatusOk = "Ok"

// GetAccessBridgeStatus retrieves the access bridge status for a deployment
func GetAccessBridgeStatus(ctx context.Context, c *Client, deploymentID string) (*AccessBridgeStatus, error) {
	path := fmt.Sprintf("%s/%s/access_bridge", deploymentsEndpoint, deploymentID)
	var resp AccessBridgeStatus
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// WaitForAccessBridge polls the access bridge status until it becomes "Ok" or times out
func WaitForAccessBridge(ctx context.Context, c *Client, deploymentID string) error {
	return waitForStatus(ctx, func() (status, statusDetail string, err error) {
		resp, err := GetAccessBridgeStatus(ctx, c, deploymentID)
		if err != nil {
			return "", "", err
		}
		// Map bridge status to the waitForStatus terminal states
		if resp.Status == AccessBridgeStatusOk {
			return "ok", "", nil
		}
		// Non-terminal: keep polling (e.g. "down", "disconnected")
		return resp.Status, "", nil
	})
}
