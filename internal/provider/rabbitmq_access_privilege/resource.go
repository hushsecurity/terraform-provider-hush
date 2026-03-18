package rabbitmq_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage RabbitMQ access privileges in the Hush Security platform.",
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

	input := &client.CreateRabbitmqAccessPrivilegeInput{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	input.Permissions = expandPermissions(d.Get("permissions").([]any))

	if v, ok := d.GetOk("tags"); ok {
		tagsList := v.([]any)
		tags := make([]string, len(tagsList))
		for i, item := range tagsList {
			tags[i] = item.(string)
		}
		input.Tags = tags
	}

	privilege, err := client.CreateRabbitmqAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetRabbitmqAccessPrivilege(ctx, c, id)
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
		"permissions": flattenPermissions(privilege.Permissions),
		"type":        privilege.Type,
	}

	if privilege.Tags != nil {
		fields["tags"] = privilege.Tags
	} else {
		fields["tags"] = []string{}
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

	input := &client.UpdateRabbitmqAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("permissions") {
		permissions := expandPermissions(d.Get("permissions").([]any))
		input.Permissions = &permissions
	}
	if d.HasChange("tags") {
		tagsList := d.Get("tags").([]any)
		tags := make([]string, len(tagsList))
		for i, item := range tagsList {
			tags[i] = item.(string)
		}
		input.Tags = &tags
	}

	_, err := client.UpdateRabbitmqAccessPrivilege(ctx, c, id, input)
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
