package client

import (
	"context"
	"fmt"
	"net/http"
)

// CustomAppCatalogPrefix namespaces a custom app's id in the catalog: an
// application is created from custom app "foo" as app_catalog_id
// "custom-foo", which can never collide with a built-in entry.
const CustomAppCatalogPrefix = "custom-"

// The kinds of fixed credential an MCP server can take in place of OAuth.
const (
	MCPAuthTypeBearer = "bearer"
	MCPAuthTypeBasic  = "basic"
	MCPAuthTypeHeader = "header"
)

var MCPAuthTypes = []string{MCPAuthTypeBearer, MCPAuthTypeBasic, MCPAuthTypeHeader}

// MCPAuthSecretField names, per auth type, the member that carries the secret.
var MCPAuthSecretField = map[string]string{
	MCPAuthTypeBearer: "token",
	MCPAuthTypeBasic:  "password",
	MCPAuthTypeHeader: "value",
}

// MCPAuth is the non-secret part of a fixed credential, which is what the API
// returns: the secret itself comes back masked, and is not decoded.
type MCPAuth struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"` // basic only
	Name     string `json:"name,omitempty"`     // header only
}

// CustomMCPTool is a tool as a custom app describes it. The operation the
// gateway applies to it belongs to each application, not to the custom app.
type CustomMCPTool struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type CustomMCPApplication struct {
	Name         string            `json:"name"`
	AppCatalogID string            `json:"app_catalog_id"`
	DisplayName  string            `json:"display_name"`
	Description  *string           `json:"description"`
	URLs         []URLOption       `json:"urls"`
	Scopes       []string          `json:"scopes"`
	Tools        []CustomMCPTool   `json:"tools"`
	OAuthRelay   bool              `json:"oauth_relay"`
	Headers      map[string]string `json:"headers"`
	ClientID     *string           `json:"client_id"`
	Auth         *MCPAuth          `json:"auth"`
}

// CustomMCPApplicationInput is the create body. Auth is the whole credential,
// secret included, keyed as the API expects for its type: build it with
// MCPAuthInput.
type CustomMCPApplicationInput struct {
	Name         string            `json:"name"`
	DisplayName  string            `json:"display_name"`
	Description  *string           `json:"description,omitempty"`
	URLs         []URLOption       `json:"urls"`
	Scopes       []string          `json:"scopes"`
	Tools        []CustomMCPTool   `json:"tools,omitempty"`
	OAuthRelay   bool              `json:"oauth_relay"`
	Headers      map[string]string `json:"headers"`
	ClientID     *string           `json:"client_id,omitempty"`
	ClientSecret *string           `json:"client_secret,omitempty"`
	Auth         map[string]string `json:"auth,omitempty"`
}

// MCPAuthInput builds the credential a create or update sends.
func MCPAuthInput(auth MCPAuth, secret string) map[string]string {
	out := map[string]string{"type": auth.Type, MCPAuthSecretField[auth.Type]: secret}
	if auth.Username != "" {
		out["username"] = auth.Username
	}
	if auth.Name != "" {
		out["name"] = auth.Name
	}
	return out
}

// CustomMCPApplicationUpdate is a patch: a key that is absent is left as it
// is, so every field is sent only when it changed. Description, ClientID,
// ClientSecret and Auth can also be cleared, which takes an explicit null --
// hence the double pointer: nil leaves the field out, a pointer to nil sends
// null.
type CustomMCPApplicationUpdate struct {
	DisplayName  *string            `json:"display_name,omitempty"`
	Description  **string           `json:"description,omitempty"`
	URLs         []URLOption        `json:"urls,omitempty"`
	Scopes       *[]string          `json:"scopes,omitempty"`
	Tools        *[]CustomMCPTool   `json:"tools,omitempty"`
	OAuthRelay   *bool              `json:"oauth_relay,omitempty"`
	Headers      *map[string]string `json:"headers,omitempty"`
	ClientID     **string           `json:"client_id,omitempty"`
	ClientSecret **string           `json:"client_secret,omitempty"`
	Auth         *map[string]string `json:"auth,omitempty"`
}

const customMCPApplicationsEndpoint = "/v1/applications/custom"

func CreateCustomMCPApplication(ctx context.Context, c *Client, input *CustomMCPApplicationInput) (*CustomMCPApplication, error) {
	var app CustomMCPApplication
	if err := c.doRequest(ctx, http.MethodPost, customMCPApplicationsEndpoint+"/mcp", input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func GetCustomMCPApplication(ctx context.Context, c *Client, name string) (*CustomMCPApplication, error) {
	var app CustomMCPApplication
	path := fmt.Sprintf("%s/%s/mcp", customMCPApplicationsEndpoint, name)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateCustomMCPApplication also rewrites every application created from
// the custom app: heimdall cascades addresses, tools, scopes, headers and
// credentials to them.
func UpdateCustomMCPApplication(ctx context.Context, c *Client, name string, input *CustomMCPApplicationUpdate) (*CustomMCPApplication, error) {
	var app CustomMCPApplication
	path := fmt.Sprintf("%s/%s/mcp", customMCPApplicationsEndpoint, name)
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteCustomMCPApplication answers 409 while any application still uses
// the custom app.
func DeleteCustomMCPApplication(ctx context.Context, c *Client, name string) error {
	path := fmt.Sprintf("%s/%s", customMCPApplicationsEndpoint, name)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
