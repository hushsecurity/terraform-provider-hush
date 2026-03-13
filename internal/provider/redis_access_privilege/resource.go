package redis_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Redis access privileges in the Hush Security platform.",
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

	input := &client.CreateRedisAccessPrivilegeInput{
		Name:   d.Get("name").(string),
		Grants: expandGrants(d.Get("grants").([]any)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	keys := make([]string, 0)
	if v, ok := d.GetOk("keys"); ok {
		for _, item := range v.([]any) {
			keys = append(keys, item.(string))
		}
	}
	input.Keys = keys

	if v, ok := d.GetOk("channels"); ok {
		channels := make([]string, 0)
		for _, item := range v.([]any) {
			channels = append(channels, item.(string))
		}
		input.Channels = channels
	}

	privilege, err := client.CreateRedisAccessPrivilege(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(privilege.ID)

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

	privilege, err := client.GetRedisAccessPrivilege(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(privilege.ID)

	fields := map[string]any{
		"name":        privilege.Name,
		"description": privilege.Description,
		"grants":      flattenGrants(privilege.Grants),
		"keys":        privilege.Keys,
		"channels":    privilege.Channels,
		"type":        privilege.Type,
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

	input := &client.UpdateRedisAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("grants") {
		grants := expandGrants(d.Get("grants").([]any))
		input.Grants = &grants
	}
	if d.HasChange("keys") {
		list := d.Get("keys").([]any)
		keys := make([]string, len(list))
		for i, item := range list {
			keys[i] = item.(string)
		}
		input.Keys = &keys
	}
	if d.HasChange("channels") {
		list := d.Get("channels").([]any)
		channels := make([]string, len(list))
		for i, item := range list {
			channels[i] = item.(string)
		}
		input.Channels = &channels
	}

	_, err := client.UpdateRedisAccessPrivilege(ctx, c, id, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, meta)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	err := client.DeleteAccessPrivilege(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
