package cloudinary_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Cloudinary access privileges in the Hush Security platform: the product environment and the role of the keys issued with a `hush_cloudinary_access_credential`.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: validateGrant,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func expandGrant(d *schema.ResourceData) client.CloudinaryGrant {
	return client.CloudinaryGrant{
		CloudName:       d.Get("cloud_name").(string),
		RoleID:          d.Get("role_id").(string),
		SystemPolicyIDs: requestPolicies(d),
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateCloudinaryAccessPrivilegeInput{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Grant:       expandGrant(d),
	}

	privilege, err := client.CreateCloudinaryAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetCloudinaryAccessPrivilege(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(privilege.ID)

	fields := map[string]any{
		"name":              privilege.Name,
		"description":       privilege.Description,
		"cloud_name":        privilege.CloudName,
		"role_id":           privilege.RoleID,
		"system_policy_ids": privilege.SystemPolicyIDs,
		"type":              privilege.Type,
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

	input := &client.UpdateCloudinaryAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	// Every grant field goes with any change to one: the Hush API takes a
	// switch between a role and policies only with the other cleared.
	if d.HasChanges("cloud_name", "role_id", "system_policy_ids") {
		grant := expandGrant(d)
		input.Grant = &grant
	}

	_, err := client.UpdateCloudinaryAccessPrivilege(ctx, c, id, input)
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
