package client

import (
	"context"
	"fmt"
	"net/http"
	"slices"
)

// Catalog entries heimdall serves through their own routes, because the
// application carries settings the generic MCP routes have no field for.
var (
	// A Google Workspace server, which bills its quota to a Google Cloud
	// project the application names.
	GoogleMCPCatalogIDs = []string{"gcalendar", "gdocs", "gdrive", "gmail", "gpeople", "gsheets", "gslides"}
	// QuickBooks, whose server is scoped to one company.
	QuickBooksMCPCatalogID = "quickbooks"
)

// The operations the gateway can apply to a tool call.
var ToolOperations = []string{"allow", "block", "user_consent"}

// The agent types an application can be limited to.
var AgentTypes = []string{"claude-code", "claude", "cursor", "windsurf", "vscode", "openclaw"}

type ToolGroup struct {
	Type      string `json:"type"`
	Operation string `json:"operation"`
}

type MCPApplication struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	Type               string       `json:"type"`
	AppCatalogID       string       `json:"app_catalog_id"`
	CatalogDisplayName *string      `json:"catalog_display_name"`
	DisplayName        string       `json:"display_name"`
	Description        *string      `json:"description"`
	DeploymentIDs      []string     `json:"deployment_ids"`
	AllowedAgents      []string     `json:"allowed_agents"`
	Enabled            bool         `json:"enabled"`
	Assignments        []Assignment `json:"assignments"`
	AssignAll          bool         `json:"assign_all"`
	URL                *string      `json:"url"`
	Hosted             bool         `json:"hosted"`
	Scopes             []string     `json:"scopes"`
	ToolGroups         []ToolGroup  `json:"tool_groups"`
	Tools              []MCPTool    `json:"tools"`
	ClientID           *string      `json:"client_id"`
	OAuthRelay         bool         `json:"oauth_relay"`

	// Returned by the Google and QuickBooks routes only.
	GoogleProjectID     *string `json:"google_project_id,omitempty"`
	QuickBooksCompanyID *string `json:"quickbooks_company_id,omitempty"`
	QuickBooksSandbox   *bool   `json:"quickbooks_sandbox,omitempty"`
}

// The sides of a request an assignment predicate can look at, and the ways
// it can compare.
var (
	AssignmentSources = []string{"user", "agent"}
	AssignmentOps     = []string{"eq", "neq", "in", "nin"}
)

// Assignment grants an application to the users and agents it matches: all
// of its conditions must hold, and a condition holds when any of its
// predicates does. The create takes none; they are patched in afterwards.
type Assignment struct {
	AllOf []AssignmentCondition `json:"all_of"`
}

type AssignmentCondition struct {
	AnyOf []AssignmentPredicate `json:"any_of"`
}

// AssignmentPredicate compares one property of the user or the agent. Value
// is a string for eq and neq, and a list of strings for in and nin.
type AssignmentPredicate struct {
	Source   string `json:"source"`
	Property string `json:"property"`
	Op       string `json:"op"`
	Value    any    `json:"value"`
}

// MCPApplicationInput is the create body. The Google and QuickBooks fields go
// only to their own routes, whose models require them; the generic route
// refuses them, hence omitempty.
type MCPApplicationInput struct {
	DisplayName   string   `json:"display_name"`
	Description   *string  `json:"description,omitempty"`
	DeploymentIDs []string `json:"deployment_ids"`
	AllowedAgents []string `json:"allowed_agents,omitempty"`
	Enabled       bool     `json:"enabled"`
	ClientID      *string  `json:"client_id,omitempty"`
	ClientSecret  *string  `json:"client_secret,omitempty"`
	URLLabel      *string  `json:"url_label,omitempty"`

	GoogleProjectID     *string `json:"google_project_id,omitempty"`
	QuickBooksCompanyID *string `json:"quickbooks_company_id,omitempty"`
	QuickBooksSandbox   *bool   `json:"quickbooks_sandbox,omitempty"`
}

