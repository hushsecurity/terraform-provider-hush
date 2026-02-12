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

// Shared delete function for all access privileges

func DeleteAccessPrivilege(ctx context.Context, c *Client, id string) error {
	path := fmt.Sprintf("%s/%s", accessPrivilegesEndpoint, id)
	if err := c.doRequest(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return err
	}
	return nil
}
