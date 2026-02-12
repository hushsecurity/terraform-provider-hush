package client

import (
	"context"
	"fmt"
	"net/http"
)

const (
	AccessCredentialTypePostgres AccessCredentialType = "postgres"
	AccessCredentialTypeMongoDB  AccessCredentialType = "mongodb"
	AccessCredentialTypeMySQL    AccessCredentialType = "mysql"
	AccessCredentialTypeMariaDB  AccessCredentialType = "mariadb"
	AccessCredentialTypeOpenAI   AccessCredentialType = "openai"
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

// MongoDB

type MongoDBAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	DBName        string               `json:"db_name,omitempty"`
	Host          string               `json:"host,omitempty"`
	Port          int                  `json:"port,omitempty"`
	Username      string               `json:"username,omitempty"`
	AuthSource    string               `json:"auth_source,omitempty"`
	TLS           bool                 `json:"tls,omitempty"`
	TLSCA         string               `json:"tls_ca,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateMongoDBAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	DBName        string   `json:"db_name"`
	Host          string   `json:"host"`
	Port          int      `json:"port,omitempty"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	AuthSource    string   `json:"auth_source,omitempty"`
	TLS           bool     `json:"tls,omitempty"`
	TLSCA         string   `json:"tls_ca,omitempty"`
}

type UpdateMongoDBAccessCredentialInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
	DBName        *string   `json:"db_name,omitempty"`
	Host          *string   `json:"host,omitempty"`
	Port          *int      `json:"port,omitempty"`
	Username      *string   `json:"username,omitempty"`
	Password      *string   `json:"password,omitempty"`
	AuthSource    *string   `json:"auth_source,omitempty"`
	TLS           *bool     `json:"tls,omitempty"`
	TLSCA         *string   `json:"tls_ca,omitempty"`
}

func CreateMongoDBAccessCredential(ctx context.Context, c *Client, input *CreateMongoDBAccessCredentialInput) (*MongoDBAccessCredential, error) {
	path := accessCredentialsEndpoint + "/mongodb"
	var resp MongoDBAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetMongoDBAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetMongoDBAccessCredential(ctx context.Context, c *Client, id string) (*MongoDBAccessCredential, error) {
	path := fmt.Sprintf("%s/mongodb/%s", accessCredentialsEndpoint, id)
	var resp MongoDBAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateMongoDBAccessCredential(ctx context.Context, c *Client, id string, input *UpdateMongoDBAccessCredentialInput) (*MongoDBAccessCredential, error) {
	path := fmt.Sprintf("%s/mongodb/%s", accessCredentialsEndpoint, id)
	var resp MongoDBAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetMongoDBAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m MongoDBAccessCredential) statusFields() (string, string) {
	return m.Status, m.StatusDetail
}

// MySQL

type MySQLAccessCredential struct {
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

type CreateMySQLAccessCredentialInput struct {
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

type UpdateMySQLAccessCredentialInput struct {
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

func CreateMySQLAccessCredential(ctx context.Context, c *Client, input *CreateMySQLAccessCredentialInput) (*MySQLAccessCredential, error) {
	path := accessCredentialsEndpoint + "/mysql"
	var resp MySQLAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetMySQLAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetMySQLAccessCredential(ctx context.Context, c *Client, id string) (*MySQLAccessCredential, error) {
	path := fmt.Sprintf("%s/mysql/%s", accessCredentialsEndpoint, id)
	var resp MySQLAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateMySQLAccessCredential(ctx context.Context, c *Client, id string, input *UpdateMySQLAccessCredentialInput) (*MySQLAccessCredential, error) {
	path := fmt.Sprintf("%s/mysql/%s", accessCredentialsEndpoint, id)
	var resp MySQLAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetMySQLAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m MySQLAccessCredential) statusFields() (string, string) {
	return m.Status, m.StatusDetail
}

// MariaDB

type MariaDBAccessCredential struct {
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

type CreateMariaDBAccessCredentialInput struct {
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

type UpdateMariaDBAccessCredentialInput struct {
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

func CreateMariaDBAccessCredential(ctx context.Context, c *Client, input *CreateMariaDBAccessCredentialInput) (*MariaDBAccessCredential, error) {
	path := accessCredentialsEndpoint + "/mariadb"
	var resp MariaDBAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetMariaDBAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetMariaDBAccessCredential(ctx context.Context, c *Client, id string) (*MariaDBAccessCredential, error) {
	path := fmt.Sprintf("%s/mariadb/%s", accessCredentialsEndpoint, id)
	var resp MariaDBAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateMariaDBAccessCredential(ctx context.Context, c *Client, id string, input *UpdateMariaDBAccessCredentialInput) (*MariaDBAccessCredential, error) {
	path := fmt.Sprintf("%s/mariadb/%s", accessCredentialsEndpoint, id)
	var resp MariaDBAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetMariaDBAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m MariaDBAccessCredential) statusFields() (string, string) {
	return m.Status, m.StatusDetail
}

// OpenAI

type OpenAIAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	ProjectID     string               `json:"project_id,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateOpenAIAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	APIKey        string   `json:"api_key"`
	ProjectID     string   `json:"project_id,omitempty"`
}

type UpdateOpenAIAccessCredentialInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
	APIKey        *string   `json:"api_key,omitempty"`
	ProjectID     *string   `json:"project_id,omitempty"`
}

func CreateOpenAIAccessCredential(ctx context.Context, c *Client, input *CreateOpenAIAccessCredentialInput) (*OpenAIAccessCredential, error) {
	path := accessCredentialsEndpoint + "/openai"
	var resp OpenAIAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetOpenAIAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetOpenAIAccessCredential(ctx context.Context, c *Client, id string) (*OpenAIAccessCredential, error) {
	path := fmt.Sprintf("%s/openai/%s", accessCredentialsEndpoint, id)
	var resp OpenAIAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateOpenAIAccessCredential(ctx context.Context, c *Client, id string, input *UpdateOpenAIAccessCredentialInput) (*OpenAIAccessCredential, error) {
	path := fmt.Sprintf("%s/openai/%s", accessCredentialsEndpoint, id)
	var resp OpenAIAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetOpenAIAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (o OpenAIAccessCredential) statusFields() (string, string) {
	return o.Status, o.StatusDetail
}
