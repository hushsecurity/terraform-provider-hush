package apigee_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Apigee access privileges in the Hush Security platform.",
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

	apiProducts := make([]string, 0)
	if v, ok := d.GetOk("api_products"); ok {
		for _, item := range v.([]any) {
			apiProducts = append(apiProducts, item.(string))
		}
	}

	input := &client.CreateApigeeAccessPrivilegeInput{
		Name:           d.Get("name").(string),
		DeveloperEmail: d.Get("developer_email").(string),
		ProjectID:      d.Get("project_id").(string),
		APIProducts:    apiProducts,
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	if v, ok := d.GetOk("app_name"); ok {
		s := v.(string)
		input.AppName = &s
	}

	if v, ok := d.GetOk("app_config"); ok {
		input.AppConfig = expandAppConfig(v.([]any))
	}

	privilege, err := client.CreateApigeeAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetApigeeAccessPrivilege(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(privilege.ID)

	fields := map[string]any{
		"name":            privilege.Name,
		"description":     privilege.Description,
		"developer_email": privilege.DeveloperEmail,
		"project_id":      privilege.ProjectID,
		"api_products":    privilege.APIProducts,
		"app_config":      flattenAppConfig(privilege.AppConfig),
		"type":            privilege.Type,
	}

	if privilege.AppName != nil {
		fields["app_name"] = *privilege.AppName
	} else {
		fields["app_name"] = ""
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

	input := &client.UpdateApigeeAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("developer_email") {
		v := d.Get("developer_email").(string)
		input.DeveloperEmail = &v
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		input.ProjectID = &v
	}
	if d.HasChange("api_products") {
		list := d.Get("api_products").([]any)
		products := make([]string, len(list))
		for i, item := range list {
			products[i] = item.(string)
		}
		input.APIProducts = &products
	}
	if d.HasChange("app_name") {
		v := d.Get("app_name").(string)
		input.AppName = &v
	}
	if d.HasChange("app_config") {
		input.AppConfig = expandAppConfig(d.Get("app_config").([]any))
	}

	_, err := client.UpdateApigeeAccessPrivilege(ctx, c, id, input)
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
