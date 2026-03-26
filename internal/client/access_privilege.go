package client

import (
	"context"
	"fmt"
	"net/http"
)

const accessPrivilegesEndpoint = "/v1/access_privileges"

// Postgres

type PostgresGrant struct {
	Privileges  []string `json:"privileges"`
	ObjectType  string   `json:"object_type"`
	ObjectNames []string `json:"object_names,omitempty"`
	ColumnNames []string `json:"column_names,omitempty"`
	AllInSchema bool     `json:"all_in_schema,omitempty"`
}

type PostgresAccessPrivilege struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type,omitempty"`
	Grants      []PostgresGrant `json:"grants"`
}

type CreatePostgresAccessPrivilegeInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Grants      []PostgresGrant `json:"grants"`
}

type UpdatePostgresAccessPrivilegeInput struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Grants      *[]PostgresGrant `json:"grants,omitempty"`
}

func CreatePostgresAccessPrivilege(ctx context.Context, c *Client, input *CreatePostgresAccessPrivilegeInput) (*PostgresAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/postgres"
	var resp PostgresAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetPostgresAccessPrivilege(ctx context.Context, c *Client, id string) (*PostgresAccessPrivilege, error) {
	path := fmt.Sprintf("%s/postgres/%s", accessPrivilegesEndpoint, id)
	var resp PostgresAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdatePostgresAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdatePostgresAccessPrivilegeInput) (*PostgresAccessPrivilege, error) {
	path := fmt.Sprintf("%s/postgres/%s", accessPrivilegesEndpoint, id)
	var resp PostgresAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MongoDB

type MongoDBGrant struct {
	Privileges    []string `json:"privileges"`
	ResourceType  string   `json:"resource_type"`
	ResourceNames []string `json:"resource_names,omitempty"`
}

type MongoDBAccessPrivilege struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type,omitempty"`
	Grants      []MongoDBGrant `json:"grants"`
}

type CreateMongoDBAccessPrivilegeInput struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Grants      []MongoDBGrant `json:"grants"`
}

type UpdateMongoDBAccessPrivilegeInput struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Grants      *[]MongoDBGrant `json:"grants,omitempty"`
}

func CreateMongoDBAccessPrivilege(ctx context.Context, c *Client, input *CreateMongoDBAccessPrivilegeInput) (*MongoDBAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/mongodb"
	var resp MongoDBAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetMongoDBAccessPrivilege(ctx context.Context, c *Client, id string) (*MongoDBAccessPrivilege, error) {
	path := fmt.Sprintf("%s/mongodb/%s", accessPrivilegesEndpoint, id)
	var resp MongoDBAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateMongoDBAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateMongoDBAccessPrivilegeInput) (*MongoDBAccessPrivilege, error) {
	path := fmt.Sprintf("%s/mongodb/%s", accessPrivilegesEndpoint, id)
	var resp MongoDBAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MySQL

type MySQLGrant struct {
	Privileges    []string `json:"privileges"`
	ResourceType  string   `json:"resource_type"`
	ResourceNames []string `json:"resource_names,omitempty"`
}

type MySQLAccessPrivilege struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Type        string       `json:"type,omitempty"`
	Grants      []MySQLGrant `json:"grants"`
}

type CreateMySQLAccessPrivilegeInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Grants      []MySQLGrant `json:"grants"`
}

type UpdateMySQLAccessPrivilegeInput struct {
	Name        *string       `json:"name,omitempty"`
	Description *string       `json:"description,omitempty"`
	Grants      *[]MySQLGrant `json:"grants,omitempty"`
}

