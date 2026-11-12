package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	APIToolMethods      = []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD", "PATCH", "TRACE"}
	APIToolContentTypes = []string{"application/json", "application/octet-stream"}
)

// APITool is one REST operation an api app serves as a tool.
type APITool struct {
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Enabled     bool            `json:"enabled"`
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	InputSchema json.RawMessage `json:"input_schema"`
	ContentType string          `json:"content_type"`
	OperationID *string         `json:"operation_id"`
}

type CustomAPIApplication struct {
	Name           string            `json:"name"`
	AppCatalogID   string            `json:"app_catalog_id"`
	DisplayName    string            `json:"display_name"`
	Description    *string           `json:"description"`
	URL            string            `json:"url"`
	SpecURL        *string           `json:"spec_url"`
	DetectRequired bool              `json:"detect_required"`
	Headers        map[string]string `json:"headers"`
	Auth           *MCPAuth          `json:"auth"`
}

// CustomAPIApplicationInput is the create body; Auth is built with MCPAuthInput.
type CustomAPIApplicationInput struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Description *string           `json:"description,omitempty"`
	URL         string            `json:"url"`
	SpecURL     *string           `json:"spec_url,omitempty"`
	Headers     map[string]string `json:"headers"`
	Auth        map[string]any    `json:"auth,omitempty"`
}

// CustomAPIApplicationUpdate is a patch, as CustomMCPApplicationUpdate.
type CustomAPIApplicationUpdate struct {
	DisplayName *string            `json:"display_name,omitempty"`
	Description **string           `json:"description,omitempty"`
	URL         *string            `json:"url,omitempty"`
	SpecURL     **string           `json:"spec_url,omitempty"`
	Headers     *map[string]string `json:"headers,omitempty"`
	Auth        *map[string]any    `json:"auth,omitempty"`
}

func CreateCustomAPIApplication(ctx context.Context, c *Client, input *CustomAPIApplicationInput) (*CustomAPIApplication, error) {
	var app CustomAPIApplication
	if err := c.doRequest(ctx, http.MethodPost, customMCPApplicationsEndpoint+"/api", input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func GetCustomAPIApplication(ctx context.Context, c *Client, name string) (*CustomAPIApplication, error) {
	var app CustomAPIApplication
	path := fmt.Sprintf("%s/%s/api", customMCPApplicationsEndpoint, name)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func UpdateCustomAPIApplication(ctx context.Context, c *Client, name string, input *CustomAPIApplicationUpdate) (*CustomAPIApplication, error) {
	var app CustomAPIApplication
	path := fmt.Sprintf("%s/%s/api", customMCPApplicationsEndpoint, name)
	if err := c.doRequest(ctx, http.MethodPatch, path, input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func GetCustomAPIApplicationTools(ctx context.Context, c *Client, name string) ([]APITool, error) {
	var tools []APITool
	path := fmt.Sprintf("%s/%s/api/tools", customMCPApplicationsEndpoint, name)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &tools); err != nil {
		return nil, err
	}
	return tools, nil
}

// SetCustomAPIApplicationTools replaces the whole list. It answers 409 when
// applications use the app and no tool is enabled.
func SetCustomAPIApplicationTools(ctx context.Context, c *Client, name string, tools []APITool) error {
	path := fmt.Sprintf("%s/%s/api/tools", customMCPApplicationsEndpoint, name)
	return c.doRequest(ctx, http.MethodPut, path, tools, nil)
}
