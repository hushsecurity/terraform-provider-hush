package infisical_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

const resourceDescription = "Manages a Hush Security Infisical integration for scanning Infisical secrets management for exposed secrets."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: infisicalIntegrationCreate,
		ReadContext:   infisicalIntegrationRead,
		UpdateContext: infisicalIntegrationUpdate,
		DeleteContext: infisicalIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: InfisicalIntegrationResourceSchema(),
	}
}

func infisicalIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	clientSecret := writeonly.GetString(d, "client_secret", "client_secret_wo")

	input := &client.CreateInfisicalIntegrationInput{
		Name:         d.Get("name").(string),
		BaseURL:      d.Get("base_url").(string),
		ClientID:     d.Get("client_id").(string),
		ClientSecret: clientSecret,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	resp, err := client.CreateInfisicalIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return infisicalIntegrationRead(ctx, d, m)
}

func infisicalIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle credentials rotation separately
	if d.HasChanges("client_id", "client_secret", "client_secret_wo", "client_secret_wo_version") {
		clientSecret := writeonly.GetString(d, "client_secret", "client_secret_wo")

		if clientSecret != "" {
			replaceInput := &client.ReplaceInfisicalCredentialsInput{
				ClientID:     d.Get("client_id").(string),
				ClientSecret: clientSecret,
			}
			if err := client.ReplaceInfisicalCredentials(ctx, c, d.Id(), replaceInput); err != nil {
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
	input := &client.UpdateInfisicalIntegrationInput{}
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
	if d.HasChange("base_url") {
		baseURL := d.Get("base_url").(string)
		input.BaseURL = &baseURL
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateInfisicalIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return infisicalIntegrationRead(ctx, d, m)
}

func infisicalIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteInfisicalIntegration(ctx, c, d.Id())
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
