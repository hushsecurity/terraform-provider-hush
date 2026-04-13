package client

import (
	"context"
	"fmt"
	"net/http"
)

const (
	AccessCredentialTypePostgres      AccessCredentialType = "postgres"
	AccessCredentialTypeMongoDB       AccessCredentialType = "mongodb"
	AccessCredentialTypeMySQL         AccessCredentialType = "mysql"
	AccessCredentialTypeMariaDB       AccessCredentialType = "mariadb"
	AccessCredentialTypeOpenAI        AccessCredentialType = "openai"
	AccessCredentialTypeGemini        AccessCredentialType = "gemini"
	AccessCredentialTypeGrok          AccessCredentialType = "grok"
	AccessCredentialTypeRedis         AccessCredentialType = "redis"
	AccessCredentialTypeBedrock       AccessCredentialType = "bedrock"
	AccessCredentialTypeApigee        AccessCredentialType = "apigee"
	AccessCredentialTypeElasticsearch AccessCredentialType = "elasticsearch"
	AccessCredentialTypeRabbitmq      AccessCredentialType = "rabbitmq"
	AccessCredentialTypeGCPSA         AccessCredentialType = "gcp_service_account"
	AccessCredentialTypeAzureApp      AccessCredentialType = "azure_app"
	AccessCredentialTypeAWSAccessKey  AccessCredentialType = "aws_access_key"
	AccessCredentialTypeTwilio        AccessCredentialType = "twilio"
	AccessCredentialTypeSnowflake     AccessCredentialType = "snowflake"
	AccessCredentialTypeAWSWIF        AccessCredentialType = "aws_wif"
	AccessCredentialTypeGCPWIF        AccessCredentialType = "gcp_wif"
	AccessCredentialTypeGitlab        AccessCredentialType = "gitlab"
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
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	DBName      *string `json:"db_name,omitempty"`
	Host        *string `json:"host,omitempty"`
	Port        *int    `json:"port,omitempty"`
	SSLMode     *string `json:"ssl_mode,omitempty"`
	SSLCA       *string `json:"ssl_ca,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
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
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	DBName      *string `json:"db_name,omitempty"`
	Host        *string `json:"host,omitempty"`
	Port        *int    `json:"port,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	AuthSource  *string `json:"auth_source,omitempty"`
	TLS         *bool   `json:"tls,omitempty"`
	TLSCA       *string `json:"tls_ca,omitempty"`
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
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	DBName      *string `json:"db_name,omitempty"`
	Host        *string `json:"host,omitempty"`
	Port        *int    `json:"port,omitempty"`
	SSLMode     *string `json:"ssl_mode,omitempty"`
	SSLCA       *string `json:"ssl_ca,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
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
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	DBName      *string `json:"db_name,omitempty"`
	Host        *string `json:"host,omitempty"`
	Port        *int    `json:"port,omitempty"`
	SSLMode     *string `json:"ssl_mode,omitempty"`
	SSLCA       *string `json:"ssl_ca,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
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
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	APIKey      *string `json:"api_key,omitempty"`
	ProjectID   *string `json:"project_id,omitempty"`
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

// Gemini

type GeminiAccessCredential struct {
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

type CreateGeminiAccessCredentialInput struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	DeploymentIDs     []string `json:"deployment_ids"`
	ServiceAccountKey string   `json:"service_account_key,omitempty"`
	ProjectID         string   `json:"project_id"`
}

type UpdateGeminiAccessCredentialInput struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	ServiceAccountKey *string `json:"service_account_key,omitempty"`
	ProjectID         *string `json:"project_id,omitempty"`
}

