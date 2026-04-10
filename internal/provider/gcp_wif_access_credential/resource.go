package gcp_wif_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage GCP WIF (Workload Identity Federation) access credentials in the Hush Security platform.",
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

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateGcpWifAccessCredentialInput{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		DeploymentIDs:      deploymentIDs,
		ProjectNumber:      d.Get("project_number").(string),
		PoolID:             d.Get("pool_id").(string),
		WorkloadProviderID: d.Get("workload_provider_id").(string),
	}

	if v, ok := d.GetOk("audience"); ok {
		input.Audience = v.(string)
	}

	credential, err := client.CreateGcpWifAccessCredential(ctx, c, input)
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

	credential, err := client.GetGcpWifAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":                 credential.Name,
		"description":          credential.Description,
		"deployment_ids":       credential.DeploymentIDs,
		"project_number":       credential.ProjectNumber,
		"pool_id":              credential.PoolID,
		"workload_provider_id": credential.WorkloadProviderID,
		"audience":             credential.Audience,
		"issuer_url":           credential.IssuerURL,
		"type":                 string(credential.Type),
		"kind":                 credential.Kind,
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

	input := &client.UpdateGcpWifAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("deployment_ids") {
		deploymentIDs := make([]string, 0)
		if v, ok := d.GetOk("deployment_ids"); ok {
			for _, item := range v.([]any) {
				deploymentIDs = append(deploymentIDs, item.(string))
			}
		}
		input.DeploymentIDs = deploymentIDs
	}
	if d.HasChange("project_number") {
		v := d.Get("project_number").(string)
		input.ProjectNumber = &v
	}
	if d.HasChange("pool_id") {
		v := d.Get("pool_id").(string)
		input.PoolID = &v
	}
	if d.HasChange("workload_provider_id") {
		v := d.Get("workload_provider_id").(string)
		input.WorkloadProviderID = &v
	}
	if d.HasChange("audience") {
		v := d.Get("audience").(string)
		input.Audience = &v
	}

	_, err := client.UpdateGcpWifAccessCredential(ctx, c, id, input)
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
