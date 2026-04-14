package salesforce_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Salesforce access privileges in the Hush Security platform.",
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

	input := &client.CreateSalesforceAccessPrivilegeInput{
		Name:      d.Get("name").(string),
		RunAsUser: d.Get("run_as_user").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	if v, ok := d.GetOk("scopes"); ok {
		list := v.([]any)
		scopes := make([]string, len(list))
		for i, item := range list {
			scopes[i] = item.(string)
		}
		input.Scopes = scopes
	}

	privilege, err := client.CreateSalesforceAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetSalesforceAccessPrivilege(ctx, c, id)
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
		"run_as_user": privilege.RunAsUser,
		"type":        privilege.Type,
	}

	if privilege.Scopes != nil {
		fields["scopes"] = privilege.Scopes
	} else {
		fields["scopes"] = []string{}
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

	input := &client.UpdateSalesforceAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("run_as_user") {
		v := d.Get("run_as_user").(string)
		input.RunAsUser = &v
	}
	if d.HasChange("scopes") {
		list := d.Get("scopes").([]any)
		scopes := make([]string, len(list))
		for i, item := range list {
			scopes[i] = item.(string)
		}
		input.Scopes = &scopes
	}

	_, err := client.UpdateSalesforceAccessPrivilege(ctx, c, id, input)
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