func CreateGeminiAccessCredential(ctx context.Context, c *Client, input *CreateGeminiAccessCredentialInput) (*GeminiAccessCredential, error) {
	path := accessCredentialsEndpoint + "/gemini"
	var resp GeminiAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetGeminiAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGeminiAccessCredential(ctx context.Context, c *Client, id string) (*GeminiAccessCredential, error) {
	path := fmt.Sprintf("%s/gemini/%s", accessCredentialsEndpoint, id)
	var resp GeminiAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGeminiAccessCredential(ctx context.Context, c *Client, id string, input *UpdateGeminiAccessCredentialInput) (*GeminiAccessCredential, error) {
	path := fmt.Sprintf("%s/gemini/%s", accessCredentialsEndpoint, id)
	var resp GeminiAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetGeminiAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g GeminiAccessCredential) statusFields() (string, string) {
	return g.Status, g.StatusDetail
}

// Grok

type GrokAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	TeamID        string               `json:"team_id,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateGrokAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	APIKey        string   `json:"api_key"`
	TeamID        string   `json:"team_id"`
}

type UpdateGrokAccessCredentialInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
	APIKey        *string   `json:"api_key,omitempty"`
	TeamID        *string   `json:"team_id,omitempty"`
}

func CreateGrokAccessCredential(ctx context.Context, c *Client, input *CreateGrokAccessCredentialInput) (*GrokAccessCredential, error) {
	path := accessCredentialsEndpoint + "/grok"
	var resp GrokAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetGrokAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGrokAccessCredential(ctx context.Context, c *Client, id string) (*GrokAccessCredential, error) {
	path := fmt.Sprintf("%s/grok/%s", accessCredentialsEndpoint, id)
	var resp GrokAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGrokAccessCredential(ctx context.Context, c *Client, id string, input *UpdateGrokAccessCredentialInput) (*GrokAccessCredential, error) {
	path := fmt.Sprintf("%s/grok/%s", accessCredentialsEndpoint, id)
	var resp GrokAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetGrokAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g GrokAccessCredential) statusFields() (string, string) {
	return g.Status, g.StatusDetail
}

// Redis

type RedisAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	Host          string               `json:"host,omitempty"`
	Port          int                  `json:"port,omitempty"`
	Username      string               `json:"username,omitempty"`
	Database      int                  `json:"database"`
	TLS           bool                 `json:"tls,omitempty"`
	TLSCA         string               `json:"tls_ca,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateRedisAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Host          string   `json:"host"`
	Port          int      `json:"port,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password"`
	Database      *int     `json:"database,omitempty"`
	TLS           bool     `json:"tls,omitempty"`
	TLSCA         string   `json:"tls_ca,omitempty"`
}

type UpdateRedisAccessCredentialInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
	Host          *string   `json:"host,omitempty"`
	Port          *int      `json:"port,omitempty"`
	Username      *string   `json:"username,omitempty"`
	Password      *string   `json:"password,omitempty"`
	Database      *int      `json:"database,omitempty"`
	TLS           *bool     `json:"tls,omitempty"`
	TLSCA         *string   `json:"tls_ca,omitempty"`
}