func CreateMySQLAccessPrivilege(ctx context.Context, c *Client, input *CreateMySQLAccessPrivilegeInput) (*MySQLAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/mysql"
	var resp MySQLAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetMySQLAccessPrivilege(ctx context.Context, c *Client, id string) (*MySQLAccessPrivilege, error) {
	path := fmt.Sprintf("%s/mysql/%s", accessPrivilegesEndpoint, id)
	var resp MySQLAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateMySQLAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateMySQLAccessPrivilegeInput) (*MySQLAccessPrivilege, error) {
	path := fmt.Sprintf("%s/mysql/%s", accessPrivilegesEndpoint, id)
	var resp MySQLAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// OpenAI

type OpenAIPermission struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

type OpenAIAccessPrivilege struct {
	ID             string             `json:"id,omitempty"`
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	Type           string             `json:"type,omitempty"`
	PermissionType string             `json:"permission_type"`
	Permissions    []OpenAIPermission `json:"permissions,omitempty"`
}

type CreateOpenAIAccessPrivilegeInput struct {
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	PermissionType string             `json:"permission_type"`
	Permissions    []OpenAIPermission `json:"permissions,omitempty"`
}

type UpdateOpenAIAccessPrivilegeInput struct {
	Name           *string             `json:"name,omitempty"`
	Description    *string             `json:"description,omitempty"`
	PermissionType *string             `json:"permission_type,omitempty"`
	Permissions    *[]OpenAIPermission `json:"permissions,omitempty"`
}

func CreateOpenAIAccessPrivilege(ctx context.Context, c *Client, input *CreateOpenAIAccessPrivilegeInput) (*OpenAIAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/openai"
	var resp OpenAIAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetOpenAIAccessPrivilege(ctx context.Context, c *Client, id string) (*OpenAIAccessPrivilege, error) {
	path := fmt.Sprintf("%s/openai/%s", accessPrivilegesEndpoint, id)
	var resp OpenAIAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateOpenAIAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateOpenAIAccessPrivilegeInput) (*OpenAIAccessPrivilege, error) {
	path := fmt.Sprintf("%s/openai/%s", accessPrivilegesEndpoint, id)
	var resp OpenAIAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Grok

type GrokAccessPrivilege struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Endpoints   []string `json:"endpoints,omitempty"`
	Models      []string `json:"models,omitempty"`
}

type CreateGrokAccessPrivilegeInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Endpoints   []string `json:"endpoints,omitempty"`
	Models      []string `json:"models,omitempty"`
}

type UpdateGrokAccessPrivilegeInput struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Endpoints   *[]string `json:"endpoints,omitempty"`
	Models      *[]string `json:"models,omitempty"`
}

func CreateGrokAccessPrivilege(ctx context.Context, c *Client, input *CreateGrokAccessPrivilegeInput) (*GrokAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/grok"
	var resp GrokAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGrokAccessPrivilege(ctx context.Context, c *Client, id string) (*GrokAccessPrivilege, error) {
	path := fmt.Sprintf("%s/grok/%s", accessPrivilegesEndpoint, id)
	var resp GrokAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGrokAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateGrokAccessPrivilegeInput) (*GrokAccessPrivilege, error) {
	path := fmt.Sprintf("%s/grok/%s", accessPrivilegesEndpoint, id)
	var resp GrokAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Redis

type RedisGrant struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Name   string `json:"name"`
}

type RedisAccessPrivilege struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Type        string       `json:"type,omitempty"`
	Grants      []RedisGrant `json:"grants"`
	Keys        []string     `json:"keys"`
	Channels    []string     `json:"channels,omitempty"`
}

type CreateRedisAccessPrivilegeInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Grants      []RedisGrant `json:"grants"`
	Keys        []string     `json:"keys"`
	Channels    []string     `json:"channels,omitempty"`
}

type UpdateRedisAccessPrivilegeInput struct {
	Name        *string       `json:"name,omitempty"`
	Description *string       `json:"description,omitempty"`
	Grants      *[]RedisGrant `json:"grants,omitempty"`
	Keys        *[]string     `json:"keys,omitempty"`
	Channels    *[]string     `json:"channels,omitempty"`
}

