package statsig_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Statsig access privileges in the Hush Security platform.",
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

func stringSet(d *schema.ResourceData, key string) []string {
	set, ok := d.Get(key).(*schema.Set)
	if !ok {
		return nil
	}
	out := make([]string, 0, set.Len())
	for _, v := range set.List() {
		out = append(out, v.(string))
	}
	return out
}

func expandKey(d *schema.ResourceData) client.StatsigKey {
	return client.StatsigKey{
		KeyType:               d.Get("key_type").(string),
		Environments:          stringSet(d, "environments"),
		TargetAppID:           d.Get("target_app_id").(string),
		SecondaryTargetAppIDs: stringSet(d, "secondary_target_app_ids"),
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateStatsigAccessPrivilegeInput{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Key:         expandKey(d),
	}

	privilege, err := client.CreateStatsigAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetStatsigAccessPrivilege(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(privilege.ID)

	fields := map[string]any{
		"name":                     privilege.Name,
		"description":              privilege.Description,
		"type":                     privilege.Type,
		"key_type":                 privilege.KeyType,
		"environments":             privilege.Environments,
		"target_app_id":            privilege.TargetAppID,
		"secondary_target_app_ids": privilege.SecondaryTargetAppIDs,
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

	input := &client.UpdateStatsigAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChanges(
		"key_type", "environments", "target_app_id", "secondary_target_app_ids",
	) {
		key := expandKey(d)
		input.Key = &key
	}

	_, err := client.UpdateStatsigAccessPrivilege(ctx, c, id, input)
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
