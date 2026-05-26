package confluence_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Manages a Hush Security Confluence integration for scanning Confluence spaces for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: confluenceIntegrationCreate,
		ReadContext:   confluenceIntegrationRead,
		UpdateContext: confluenceIntegrationUpdate,
		DeleteContext: confluenceIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ConfluenceIntegrationResourceSchema(),
	}
}

func confluenceIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	apiKey := d.Get("api_key").(string)
	if apiKey == "" {
		apiKey = d.Get("api_key_wo").(string)
	}

	input := &client.CreateConfluenceIntegrationInput{
		Name:      d.Get("name").(string),
		OrgDomain: d.Get("org_domain").(string),
		User:      d.Get("user").(string),
		ApiKey:    apiKey,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	resp, err := client.CreateConfluenceIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return confluenceIntegrationRead(ctx, d, m)
}

func confluenceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle api_key rotation separately
	if d.HasChanges("api_key", "api_key_wo", "api_key_wo_version", "user") {
		apiKey := d.Get("api_key").(string)
		if apiKey == "" {
			apiKey = d.Get("api_key_wo").(string)
		}

		if apiKey != "" {
			replaceInput := &client.ReplaceConfluenceApiKeyInput{
				User:   d.Get("user").(string),
				ApiKey: apiKey,
			}
			if err := client.ReplaceConfluenceApiKey(ctx, c, d.Id(), replaceInput); err != nil {
				errResponse, ok := err.(*client.APIError)
				if ok && errResponse.StatusCode == http.StatusNotFound {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}
		}
	}

	// Handle metadata updates
	input := &client.UpdateConfluenceIntegrationInput{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateConfluenceIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return confluenceIntegrationRead(ctx, d, m)
}

func confluenceIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteConfluenceIntegration(ctx, c, d.Id())
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
