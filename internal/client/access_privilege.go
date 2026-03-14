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

// Shared delete function for all access privileges

func DeleteAccessPrivilege(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", accessPrivilegesEndpoint, id)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	return nil
}
