package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const accessCredentialsEndpoint = "/v1/access_credentials"

type AccessCredentialType string

const (
	AccessCredentialTypePlaintext AccessCredentialType = "plaintext"
	AccessCredentialTypeKV        AccessCredentialType = "kv"
)

type AccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	DeploymentIDs []string             `json:"deployment_ids"`
	Keys          []string             `json:"keys,omitempty"`
	CreatedAt     string               `json:"created_at,omitempty"`
	ModifiedAt    string               `json:"modified_at,omitempty"`
	CreatedBy     string               `json:"created_by,omitempty"`
}

type PlaintextAccessCredential struct {
	AccessCredential
	Secret string `json:"secret,omitempty"` // Only present in create requests
}

type KVItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KVAccessCredential struct {
	AccessCredential
	Items []KVItem `json:"items,omitempty"` // Only present in create requests
}

type CreatePlaintextAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Secret        string   `json:"secret"`
}

type CreateKVAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Items         []KVItem `json:"items"`
}

type UpdateAccessCredentialInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AccessCredentialListResponse struct {
	Items      []AccessCredential `json:"items"`
	Total      int                `json:"total"`
	HasNext    bool               `json:"has_next"`
	NextCursor *string            `json:"next_cursor"`
}

func CreatePlaintextAccessCredential(ctx context.Context, c *Client, input *CreatePlaintextAccessCredentialInput) (*AccessCredential, error) {
	path := accessCredentialsEndpoint + "/plaintext"
	var resp AccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func CreateKVAccessCredential(ctx context.Context, c *Client, input *CreateKVAccessCredentialInput) (*AccessCredential, error) {
	path := accessCredentialsEndpoint + "/kv"
	var resp AccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAccessCredential(ctx context.Context, c *Client, id string) (*AccessCredential, error) {
	path := fmt.Sprintf("%s/%s", accessCredentialsEndpoint, id)
	var cred AccessCredential
	err := c.doRequest(ctx, http.MethodGet, path, nil, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func GetPlaintextAccessCredential(ctx context.Context, c *Client, id string) (*AccessCredential, error) {
	path := fmt.Sprintf("%s/plaintext/%s", accessCredentialsEndpoint, id)
	var cred AccessCredential
	err := c.doRequest(ctx, http.MethodGet, path, nil, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func GetKVAccessCredential(ctx context.Context, c *Client, id string) (*AccessCredential, error) {
	path := fmt.Sprintf("%s/kv/%s", accessCredentialsEndpoint, id)
	var cred AccessCredential
	err := c.doRequest(ctx, http.MethodGet, path, nil, &cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func UpdatePlaintextAccessCredential(ctx context.Context, c *Client, id string, input *UpdateAccessCredentialInput) (*AccessCredential, error) {
	path := fmt.Sprintf("%s/plaintext/%s", accessCredentialsEndpoint, id)
	var result AccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func UpdateKVAccessCredential(ctx context.Context, c *Client, id string, input *UpdateAccessCredentialInput) (*AccessCredential, error) {
	path := fmt.Sprintf("%s/kv/%s", accessCredentialsEndpoint, id)
	var result AccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteAccessCredential(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", accessCredentialsEndpoint, id)
	err := c.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func ListAccessCredentials(ctx context.Context, c *Client, credType *AccessCredentialType) (*AccessCredentialListResponse, error) {
	params := url.Values{}
	if credType != nil {
		params.Set("type", string(*credType))
	}

	path := accessCredentialsEndpoint
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp AccessCredentialListResponse
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
