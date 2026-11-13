package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ResourceScope struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type TokenClaim struct {
	Name      string `json:"name"`
	Attribute string `json:"attribute"`
}

type ResourceServer struct {
	Name            string          `json:"name"`
	DisplayName     string          `json:"display_name"`
	Description     *string         `json:"description"`
	Audience        string          `json:"audience"`
	ScopesSupported []ResourceScope `json:"scopes_supported"`
	TokenLifetime   *int            `json:"token_lifetime"`
	TokenClaims     []TokenClaim    `json:"token_claims"`
}

const resourceServersPath = "/v1/applications/resource_servers"

func CreateResourceServer(ctx context.Context, c *Client, input *ResourceServer) (*ResourceServer, error) {
	var out ResourceServer
	if err := c.doRequest(ctx, http.MethodPost, resourceServersPath, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func GetResourceServer(ctx context.Context, c *Client, name string) (*ResourceServer, error) {
	var out ResourceServer
	if err := c.doRequest(ctx, http.MethodGet, resourceServersPath+"/"+url.PathEscape(name), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateResourceServer sends every field but the name, which is fixed.
func UpdateResourceServer(ctx context.Context, c *Client, input *ResourceServer) (*ResourceServer, error) {
	body := map[string]any{
		"display_name":     input.DisplayName,
		"description":      input.Description,
		"audience":         input.Audience,
		"scopes_supported": input.ScopesSupported,
		"token_lifetime":   input.TokenLifetime,
		"token_claims":     input.TokenClaims,
	}
	var out ResourceServer
	if err := c.doRequest(ctx, http.MethodPatch, resourceServersPath+"/"+url.PathEscape(input.Name), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func DeleteResourceServer(ctx context.Context, c *Client, name string) error {
	return c.doRequest(ctx, http.MethodDelete, resourceServersPath+"/"+url.PathEscape(name), nil, nil)
}

type ResourceApplicationInput struct {
	DisplayName   string   `json:"display_name"`
	Description   *string  `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	Enabled       bool     `json:"enabled"`
	Resource      string   `json:"resource"`
	Scopes        []string `json:"scopes"`
}

func CreateResourceApplication(ctx context.Context, c *Client, input *ResourceApplicationInput) (*MCPApplication, error) {
	var app MCPApplication
	if err := c.doRequest(ctx, http.MethodPost, "/v1/applications/resource", input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func UpdateResourceApplication(ctx context.Context, c *Client, id string, input *MCPApplicationUpdate) (*MCPApplication, error) {
	var app MCPApplication
	if err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/v1/applications/%s/resource", id), input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}
