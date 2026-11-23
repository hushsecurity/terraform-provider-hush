package cloudflare_r2_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Cloudflare R2 access privileges in the Hush Security platform: the permission, buckets and jurisdiction of the R2 tokens issued with a `hush_cloudflare_r2_access_credential`.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: validateBuckets,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateCloudflareR2AccessPrivilegeInput{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Permission:   d.Get("permission").(string),
		Buckets:      requestBuckets(d),
		Jurisdiction: d.Get("jurisdiction").(string),
	}

	privilege, err := client.CreateCloudflareR2AccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetCloudflareR2AccessPrivilege(ctx, c, id)
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
		"permission":  privilege.Permission,
		"buckets":     privilege.Buckets,
		// null buckets are every bucket, except on an admin permission
		"all_buckets":  !adminPermissions[privilege.Permission] && privilege.Buckets == nil,
		"jurisdiction": privilege.Jurisdiction,
		"type":         privilege.Type,
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

	input := &client.UpdateCloudflareR2AccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("permission") {
		v := d.Get("permission").(string)
		input.Permission = &v
	}
	// Sent with any change to the grant, as the Hush API keeps the stored
	// buckets otherwise: an admin permission refuses a list, and a change from
	// admin to object needs the buckets in the same request.
	if d.HasChanges("permission", "buckets", "all_buckets") {
		input.Buckets = client.NewBucketsUpdate(requestBuckets(d))
	}
	if d.HasChange("jurisdiction") {
		v := d.Get("jurisdiction").(string)
		input.Jurisdiction = &v
	}

	_, err := client.UpdateCloudflareR2AccessPrivilege(ctx, c, id, input)
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
