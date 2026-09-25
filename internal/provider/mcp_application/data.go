package mcp_application

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/dsschema"
)

const (
	dataSourceDescription = "An MCP application, read by id or by display name. The client secret is " +
		"never returned, so the data source has no attribute for it."
	dataIDDesc          = "The application's id. One of id or display_name is required."
	dataDisplayNameDesc = "The name the application is shown under, which is unique within the organization. One of id or display_name is required."
)

func DataSource() *schema.Resource {
	s := dsschema.FromResource(ResourceSchema(), dsschema.Options{
		Drop: []string{"client_secret", "client_secret_wo", "client_secret_wo_version"},
		Descriptions: map[string]string{
			"app_catalog_id":    "The catalog entry the application was created from, or 'custom-<name>' for a custom app.",
			"deployment_ids":    "The deployments whose gateways serve the application.",
			"url_label":         "The label of the entry's address the application uses, lowercased, or empty when the entry has one address.",
			"allowed_agents":    "The agent types the application is limited to; empty when any agent may use it.",
			"scopes":            "The OAuth scopes the application requests from the server.",
			"client_id":         "The client id of the OAuth app registered with the server, or empty when there is none.",
			"google_project_id": "The Google Cloud project a Google Workspace server bills its quota to; empty for other entries.",
			"quickbooks":        "The QuickBooks company, for the 'quickbooks' entry.",
			"assign_all":        "Whether the application is granted to every user of the organization.",
			"assignment":        "The rules granting the application: it is granted by any rule that matches, and a rule matches when all of its conditions do.",
			"tool_defaults":     "What the gateway does with a call to a tool of each class, unless the tool has an operation of its own.",
			"tool_operation":    "The tools with an operation of their own, overriding their class default.",
		},
	})
	s["id"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	dsschema.Lookup(s, "id", dataIDDesc, "display_name")
	dsschema.Lookup(s, "display_name", dataDisplayNameDesc, "id")
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: applicationDataRead,
		Schema:      s,
	}
}

func applicationDataRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	id, appCatalogID := d.Get("id").(string), ""
	if id == "" {
		name, ok := d.GetOk("display_name")
		if !ok {
			return diag.Errorf("one of `id` or `display_name` must be specified")
		}
		apps, err := client.ListMCPApplications(ctx, c)
		if err != nil {
			return diag.FromErr(err)
		}
		// heimdall has no display name filter, so the list is matched here.
		var matches []client.MCPApplication
		for _, app := range apps {
			if app.DisplayName == name.(string) {
				matches = append(matches, app)
			}
		}
		switch len(matches) {
		case 0:
			return diag.Errorf("no MCP application found with display_name: %s", name)
		case 1:
		default:
			return diag.Errorf("multiple MCP applications found with display_name: %s, please use id instead", name)
		}
		// The list item already says which route serves the application.
		id, appCatalogID = matches[0].ID, matches[0].AppCatalogID
	}

	app, err := readMCPApplication(ctx, c, id, appCatalogID)
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return diag.Errorf("no MCP application found with id: %s", id)
		}
		return diag.FromErr(err)
	}
	d.SetId(app.ID)
	if err := flatten(d, app); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