func CreateRedisAccessCredential(ctx context.Context, c *Client, input *CreateRedisAccessCredentialInput) (*RedisAccessCredential, error) {
	path := accessCredentialsEndpoint + "/redis"
	var resp RedisAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetRedisAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetRedisAccessCredential(ctx context.Context, c *Client, id string) (*RedisAccessCredential, error) {
	path := fmt.Sprintf("%s/redis/%s", accessCredentialsEndpoint, id)
	var resp RedisAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateRedisAccessCredential(ctx context.Context, c *Client, id string, input *UpdateRedisAccessCredentialInput) (*RedisAccessCredential, error) {
	path := fmt.Sprintf("%s/redis/%s", accessCredentialsEndpoint, id)
	var resp RedisAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetRedisAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r RedisAccessCredential) statusFields() (string, string) {
	return r.Status, r.StatusDetail
}

// Bedrock

type BedrockAccessCredential struct {
	ID                     string               `json:"id,omitempty"`
	Name                   string               `json:"name"`
	Description            string               `json:"description,omitempty"`
	Type                   AccessCredentialType `json:"type"`
	Kind                   string               `json:"kind,omitempty"`
	DeploymentIDs          []string             `json:"deployment_ids"`
	Region                 string               `json:"region"`
	AccessKeyID            *string              `json:"access_key_id,omitempty"`
	HasProviderCredentials bool                 `json:"has_provider_credentials"`
	Status                 string               `json:"status,omitempty"`
	StatusDetail           string               `json:"status_detail,omitempty"`
}

type CreateBedrockAccessCredentialInput struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	DeploymentIDs   []string `json:"deployment_ids"`
	Region          string   `json:"region"`
	AccessKeyID     *string  `json:"access_key_id,omitempty"`
	SecretAccessKey *string  `json:"secret_access_key,omitempty"`
}

type UpdateBedrockAccessCredentialInput struct {
	Name            *string   `json:"name,omitempty"`
	Description     *string   `json:"description,omitempty"`
	DeploymentIDs   *[]string `json:"deployment_ids,omitempty"`
	Region          *string   `json:"region,omitempty"`
	AccessKeyID     *string   `json:"access_key_id,omitempty"`
	SecretAccessKey *string   `json:"secret_access_key,omitempty"`
}

func CreateBedrockAccessCredential(ctx context.Context, c *Client, input *CreateBedrockAccessCredentialInput) (*BedrockAccessCredential, error) {
	path := accessCredentialsEndpoint + "/bedrock"
	var resp BedrockAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetBedrockAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetBedrockAccessCredential(ctx context.Context, c *Client, id string) (*BedrockAccessCredential, error) {
	path := fmt.Sprintf("%s/bedrock/%s", accessCredentialsEndpoint, id)
	var resp BedrockAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateBedrockAccessCredential(ctx context.Context, c *Client, id string, input *UpdateBedrockAccessCredentialInput) (*BedrockAccessCredential, error) {
	path := fmt.Sprintf("%s/bedrock/%s", accessCredentialsEndpoint, id)
	var resp BedrockAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetBedrockAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (b BedrockAccessCredential) statusFields() (string, string) {
	return b.Status, b.StatusDetail
}

// Apigee

type ApigeeAccessCredential struct {
	ID                     string               `json:"id,omitempty"`
	Name                   string               `json:"name"`
	Description            string               `json:"description,omitempty"`
	Type                   AccessCredentialType `json:"type"`
	Kind                   string               `json:"kind,omitempty"`
	DeploymentIDs          []string             `json:"deployment_ids"`
	HasProviderCredentials bool                 `json:"has_provider_credentials"`
	Status                 string               `json:"status,omitempty"`
	StatusDetail           string               `json:"status_detail,omitempty"`
}

type CreateApigeeAccessCredentialInput struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	DeploymentIDs     []string `json:"deployment_ids"`
	ServiceAccountKey *string  `json:"service_account_key,omitempty"`
}

type UpdateApigeeAccessCredentialInput struct {
	Name              *string   `json:"name,omitempty"`
	Description       *string   `json:"description,omitempty"`
	DeploymentIDs     *[]string `json:"deployment_ids,omitempty"`
	ServiceAccountKey *string   `json:"service_account_key,omitempty"`
}

func CreateApigeeAccessCredential(ctx context.Context, c *Client, input *CreateApigeeAccessCredentialInput) (*ApigeeAccessCredential, error) {
	path := accessCredentialsEndpoint + "/apigee"
	var resp ApigeeAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetApigeeAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetApigeeAccessCredential(ctx context.Context, c *Client, id string) (*ApigeeAccessCredential, error) {
	path := fmt.Sprintf("%s/apigee/%s", accessCredentialsEndpoint, id)
	var resp ApigeeAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateApigeeAccessCredential(ctx context.Context, c *Client, id string, input *UpdateApigeeAccessCredentialInput) (*ApigeeAccessCredential, error) {
	path := fmt.Sprintf("%s/apigee/%s", accessCredentialsEndpoint, id)
	var resp ApigeeAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetApigeeAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a ApigeeAccessCredential) statusFields() (string, string) {
	return a.Status, a.StatusDetail
}

// Elasticsearch

type ElasticsearchAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	Host          string               `json:"host,omitempty"`
	Port          int                  `json:"port,omitempty"`
	Username      string               `json:"username,omitempty"`
	TLS           bool                 `json:"tls,omitempty"`
	TLSCA         string               `json:"tls_ca,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateElasticsearchAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Host          string   `json:"host"`
	Port          int      `json:"port,omitempty"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	TLS           bool     `json:"tls,omitempty"`
	TLSCA         string   `json:"tls_ca,omitempty"`
}

type UpdateElasticsearchAccessCredentialInput struct {
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	DeploymentIDs *[]string `json:"deployment_ids,omitempty"`
	Host          *string   `json:"host,omitempty"`
	Port          *int      `json:"port,omitempty"`
	Username      *string   `json:"username,omitempty"`
	Password      *string   `json:"password,omitempty"`
	TLS           *bool     `json:"tls,omitempty"`
	TLSCA         *string   `json:"tls_ca,omitempty"`
}

func CreateElasticsearchAccessCredential(ctx context.Context, c *Client, input *CreateElasticsearchAccessCredentialInput) (*ElasticsearchAccessCredential, error) {
	path := accessCredentialsEndpoint + "/elasticsearch"
	var resp ElasticsearchAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetElasticsearchAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetElasticsearchAccessCredential(ctx context.Context, c *Client, id string) (*ElasticsearchAccessCredential, error) {
	path := fmt.Sprintf("%s/elasticsearch/%s", accessCredentialsEndpoint, id)
	var resp ElasticsearchAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateElasticsearchAccessCredential(ctx context.Context, c *Client, id string, input *UpdateElasticsearchAccessCredentialInput) (*ElasticsearchAccessCredential, error) {
	path := fmt.Sprintf("%s/elasticsearch/%s", accessCredentialsEndpoint, id)
	var resp ElasticsearchAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetElasticsearchAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (e ElasticsearchAccessCredential) statusFields() (string, string) {
	return e.Status, e.StatusDetail
}

// RabbitMQ

type RabbitmqAccessCredential struct {
	ID             string               `json:"id,omitempty"`
	Name           string               `json:"name"`
	Description    string               `json:"description,omitempty"`
	Type           AccessCredentialType `json:"type"`
	Kind           string               `json:"kind,omitempty"`
	DeploymentIDs  []string             `json:"deployment_ids"`
	Host           string               `json:"host,omitempty"`
	Port           int                  `json:"port,omitempty"`
	ManagementPort int                  `json:"management_port,omitempty"`
	Username       string               `json:"username,omitempty"`
	Vhost          string               `json:"vhost,omitempty"`
	TLS            bool                 `json:"tls,omitempty"`
	TLSCA          string               `json:"tls_ca,omitempty"`
	Status         string               `json:"status,omitempty"`
	StatusDetail   string               `json:"status_detail,omitempty"`
}

type CreateRabbitmqAccessCredentialInput struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	DeploymentIDs  []string `json:"deployment_ids"`
	Host           string   `json:"host"`
	Port           int      `json:"port,omitempty"`
	ManagementPort int      `json:"management_port,omitempty"`
	Username       string   `json:"username,omitempty"`
	Password       string   `json:"password"`
	Vhost          string   `json:"vhost,omitempty"`
	TLS            bool     `json:"tls,omitempty"`
	TLSCA          string   `json:"tls_ca,omitempty"`
}

type UpdateRabbitmqAccessCredentialInput struct {
	Name           *string   `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	DeploymentIDs  *[]string `json:"deployment_ids,omitempty"`
	Host           *string   `json:"host,omitempty"`
	Port           *int      `json:"port,omitempty"`
	ManagementPort *int      `json:"management_port,omitempty"`
	Username       *string   `json:"username,omitempty"`
	Password       *string   `json:"password,omitempty"`
	Vhost          *string   `json:"vhost,omitempty"`
	TLS            *bool     `json:"tls,omitempty"`
	TLSCA          *string   `json:"tls_ca,omitempty"`
}

func CreateRabbitmqAccessCredential(ctx context.Context, c *Client, input *CreateRabbitmqAccessCredentialInput) (*RabbitmqAccessCredential, error) {
	path := accessCredentialsEndpoint + "/rabbitmq"
	var resp RabbitmqAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetRabbitmqAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetRabbitmqAccessCredential(ctx context.Context, c *Client, id string) (*RabbitmqAccessCredential, error) {
	path := fmt.Sprintf("%s/rabbitmq/%s", accessCredentialsEndpoint, id)
	var resp RabbitmqAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateRabbitmqAccessCredential(ctx context.Context, c *Client, id string, input *UpdateRabbitmqAccessCredentialInput) (*RabbitmqAccessCredential, error) {
	path := fmt.Sprintf("%s/rabbitmq/%s", accessCredentialsEndpoint, id)
	var resp RabbitmqAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetRabbitmqAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r RabbitmqAccessCredential) statusFields() (string, string) {
	return r.Status, r.StatusDetail
}

// GCP SA

type GCPSAAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateGCPSAAccessCredentialInput struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	DeploymentIDs     []string `json:"deployment_ids"`
	ServiceAccountKey string   `json:"service_account_key,omitempty"`
}

type UpdateGCPSAAccessCredentialInput struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	ServiceAccountKey *string `json:"service_account_key,omitempty"`
}

func CreateGCPSAAccessCredential(ctx context.Context, c *Client, input *CreateGCPSAAccessCredentialInput) (*GCPSAAccessCredential, error) {
	path := accessCredentialsEndpoint + "/gcp_sa"
	var resp GCPSAAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetGCPSAAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGCPSAAccessCredential(ctx context.Context, c *Client, id string) (*GCPSAAccessCredential, error) {
	path := fmt.Sprintf("%s/gcp_sa/%s", accessCredentialsEndpoint, id)
	var resp GCPSAAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGCPSAAccessCredential(ctx context.Context, c *Client, id string, input *UpdateGCPSAAccessCredentialInput) (*GCPSAAccessCredential, error) {
	path := fmt.Sprintf("%s/gcp_sa/%s", accessCredentialsEndpoint, id)
	var resp GCPSAAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetGCPSAAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g GCPSAAccessCredential) statusFields() (string, string) {
	return g.Status, g.StatusDetail
}

// Azure App

type AzureAppAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	TenantID      string               `json:"tenant_id,omitempty"`
	ClientID      string               `json:"client_id,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateAzureAppAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	TenantID      string   `json:"tenant_id"`
	ClientID      string   `json:"client_id"`
	ClientSecret  string   `json:"client_secret"`
}

type UpdateAzureAppAccessCredentialInput struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	TenantID     *string `json:"tenant_id,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty"`
}

func CreateAzureAppAccessCredential(ctx context.Context, c *Client, input *CreateAzureAppAccessCredentialInput) (*AzureAppAccessCredential, error) {
	path := accessCredentialsEndpoint + "/azure_app"
	var resp AzureAppAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetAzureAppAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAzureAppAccessCredential(ctx context.Context, c *Client, id string) (*AzureAppAccessCredential, error) {
	path := fmt.Sprintf("%s/azure_app/%s", accessCredentialsEndpoint, id)
	var resp AzureAppAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateAzureAppAccessCredential(ctx context.Context, c *Client, id string, input *UpdateAzureAppAccessCredentialInput) (*AzureAppAccessCredential, error) {
	path := fmt.Sprintf("%s/azure_app/%s", accessCredentialsEndpoint, id)
	var resp AzureAppAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetAzureAppAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a AzureAppAccessCredential) statusFields() (string, string) {
	return a.Status, a.StatusDetail
}

// AWS Access Key

type AWSAccessKeyAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	AccessKeyID   string               `json:"access_key_id,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateAWSAccessKeyAccessCredentialInput struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	DeploymentIDs   []string `json:"deployment_ids"`
	AccessKeyID     string   `json:"access_key_id"`
	SecretAccessKey string   `json:"secret_access_key"`
}

type UpdateAWSAccessKeyAccessCredentialInput struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	AccessKeyID     *string `json:"access_key_id,omitempty"`
	SecretAccessKey *string `json:"secret_access_key,omitempty"`
}

func CreateAWSAccessKeyAccessCredential(ctx context.Context, c *Client, input *CreateAWSAccessKeyAccessCredentialInput) (*AWSAccessKeyAccessCredential, error) {
	path := accessCredentialsEndpoint + "/aws_access_key"
	var resp AWSAccessKeyAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetAWSAccessKeyAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAWSAccessKeyAccessCredential(ctx context.Context, c *Client, id string) (*AWSAccessKeyAccessCredential, error) {
	path := fmt.Sprintf("%s/aws_access_key/%s", accessCredentialsEndpoint, id)
	var resp AWSAccessKeyAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateAWSAccessKeyAccessCredential(ctx context.Context, c *Client, id string, input *UpdateAWSAccessKeyAccessCredentialInput) (*AWSAccessKeyAccessCredential, error) {
	path := fmt.Sprintf("%s/aws_access_key/%s", accessCredentialsEndpoint, id)
	var resp AWSAccessKeyAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetAWSAccessKeyAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a AWSAccessKeyAccessCredential) statusFields() (string, string) {
	return a.Status, a.StatusDetail
}

// Twilio

type TwilioAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	AccountSID    string               `json:"account_sid"`
	APIKeySID     string               `json:"api_key_sid"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateTwilioAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	AccountSID    string   `json:"account_sid"`
	APIKeySID     string   `json:"api_key_sid"`
	APIKeySecret  string   `json:"api_key_secret"`
}

type UpdateTwilioAccessCredentialInput struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	AccountSID   *string `json:"account_sid,omitempty"`
	APIKeySID    *string `json:"api_key_sid,omitempty"`
	APIKeySecret *string `json:"api_key_secret,omitempty"`
}

func CreateTwilioAccessCredential(ctx context.Context, c *Client, input *CreateTwilioAccessCredentialInput) (*TwilioAccessCredential, error) {
	path := accessCredentialsEndpoint + "/twilio"
	var resp TwilioAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetTwilioAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetTwilioAccessCredential(ctx context.Context, c *Client, id string) (*TwilioAccessCredential, error) {
	path := fmt.Sprintf("%s/twilio/%s", accessCredentialsEndpoint, id)
	var resp TwilioAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateTwilioAccessCredential(ctx context.Context, c *Client, id string, input *UpdateTwilioAccessCredentialInput) (*TwilioAccessCredential, error) {
	path := fmt.Sprintf("%s/twilio/%s", accessCredentialsEndpoint, id)
	var resp TwilioAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetTwilioAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (t TwilioAccessCredential) statusFields() (string, string) {
	return t.Status, t.StatusDetail
}

// Snowflake

type SnowflakeAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	Account       string               `json:"account,omitempty"`
	Warehouse     string               `json:"warehouse,omitempty"`
	Database      string               `json:"database,omitempty"`
	Schema        string               `json:"schema,omitempty"`
	Role          string               `json:"role,omitempty"`
	Username      string               `json:"username,omitempty"`
	AuthMethod    string               `json:"auth_method,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateSnowflakeAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Account       string   `json:"account"`
	Warehouse     string   `json:"warehouse"`
	Database      string   `json:"database"`
	Schema        string   `json:"schema,omitempty"`
	Role          string   `json:"role,omitempty"`
	Username      string   `json:"username"`
	Password      string   `json:"password,omitempty"`
	PrivateKey    string   `json:"private_key,omitempty"`
	AuthMethod    string   `json:"auth_method"`
}

type UpdateSnowflakeAccessCredentialInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Account     *string `json:"account,omitempty"`
	Warehouse   *string `json:"warehouse,omitempty"`
	Database    *string `json:"database,omitempty"`
	Schema      *string `json:"schema,omitempty"`
	Role        *string `json:"role,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	PrivateKey  *string `json:"private_key,omitempty"`
	AuthMethod  *string `json:"auth_method,omitempty"`
}

func CreateSnowflakeAccessCredential(ctx context.Context, c *Client, input *CreateSnowflakeAccessCredentialInput) (*SnowflakeAccessCredential, error) {
	path := accessCredentialsEndpoint + "/snowflake"
	var resp SnowflakeAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetSnowflakeAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetSnowflakeAccessCredential(ctx context.Context, c *Client, id string) (*SnowflakeAccessCredential, error) {
	path := fmt.Sprintf("%s/snowflake/%s", accessCredentialsEndpoint, id)
	var resp SnowflakeAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateSnowflakeAccessCredential(ctx context.Context, c *Client, id string, input *UpdateSnowflakeAccessCredentialInput) (*SnowflakeAccessCredential, error) {
	path := fmt.Sprintf("%s/snowflake/%s", accessCredentialsEndpoint, id)
	var resp SnowflakeAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetSnowflakeAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s SnowflakeAccessCredential) statusFields() (string, string) {
	return s.Status, s.StatusDetail
}

// AWS WIF

type AwsWifAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	Audience      string               `json:"audience,omitempty"`
	IssuerURL     string               `json:"issuer_url,omitempty"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateAwsWifAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
}

type UpdateAwsWifAccessCredentialInput struct {
	Name          *string  `json:"name,omitempty"`
	Description   *string  `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids,omitempty"`
}

func CreateAwsWifAccessCredential(ctx context.Context, c *Client, input *CreateAwsWifAccessCredentialInput) (*AwsWifAccessCredential, error) {
	path := accessCredentialsEndpoint + "/aws_wif"
	var resp AwsWifAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetAwsWifAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAwsWifAccessCredential(ctx context.Context, c *Client, id string) (*AwsWifAccessCredential, error) {
	path := fmt.Sprintf("%s/aws_wif/%s", accessCredentialsEndpoint, id)
	var resp AwsWifAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateAwsWifAccessCredential(ctx context.Context, c *Client, id string, input *UpdateAwsWifAccessCredentialInput) (*AwsWifAccessCredential, error) {
	path := fmt.Sprintf("%s/aws_wif/%s", accessCredentialsEndpoint, id)
	var resp AwsWifAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetAwsWifAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a AwsWifAccessCredential) statusFields() (string, string) {
	return a.Status, a.StatusDetail
}

// Gitlab

type GitlabAccessCredential struct {
	ID            string               `json:"id,omitempty"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          AccessCredentialType `json:"type"`
	Kind          string               `json:"kind,omitempty"`
	DeploymentIDs []string             `json:"deployment_ids"`
	BaseURL       string               `json:"base_url"`
	ResourceType  string               `json:"resource_type"`
	ResourceID    string               `json:"resource_id"`
	Status        string               `json:"status,omitempty"`
	StatusDetail  string               `json:"status_detail,omitempty"`
}

type CreateGitlabAccessCredentialInput struct {
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Token         string   `json:"token"`
	BaseURL       string   `json:"base_url"`
	ResourceType  string   `json:"resource_type"`
	ResourceID    string   `json:"resource_id"`
}

type UpdateGitlabAccessCredentialInput struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Token        *string `json:"token,omitempty"`
	BaseURL      *string `json:"base_url,omitempty"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
}

func CreateGitlabAccessCredential(ctx context.Context, c *Client, input *CreateGitlabAccessCredentialInput) (*GitlabAccessCredential, error) {
	path := accessCredentialsEndpoint + "/gitlab"
	var resp GitlabAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetGitlabAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGitlabAccessCredential(ctx context.Context, c *Client, id string) (*GitlabAccessCredential, error) {
	path := fmt.Sprintf("%s/gitlab/%s", accessCredentialsEndpoint, id)
	var resp GitlabAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGitlabAccessCredential(ctx context.Context, c *Client, id string, input *UpdateGitlabAccessCredentialInput) (*GitlabAccessCredential, error) {
	path := fmt.Sprintf("%s/gitlab/%s", accessCredentialsEndpoint, id)
	var resp GitlabAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetGitlabAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g GitlabAccessCredential) statusFields() (string, string) {
	return g.Status, g.StatusDetail
}

// GCP WIF

type GcpWifAccessCredential struct {
	ID                 string               `json:"id,omitempty"`
	Name               string               `json:"name"`
	Description        string               `json:"description,omitempty"`
	Type               AccessCredentialType `json:"type"`
	Kind               string               `json:"kind,omitempty"`
	DeploymentIDs      []string             `json:"deployment_ids"`
	ProjectNumber      string               `json:"project_number,omitempty"`
	PoolID             string               `json:"pool_id,omitempty"`
	WorkloadProviderID string               `json:"workload_provider_id,omitempty"`
	Audience           string               `json:"audience,omitempty"`
	IssuerURL          string               `json:"issuer_url,omitempty"`
	Status             string               `json:"status,omitempty"`
	StatusDetail       string               `json:"status_detail,omitempty"`
}

type CreateGcpWifAccessCredentialInput struct {
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	DeploymentIDs      []string `json:"deployment_ids"`
	ProjectNumber      string   `json:"project_number"`
	PoolID             string   `json:"pool_id"`
	WorkloadProviderID string   `json:"workload_provider_id"`
	Audience           string   `json:"audience,omitempty"`
}

type UpdateGcpWifAccessCredentialInput struct {
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	DeploymentIDs      []string `json:"deployment_ids,omitempty"`
	ProjectNumber      *string  `json:"project_number,omitempty"`
	PoolID             *string  `json:"pool_id,omitempty"`
	WorkloadProviderID *string  `json:"workload_provider_id,omitempty"`
	Audience           *string  `json:"audience,omitempty"`
}

func CreateGcpWifAccessCredential(ctx context.Context, c *Client, input *CreateGcpWifAccessCredentialInput) (*GcpWifAccessCredential, error) {
	path := accessCredentialsEndpoint + "/gcp_wif"
	var resp GcpWifAccessCredential
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, resp.ID, GetGcpWifAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGcpWifAccessCredential(ctx context.Context, c *Client, id string) (*GcpWifAccessCredential, error) {
	path := fmt.Sprintf("%s/gcp_wif/%s", accessCredentialsEndpoint, id)
	var resp GcpWifAccessCredential
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGcpWifAccessCredential(ctx context.Context, c *Client, id string, input *UpdateGcpWifAccessCredentialInput) (*GcpWifAccessCredential, error) {
	path := fmt.Sprintf("%s/gcp_wif/%s", accessCredentialsEndpoint, id)
	var resp GcpWifAccessCredential
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	if err := waitForResourceStatus(ctx, c, id, GetGcpWifAccessCredential); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (g GcpWifAccessCredential) statusFields() (string, string) {
	return g.Status, g.StatusDetail
}