func CreateRedisAccessPrivilege(ctx context.Context, c *Client, input *CreateRedisAccessPrivilegeInput) (*RedisAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/redis"
	var resp RedisAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetRedisAccessPrivilege(ctx context.Context, c *Client, id string) (*RedisAccessPrivilege, error) {
	path := fmt.Sprintf("%s/redis/%s", accessPrivilegesEndpoint, id)
	var resp RedisAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateRedisAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateRedisAccessPrivilegeInput) (*RedisAccessPrivilege, error) {
	path := fmt.Sprintf("%s/redis/%s", accessPrivilegesEndpoint, id)
	var resp RedisAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Apigee

type ApigeeAppConfig struct {
	DisplayName string `json:"display_name"`
}

type ApigeeAccessPrivilege struct {
	ID             string           `json:"id,omitempty"`
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	Type           string           `json:"type"`
	DeveloperEmail string           `json:"developer_email"`
	ProjectID      string           `json:"project_id"`
	APIProducts    []string         `json:"api_products"`
	AppName        *string          `json:"app_name,omitempty"`
	AppConfig      *ApigeeAppConfig `json:"app_config,omitempty"`
}

type CreateApigeeAccessPrivilegeInput struct {
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	DeveloperEmail string           `json:"developer_email"`
	ProjectID      string           `json:"project_id"`
	APIProducts    []string         `json:"api_products"`
	AppName        *string          `json:"app_name,omitempty"`
	AppConfig      *ApigeeAppConfig `json:"app_config,omitempty"`
}

type UpdateApigeeAccessPrivilegeInput struct {
	Name           *string          `json:"name,omitempty"`
	Description    *string          `json:"description,omitempty"`
	DeveloperEmail *string          `json:"developer_email,omitempty"`
	ProjectID      *string          `json:"project_id,omitempty"`
	APIProducts    *[]string        `json:"api_products,omitempty"`
	AppName        *string          `json:"app_name,omitempty"`
	AppConfig      *ApigeeAppConfig `json:"app_config,omitempty"`
}

func CreateApigeeAccessPrivilege(ctx context.Context, c *Client, input *CreateApigeeAccessPrivilegeInput) (*ApigeeAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/apigee"
	var resp ApigeeAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetApigeeAccessPrivilege(ctx context.Context, c *Client, id string) (*ApigeeAccessPrivilege, error) {
	path := fmt.Sprintf("%s/apigee/%s", accessPrivilegesEndpoint, id)
	var resp ApigeeAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateApigeeAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateApigeeAccessPrivilegeInput) (*ApigeeAccessPrivilege, error) {
	path := fmt.Sprintf("%s/apigee/%s", accessPrivilegesEndpoint, id)
	var resp ApigeeAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Elasticsearch

type ElasticsearchIndexPrivilege struct {
	Names      []string `json:"names"`
	Privileges []string `json:"privileges"`
}

type ElasticsearchGrant struct {
	Cluster []string                      `json:"cluster,omitempty"`
	Indices []ElasticsearchIndexPrivilege `json:"indices,omitempty"`
}

type ElasticsearchAccessPrivilege struct {
	ID          string             `json:"id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"type,omitempty"`
	Grant       ElasticsearchGrant `json:"grant"`
}

type CreateElasticsearchAccessPrivilegeInput struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Grant       ElasticsearchGrant `json:"grant"`
}

type UpdateElasticsearchAccessPrivilegeInput struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	Grant       *ElasticsearchGrant `json:"grant,omitempty"`
}

func CreateElasticsearchAccessPrivilege(ctx context.Context, c *Client, input *CreateElasticsearchAccessPrivilegeInput) (*ElasticsearchAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/elasticsearch"
	var resp ElasticsearchAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetElasticsearchAccessPrivilege(ctx context.Context, c *Client, id string) (*ElasticsearchAccessPrivilege, error) {
	path := fmt.Sprintf("%s/elasticsearch/%s", accessPrivilegesEndpoint, id)
	var resp ElasticsearchAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateElasticsearchAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateElasticsearchAccessPrivilegeInput) (*ElasticsearchAccessPrivilege, error) {
	path := fmt.Sprintf("%s/elasticsearch/%s", accessPrivilegesEndpoint, id)
	var resp ElasticsearchAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RabbitMQ

type RabbitmqPermissionEntry struct {
	Vhost     string `json:"vhost"`
	Configure string `json:"configure"`
	Write     string `json:"write"`
	Read      string `json:"read"`
}

type RabbitmqAccessPrivilege struct {
	ID          string                    `json:"id,omitempty"`
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Type        string                    `json:"type,omitempty"`
	Permissions []RabbitmqPermissionEntry `json:"permissions"`
	Tags        []string                  `json:"tags,omitempty"`
}

type CreateRabbitmqAccessPrivilegeInput struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Permissions []RabbitmqPermissionEntry `json:"permissions"`
	Tags        []string                  `json:"tags,omitempty"`
}

type UpdateRabbitmqAccessPrivilegeInput struct {
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Permissions *[]RabbitmqPermissionEntry `json:"permissions,omitempty"`
	Tags        *[]string                  `json:"tags,omitempty"`
}

func CreateRabbitmqAccessPrivilege(ctx context.Context, c *Client, input *CreateRabbitmqAccessPrivilegeInput) (*RabbitmqAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/rabbitmq"
	var resp RabbitmqAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetRabbitmqAccessPrivilege(ctx context.Context, c *Client, id string) (*RabbitmqAccessPrivilege, error) {
	path := fmt.Sprintf("%s/rabbitmq/%s", accessPrivilegesEndpoint, id)
	var resp RabbitmqAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateRabbitmqAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateRabbitmqAccessPrivilegeInput) (*RabbitmqAccessPrivilege, error) {
	path := fmt.Sprintf("%s/rabbitmq/%s", accessPrivilegesEndpoint, id)
	var resp RabbitmqAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GCP SA

type GCPSaConf struct {
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
}

type GCPSAAccessPrivilege struct {
	ID          string     `json:"id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"type,omitempty"`
	ProjectID   string     `json:"project_id,omitempty"`
	SaEmail     string     `json:"sa_email,omitempty"`
	SaConf      *GCPSaConf `json:"sa_conf,omitempty"`
}

type CreateGCPSAAccessPrivilegeInput struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	ProjectID   string     `json:"project_id"`
	SaEmail     string     `json:"sa_email,omitempty"`
	SaConf      *GCPSaConf `json:"sa_conf,omitempty"`
}

type UpdateGCPSAAccessPrivilegeInput struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	ProjectID   *string    `json:"project_id,omitempty"`
	SaEmail     *string    `json:"sa_email,omitempty"`
	SaConf      *GCPSaConf `json:"sa_conf,omitempty"`
}

func CreateGCPSAAccessPrivilege(ctx context.Context, c *Client, input *CreateGCPSAAccessPrivilegeInput) (*GCPSAAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/gcp_sa"
	var resp GCPSAAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetGCPSAAccessPrivilege(ctx context.Context, c *Client, id string) (*GCPSAAccessPrivilege, error) {
	path := fmt.Sprintf("%s/gcp_sa/%s", accessPrivilegesEndpoint, id)
	var resp GCPSAAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateGCPSAAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateGCPSAAccessPrivilegeInput) (*GCPSAAccessPrivilege, error) {
	path := fmt.Sprintf("%s/gcp_sa/%s", accessPrivilegesEndpoint, id)
	var resp GCPSAAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Azure App

type AzureAppRole struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

type AzureAppConfig struct {
	DisplayName string         `json:"display_name"`
	Roles       []AzureAppRole `json:"roles,omitempty"`
}

type AzureAppAccessPrivilege struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type,omitempty"`
	AppID       string          `json:"app_id,omitempty"`
	AppConfig   *AzureAppConfig `json:"app_config,omitempty"`
}

type CreateAzureAppAccessPrivilegeInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	AppID       string          `json:"app_id,omitempty"`
	AppConfig   *AzureAppConfig `json:"app_config,omitempty"`
}

type UpdateAzureAppAccessPrivilegeInput struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	AppID       *string         `json:"app_id,omitempty"`
	AppConfig   *AzureAppConfig `json:"app_config,omitempty"`
}

func CreateAzureAppAccessPrivilege(ctx context.Context, c *Client, input *CreateAzureAppAccessPrivilegeInput) (*AzureAppAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/azure_app"
	var resp AzureAppAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAzureAppAccessPrivilege(ctx context.Context, c *Client, id string) (*AzureAppAccessPrivilege, error) {
	path := fmt.Sprintf("%s/azure_app/%s", accessPrivilegesEndpoint, id)
	var resp AzureAppAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateAzureAppAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateAzureAppAccessPrivilegeInput) (*AzureAppAccessPrivilege, error) {
	path := fmt.Sprintf("%s/azure_app/%s", accessPrivilegesEndpoint, id)
	var resp AzureAppAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AWS Access Key

type AWSAccessKeyAccessPrivilege struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Policies    []string `json:"policies"`
}

type CreateAWSAccessKeyAccessPrivilegeInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Policies    []string `json:"policies"`
}

type UpdateAWSAccessKeyAccessPrivilegeInput struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Policies    *[]string `json:"policies,omitempty"`
}

func CreateAWSAccessKeyAccessPrivilege(ctx context.Context, c *Client, input *CreateAWSAccessKeyAccessPrivilegeInput) (*AWSAccessKeyAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/aws_access_key"
	var resp AWSAccessKeyAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetAWSAccessKeyAccessPrivilege(ctx context.Context, c *Client, id string) (*AWSAccessKeyAccessPrivilege, error) {
	path := fmt.Sprintf("%s/aws_access_key/%s", accessPrivilegesEndpoint, id)
	var resp AWSAccessKeyAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateAWSAccessKeyAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateAWSAccessKeyAccessPrivilegeInput) (*AWSAccessKeyAccessPrivilege, error) {
	path := fmt.Sprintf("%s/aws_access_key/%s", accessPrivilegesEndpoint, id)
	var resp AWSAccessKeyAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Twilio

type TwilioAccessPrivilege struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Type           string   `json:"type,omitempty"`
	PermissionType string   `json:"permission_type"`
	Permissions    []string `json:"permissions,omitempty"`
}

type CreateTwilioAccessPrivilegeInput struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	PermissionType string   `json:"permission_type"`
	Permissions    []string `json:"permissions,omitempty"`
}

type UpdateTwilioAccessPrivilegeInput struct {
	Name           *string   `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	PermissionType *string   `json:"permission_type,omitempty"`
	Permissions    *[]string `json:"permissions,omitempty"`
}

func CreateTwilioAccessPrivilege(ctx context.Context, c *Client, input *CreateTwilioAccessPrivilegeInput) (*TwilioAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/twilio"
	var resp TwilioAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetTwilioAccessPrivilege(ctx context.Context, c *Client, id string) (*TwilioAccessPrivilege, error) {
	path := fmt.Sprintf("%s/twilio/%s", accessPrivilegesEndpoint, id)
	var resp TwilioAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateTwilioAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateTwilioAccessPrivilegeInput) (*TwilioAccessPrivilege, error) {
	path := fmt.Sprintf("%s/twilio/%s", accessPrivilegesEndpoint, id)
	var resp TwilioAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Snowflake

type SnowflakeGrant struct {
	Privileges    []string `json:"privileges"`
	ResourceType  string   `json:"resource_type"`
	ResourceNames []string `json:"resource_names,omitempty"`
}

type SnowflakeAccessPrivilege struct {
	ID          string           `json:"id,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Type        string           `json:"type,omitempty"`
	Grants      []SnowflakeGrant `json:"grants"`
}

type CreateSnowflakeAccessPrivilegeInput struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Grants      []SnowflakeGrant `json:"grants"`
}

type UpdateSnowflakeAccessPrivilegeInput struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Grants      *[]SnowflakeGrant `json:"grants,omitempty"`
}

func CreateSnowflakeAccessPrivilege(ctx context.Context, c *Client, input *CreateSnowflakeAccessPrivilegeInput) (*SnowflakeAccessPrivilege, error) {
	path := accessPrivilegesEndpoint + "/snowflake"
	var resp SnowflakeAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPost, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetSnowflakeAccessPrivilege(ctx context.Context, c *Client, id string) (*SnowflakeAccessPrivilege, error) {
	path := fmt.Sprintf("%s/snowflake/%s", accessPrivilegesEndpoint, id)
	var resp SnowflakeAccessPrivilege
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func UpdateSnowflakeAccessPrivilege(ctx context.Context, c *Client, id string, input *UpdateSnowflakeAccessPrivilegeInput) (*SnowflakeAccessPrivilege, error) {
	path := fmt.Sprintf("%s/snowflake/%s", accessPrivilegesEndpoint, id)
	var resp SnowflakeAccessPrivilege
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Shared delete function for all access privileges

func DeleteAccessPrivilege(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", accessPrivilegesEndpoint, id)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	return nil
}