// MCPApplicationUpdate is a patch: an absent key is left as it is. Fields that
// can be cleared are double pointers, so a pointer to nil sends null.
type MCPApplicationUpdate struct {
	DisplayName   *string       `json:"display_name,omitempty"`
	Description   **string      `json:"description,omitempty"`
	DeploymentIDs *[]string     `json:"deployment_ids,omitempty"`
	Scopes        *[]string     `json:"scopes,omitempty"`
	Assignments   *[]Assignment `json:"assignments,omitempty"`
	AssignAll     *bool         `json:"assign_all,omitempty"`
	AllowedAgents **[]string    `json:"allowed_agents,omitempty"`
	Enabled       *bool         `json:"enabled,omitempty"`
	ClientID      **string      `json:"client_id,omitempty"`
	ClientSecret  **string      `json:"client_secret,omitempty"`

	GoogleProjectID     *string `json:"google_project_id,omitempty"`
	QuickBooksCompanyID *string `json:"quickbooks_company_id,omitempty"`
	QuickBooksSandbox   *bool   `json:"quickbooks_sandbox,omitempty"`
}

// ToolOperationChange sets one tool's operation; a nil Operation returns the
// tool to its class default.
type ToolOperationChange struct {
	Name      string  `json:"name"`
	Operation *string `json:"operation"`
}

type ToolGroupOperationChange struct {
	Type      string `json:"type"`
	Operation string `json:"operation"`
}

func IsGoogleMCPCatalogID(appCatalogID string) bool {
	return slices.Contains(GoogleMCPCatalogIDs, appCatalogID)
}

// mcpApplicationPath is where an application is read and patched. An entry
// with its own routes must be reached through them: the generic route neither
// returns nor accepts the fields only they carry.
func mcpApplicationPath(id, appCatalogID string) string {
	path := fmt.Sprintf("/v1/applications/%s/mcp", id)
	if IsGoogleMCPCatalogID(appCatalogID) || appCatalogID == QuickBooksMCPCatalogID {
		path += "/" + appCatalogID
	}
	return path
}

// CreateMCPApplication creates an application from a catalog entry, or from
// a custom app by its custom-<name> id. The route is the same either way:
// heimdall registers the Google and QuickBooks ones ahead of the generic one.
func CreateMCPApplication(ctx context.Context, c *Client, appCatalogID string, input *MCPApplicationInput) (*MCPApplication, error) {
	var app MCPApplication
	path := fmt.Sprintf("/v1/applications/mcp/%s", appCatalogID)
	if err := c.doRequest(ctx, http.MethodPost, path, input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func GetMCPApplication(ctx context.Context, c *Client, id, appCatalogID string) (*MCPApplication, error) {
	var app MCPApplication
	if err := c.doRequest(ctx, http.MethodGet, mcpApplicationPath(id, appCatalogID), nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// GetApplication reads any application by id alone, which is how an import
// learns the catalog entry that decides the path everything else uses.
func GetApplication(ctx context.Context, c *Client, id string) (*MCPApplication, error) {
	var app MCPApplication
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/applications/%s", id), nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func UpdateMCPApplication(ctx context.Context, c *Client, id, appCatalogID string, input *MCPApplicationUpdate) (*MCPApplication, error) {
	var app MCPApplication
	if err := c.doRequest(ctx, http.MethodPatch, mcpApplicationPath(id, appCatalogID), input, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func DeleteApplication(ctx context.Context, c *Client, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/v1/applications/%s", id), nil, nil)
}

// ChangeMCPToolOperations answers 404 for a tool the application does not
// have.
func ChangeMCPToolOperations(ctx context.Context, c *Client, id string, changes []ToolOperationChange) (*MCPApplication, error) {
	var app MCPApplication
	path := fmt.Sprintf("/v1/applications/%s/mcp/tool_operations", id)
	if err := c.doRequest(ctx, http.MethodPost, path, changes, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func ChangeMCPToolGroupOperations(ctx context.Context, c *Client, id string, changes []ToolGroupOperationChange) (*MCPApplication, error) {
	var app MCPApplication
	path := fmt.Sprintf("/v1/applications/%s/mcp/tool_group_operations", id)
	if err := c.doRequest(ctx, http.MethodPost, path, changes, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

type mcpApplicationListResponse struct {
	Items    []MCPApplication `json:"items"`
	NextPage *string          `json:"next_page"`
}

// ListMCPApplications lists the organization's MCP applications. The list
// carries the summary fields only -- read one through GetMCPApplication for
// the rest.
func ListMCPApplications(ctx context.Context, c *Client) ([]MCPApplication, error) {
	return collectPages(func(cursor string) ([]MCPApplication, *string, error) {
		var resp mcpApplicationListResponse
		if err := c.doRequest(ctx, http.MethodGet, withCursor("/v1/applications?type=mcp", cursor), nil, &resp); err != nil {
			return nil, nil, err
		}
		return resp.Items, resp.NextPage, nil
	})
}
