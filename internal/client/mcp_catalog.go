package client

import (
	"context"
	"fmt"
	"net/http"
)

// The classes a tool belongs to, each with a default operation.
var ToolTypes = []string{"read", "write", "destructive"}

// MCPTool is one tool an MCP server offers, as the catalog and applications
// describe it. Type is its class (read, write, destructive); Operation is what
// the gateway does when an agent calls it, and is empty wherever the tool
// takes its class default instead.
type MCPTool struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Operation   *string `json:"operation,omitempty"`
}

// URLOption is one address an MCP server is reachable at. Label tells the
// options apart when there is more than one, and is empty when there is not.
type URLOption struct {
	Label *string `json:"label"`
	URL   string  `json:"url"`
}

// MCPCatalogEntry is a built-in MCP server an application can be created from.
type MCPCatalogEntry struct {
	AppCatalogID       string      `json:"app_catalog_id"`
	DisplayName        string      `json:"display_name"`
	Category           string      `json:"category"`
	URLs               []URLOption `json:"urls"`
	Hosted             bool        `json:"hosted"`
	Scopes             []string    `json:"scopes"`
	Tools              []MCPTool   `json:"tools"`
	ManualRegistration bool        `json:"manual_registration"`
	OAuthRelay         bool        `json:"oauth_relay"`
}

// GetMCPCatalogEntry reads one entry of the built-in catalog. A custom app's
// id (custom-<name>) is not in it and answers 404.
func GetMCPCatalogEntry(ctx context.Context, c *Client, appCatalogID string) (*MCPCatalogEntry, error) {
	var entry MCPCatalogEntry
	path := fmt.Sprintf("/v1/applications/catalog/mcp/%s", appCatalogID)
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
