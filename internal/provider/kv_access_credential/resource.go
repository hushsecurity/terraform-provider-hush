package kv_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage key-value access credentials in the Hush Security platform. KV credentials store multiple key-value pairs that can be delivered as separate environment variables to specified deployments.",
		CreateContext: kvAccessCredentialCreate,
		ReadContext:   kvAccessCredentialRead,
		UpdateContext: kvAccessCredentialUpdate,
		DeleteContext: kvAccessCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: KVAccessCredentialResourceSchema(),
	}
}

func kvAccessCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		list := v.([]interface{})
		for _, item := range list {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	items := make([]client.KVItem, 0)
	if v, ok := d.GetOk("items"); ok {
		itemsList := v.([]interface{})
		for _, item := range itemsList {
			itemMap := item.(map[string]interface{})
			items = append(items, client.KVItem{
				Key:   itemMap["key"].(string),
				Value: itemMap["value"].(string),
			})
		}
	}

	input := &client.CreateKVAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		Items:         items,
	}

	credential, err := client.CreateKVAccessCredential(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	return kvAccessCredentialRead(ctx, d, meta)
}

func kvAccessCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	credential, err := client.GetKVAccessCredential(ctx, c, id)
	if err != nil {
		// Handle 404 by removing from state
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	if err := d.Set("name", credential.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", credential.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("deployment_ids", credential.DeploymentIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("keys", credential.Keys); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", string(credential.Type)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", credential.CreatedAt); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("modified_at", credential.ModifiedAt); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func kvAccessCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	input := &client.UpdateAccessCredentialInput{}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		input.Description = &description
	}

	_, err := client.UpdateKVAccessCredential(ctx, c, id, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return kvAccessCredentialRead(ctx, d, meta)
}

func kvAccessCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
