package sonatype_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

const resourceDescription = "Manages a Hush Security Sonatype integration for scanning Sonatype repositories for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: sonatypeIntegrationCreate,
		ReadContext:   sonatypeIntegrationRead,
		UpdateContext: sonatypeIntegrationUpdate,
		DeleteContext: sonatypeIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: SonatypeIntegrationResourceSchema(),
	}
}

func sonatypeIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	apiKey := writeonly.GetString(d, "api_key", "api_key_wo")

	input := &client.CreateSonatypeIntegrationInput{
		Name:   d.Get("name").(string),
		User:   d.Get("user").(string),
		ApiKey: apiKey,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	if orgURL := d.Get("org_url").(string); orgURL != "" {
		input.OrgURL = orgURL
	}

	if onpremID := d.Get("onprem_deployment_id").(string); onpremID != "" {
		input.OnpremDeploymentID = onpremID
	}

	resp, err := client.CreateSonatypeIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return sonatypeIntegrationRead(ctx, d, m)
}

func sonatypeIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle api_key rotation separately
	if d.HasChanges("api_key", "api_key_wo", "api_key_wo_version", "user") {
		apiKey := writeonly.GetString(d, "api_key", "api_key_wo")

		if apiKey != "" {
			replaceInput := &client.ReplaceSonatypeApiKeyInput{
				User:   d.Get("user").(string),
				ApiKey: apiKey,
			}
			if err := client.ReplaceSonatypeApiKey(ctx, c, d.Id(), replaceInput); err != nil {
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
	input := &client.UpdateSonatypeIntegrationInput{}
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
	if d.HasChange("org_url") {
		orgURL := d.Get("org_url").(string)
		input.OrgURL = &orgURL
		hasChanges = true
	}
	if d.HasChange("onprem_deployment_id") {
		onpremID := d.Get("onprem_deployment_id").(string)
		input.OnpremDeploymentID = &onpremID
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateSonatypeIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return sonatypeIntegrationRead(ctx, d, m)
}

func sonatypeIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteSonatypeIntegration(ctx, c, d.Id())
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
