package gcp_sa_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage GCP SA access privileges in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: forceNewOnTypeSwitch,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func forceNewOnTypeSwitch(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Id() == "" {
		return nil
	}

	oldEmail, newEmail := d.GetChange("sa_email")
	oldConf, newConf := d.GetChange("sa_config")
	hadEmail := oldEmail.(string) != ""
	hasEmail := newEmail.(string) != ""
	hadConf := len(oldConf.([]any)) > 0
	hasConf := len(newConf.([]any)) > 0

	switching := hadEmail && !hasEmail && hasConf ||
		hadConf && !hasConf && hasEmail
	if switching {
		if err := d.ForceNew("sa_email"); err != nil {
			return err
		}
		if err := d.ForceNew("sa_config"); err != nil {
			return err
		}
	}

	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateGCPSAAccessPrivilegeInput{
		Name:      d.Get("name").(string),
		ProjectID: d.Get("project_id").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	if v, ok := d.GetOk("sa_email"); ok {
		input.SaEmail = v.(string)
	}

	if v, ok := d.GetOk("sa_config"); ok {
		input.SaConf = expandSaConf(v.([]any))
	}

	privilege, err := client.CreateGCPSAAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetGCPSAAccessPrivilege(ctx, c, id)
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
		"project_id":  privilege.ProjectID,
		"sa_email":    privilege.SaEmail,
		"sa_config":   flattenSaConf(privilege.SaConf),
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

	input := &client.UpdateGCPSAAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("project_id") {
		v := d.Get("project_id").(string)
		input.ProjectID = &v
	}
	if d.HasChange("sa_email") {
		v := d.Get("sa_email").(string)
		input.SaEmail = &v
	}
	if d.HasChange("sa_config") {
		if v, ok := d.GetOk("sa_config"); ok {
			input.SaConf = expandSaConf(v.([]any))
		}
	}

	_, err := client.UpdateGCPSAAccessPrivilege(ctx, c, id, input)
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

func expandSaConf(raw []any) *client.GCPSaConf {
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]any)
	conf := &client.GCPSaConf{
		DisplayName: m["display_name"].(string),
	}
	if v, ok := m["roles"]; ok {
		rolesList := v.([]any)
		roles := make([]string, len(rolesList))
		for i, r := range rolesList {
			roles[i] = r.(string)
		}
		conf.Roles = roles
	}
	return conf
}

func flattenSaConf(conf *client.GCPSaConf) []any {
	if conf == nil {
		return nil
	}
	return []any{
		map[string]any{
			"display_name": conf.DisplayName,
			"roles":        conf.Roles,
		},
	}
}
