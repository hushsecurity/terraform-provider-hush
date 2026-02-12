package client

import (
	"context"
	"fmt"
	"net/http"
)

const (
	AccessCredentialTypePostgres AccessCredentialType = "postgres"
)

// Postgres

type PostgresAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	DBName        string               `json:"db_name,omitempty"`
	Host          string               `json:"host,omitempty"`
	Port          int                  `json:"port,omitempty"`
	SSLMode       string               `json:"ssl_mode,omitempty"`
	SSLCA         string               `json:"ssl_ca,omitempty"`
	Username      string               `json:"username,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreatePostgresAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	DBName        string   `json:"db_name"`
	Host          string   `json:"host"`
	Port          int      `json:"port,omitempty"`
	SSLMode       string   `json:"ssl_mode,omitempty"`
	SSLCA         string   `json:"ssl_ca,omitempty"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
}

type UpdatePostgresAccessCredentialInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
	DBName        *string   `json:"db_name,omitempty"`
	Host          *string   `json:"host,omitempty"`
	Port          *int      `json:"port,omitempty"`
	SSLMode       *string   `json:"ssl_mode,omitempty"`
	SSLCA         *string   `json:"ssl_ca,omitempty"`
	Username      *string   `json:"username,omitempty"`
	Password      *string   `json:"password,omitempty"`
}

func CreatePostgresAccessCredential(ctx context.Context, c *Client, input *CreatePostgresAccessCredentialInput) (*PostgresAccessCredential, error) {
	path := accessCredentialsEndpoint + "/postgres"
	var resp PostgresAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetPostgresAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetPostgresAccessCredential(ctx context.Context, c *Client, id string) (*PostgresAccessCredential, error) {
	path := fmt.Sprintf("%s/postgres/%s", accessCredentialsEndpoint, id)
	var resp PostgresAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdatePostgresAccessCredential(ctx context.Context, c *Client, id string, input *UpdatePostgresAccessCredentialInput) (*PostgresAccessCredential, error) {
	path := fmt.Sprintf("%s/postgres/%s", accessCredentialsEndpoint, id)
	var resp PostgresAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetPostgresAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p PostgresAccessCredential) statusFields() (string, string) {
	return p.Status, p.StatusDetail
}
