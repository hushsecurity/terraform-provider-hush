package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const secretStoresEndpoint = "/v1/secret_stores"

// Secret store kinds (the config discriminator).
const (
	SecretStoreKindAWSSM      = "aws_sm"
	SecretStoreKindAWSSSM     = "aws_ssm"
	SecretStoreKindGCPSM      = "gcp_sm"
	SecretStoreKindK8sSecrets = "k8s_secrets"
)

// SecretStoreConfig is the backend's discriminated config union flattened into a
// single struct. Only the fields relevant to Kind are populated; the omitempty
// tags keep the rest out of the request so the strict backend model never sees a
// field belonging to another kind.
type SecretStoreConfig struct {
	Kind      string `json:"kind"`
	Prefix    string `json:"prefix"`
	Region    string `json:"region,omitempty"`
	KmsKeyID  string `json:"kms_key_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type SecretStore struct {
	ID            string            `json:"id,omitempty"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Config        SecretStoreConfig `json:"config"`
	DeploymentIDs []string          `json:"deployment_ids"`
	Status        string            `json:"status,omitempty"`
	StatusDetail  string            `json:"status_detail,omitempty"`
	CreatedAt     string            `json:"created_at,omitempty"`
	ModifiedAt    string            `json:"modified_at,omitempty"`
	CreatedBy     string            `json:"created_by,omitempty"`
	ModifiedBy    string            `json:"modified_by,omitempty"`
}

type CreateSecretStoreInput struct {
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	DeploymentIDs []string          `json:"deployment_ids"`
	Config        SecretStoreConfig `json:"config"`
}

// UpdateSecretStoreInput carries only the mutable fields. Config is immutable
// server-side, so it is absent here. DeploymentIDs is a pointer so an empty list
// (clearing all associations) can be distinguished from "leave unchanged".
type UpdateSecretStoreInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
}

// SecretStoreListResponse matches the backend CursorPage shape.
type SecretStoreListResponse struct {
	Items        []SecretStore `json:"items"`
	PageNumber   int           `json:"page_number"`
	Total        *int          `json:"total"`
	PreviousPage *string       `json:"previous_page"`
	NextPage     *string       `json:"next_page"`
}

func CreateSecretStore(ctx context.Context, c *Client, input *CreateSecretStoreInput) (*SecretStore, error) {
	var resp SecretStore
	if err := c.doRequest(ctx, http.MethodPost, secretStoresEndpoint, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetSecretStore(ctx context.Context, c *Client, id string) (*SecretStore, error) {
	path := fmt.Sprintf("%s/%s", secretStoresEndpoint, id)
	var resp SecretStore
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateSecretStore(ctx context.Context, c *Client, id string, input *UpdateSecretStoreInput) (*SecretStore, error) {
	path := fmt.Sprintf("%s/%s", secretStoresEndpoint, id)
	var resp SecretStore
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func DeleteSecretStore(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", secretStoresEndpoint, id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// GetSecretStoresByName returns all secret stores whose name matches, using the
// backend's server-side name filter and paging through every result. Names are not
// unique, so this may return more than one store.
func GetSecretStoresByName(ctx context.Context, c *Client, name string) ([]SecretStore, error) {
	base := fmt.Sprintf("%s?name=%s", secretStoresEndpoint, url.QueryEscape(name))
	return collectPages(func(cursor string) ([]SecretStore, *string, error) {
		var page SecretStoreListResponse
		if err := c.doRequest(ctx, http.MethodGet, withCursor(base, cursor), nil, &page); err != nil {
			return nil, nil, err
		}
		return page.Items, page.NextPage, nil
	})
}
