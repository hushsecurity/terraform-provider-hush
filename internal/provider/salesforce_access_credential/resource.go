package salesforce_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Salesforce dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

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

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateSalesforceAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		InstanceURL:   d.Get("instance_url").(string),
		ClientID:      d.Get("client_id").(string),
		ClientSecret:  getClientSecret(d),
	}

	credential, err := client.CreateSalesforceAccessCredential(ctx, c, input)
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

	credential, err := client.GetSalesforceAccessCredential(ctx, c, id)
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
		"instance_url":   credential.InstanceURL,
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

	input := &client.UpdateSalesforceAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("instance_url") {
		v := d.Get("instance_url").(string)
		input.InstanceURL = &v
	}
	if d.HasChange("client_id") {
		v := d.Get("client_id").(string)
		input.ClientID = &v
	}
	if d.HasChange("client_secret") || d.HasChange("client_secret_wo") || d.HasChange("client_secret_wo_version") {
		v := getClientSecret(d)
		input.ClientSecret = &v
	}

	_, err := client.UpdateSalesforceAccessCredential(ctx, c, id, input)
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
