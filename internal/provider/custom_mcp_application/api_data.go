package custom_mcp_application

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/dsschema"
)

func APIDataSource() *schema.Resource {
	s := dsschema.FromResource(APIResourceSchema(), dsschema.Options{
		Drop: []string{"auth.secret", "auth.secret_wo", "auth.secret_wo_version"},
		Descriptions: map[string]string{
			"tool":    "The API's tools, as detected or as written.",
			"headers": "Extra HTTP headers the gateway sends with every request to the API.",
			"auth":    "The fixed credential the gateway presents to the API, without its secret; empty when there is none.",
		},
	})
	s["name"].Required = true
	s["name"].Computed = false
	s["name"].Description = dataNameDesc
	return &schema.Resource{
		Description: "A custom API app, read by name, with its tools. Its secret is never " +
			"returned, so the data source has no attribute for it.",
		ReadContext: customAPIAppDataRead,
		Schema:      s,
	}
}

func customAPIAppDataRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	name := d.Get("name").(string)

	app, err := client.GetCustomAPIApplication(ctx, c, name)
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return diag.Errorf("no custom API app found with name: %s", name)
		}
		return diag.FromErr(err)
	}
	tools, err := client.GetCustomAPIApplicationTools(ctx, c, name)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(app.Name)
	if err := flattenAPI(d, app, tools, false); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
