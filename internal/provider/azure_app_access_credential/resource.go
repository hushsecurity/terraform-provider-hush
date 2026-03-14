package azure_app_access_credential

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Azure app dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: validateCredentialPairing,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func validateCredentialPairing(_ context.Context, d *schema.ResourceDiff, _ any) error {
	_, hasClientID := d.GetOk("client_id")
	_, hasSecret := d.GetOk("client_secret")
	_, hasSecretWO := d.GetOk("client_secret_wo")

	if hasClientID && !hasSecret && !hasSecretWO && d.Id() == "" {
		return fmt.Errorf("client_id requires one of client_secret or client_secret_wo")
	}
	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	clientSecret := getClientSecret(d)

	input := &client.CreateAzureAppAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		TenantID:      d.Get("tenant_id").(string),
		ClientID:      d.Get("client_id").(string),
		ClientSecret:  clientSecret,
	}

	credential, err := client.CreateAzureAppAccessCredential(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	return resourceRead(ctx, d, meta)
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	id := d.Id()
	if id == "" {
		if v, ok := d.GetOk("id"); ok {
			id = v.(string)
		}
	}

	if id == "" {
		return diag.Errorf("id is required")
	}

	credential, err := client.GetAzureAppAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":           credential.Name,
		"description":    credential.Description,
		"deployment_ids": credential.DeploymentIDs,
		"tenant_id":      credential.TenantID,
		"client_id":      credential.ClientID,
		"type":           string(credential.Type),
		"kind":           credential.Kind,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	input := &client.UpdateAzureAppAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("tenant_id") {
		v := d.Get("tenant_id").(string)
		input.TenantID = &v
	}
	if d.HasChange("client_id") {
		v := d.Get("client_id").(string)
		input.ClientID = &v
	}
	if d.HasChange("client_secret") || d.HasChange("client_secret_wo") || d.HasChange("client_secret_wo_version") {
		clientSecret := getClientSecret(d)
		input.ClientSecret = &clientSecret
	}

	_, err := client.UpdateAzureAppAccessCredential(ctx, c, id, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, meta)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	err := client.DeleteAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// getClientSecret reads the client secret from either the regular
// attribute or the write-only attribute via GetRawConfig.
func getClientSecret(d *schema.ResourceData) string {
	if v, ok := d.GetOk("client_secret"); ok {
		return v.(string)
	}
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return ""
	}
	v := rawConfig.GetAttr("client_secret_wo")
	if v.IsNull() || !v.IsKnown() {
		return ""
	}
	return v.AsString()
}
