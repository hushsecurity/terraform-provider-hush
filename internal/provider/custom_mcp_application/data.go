package custom_mcp_application

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/dsschema"
)

const (
	dataSourceDescription = "A custom MCP server, read by name. Its secrets are never returned, so the " +
		"data source has no attribute for them; `auth` shows which credential is configured."
	dataNameDesc = "The custom app's name."
)

func DataSource() *schema.Resource {
	s := dsschema.FromResource(ResourceSchema(), dsschema.Options{
		Drop: []string{
			"client_secret", "client_secret_wo", "client_secret_wo_version",
			"auth.secret", "auth.secret_wo", "auth.secret_wo_version",
		},
		Descriptions: map[string]string{
			"url":       "The addresses the server is reachable at. When there are several, an application picks one by `url_label`.",
			"url.label": "The label that names the address, or empty when there is a single one.",
			"tool":      "The tools the server offers, as declared or as the gateway detected them.",
			"headers":   "Extra HTTP headers the gateway sends with every request to the server.",
			"client_id": "The client id of the OAuth app registered with the server, or empty when there is none.",
			"auth":      "The fixed credential the gateway presents to the server, without its secret; empty when there is none.",
			"auth.type": "'bearer', 'basic' or 'header'.",
			"auth.name": "The header the credential is sent in, for type 'header'.",
		},
	})
	s["name"].Required = true
	s["name"].Computed = false
	s["name"].Description = dataNameDesc
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: customAppDataRead,
		Schema:      s,
	}
}

func customAppDataRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	name := d.Get("name").(string)

	app, err := client.GetCustomMCPApplication(ctx, c, name)
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return diag.Errorf("no custom MCP app found with name: %s", name)
		}
		return diag.FromErr(err)
	}
	d.SetId(app.Name)
	if err := flatten(d, app, false); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
