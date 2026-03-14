package azure_app_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Azure app access privileges in the Hush Security platform.",
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

	oldAppID, newAppID := d.GetChange("app_id")
	oldConf, newConf := d.GetChange("app_config")
	hadAppID := oldAppID.(string) != ""
	hasAppID := newAppID.(string) != ""
	hadConf := len(oldConf.([]any)) > 0
	hasConf := len(newConf.([]any)) > 0

	switching := hadAppID && !hasAppID && hasConf ||
		hadConf && !hasConf && hasAppID
	if switching {
		if err := d.ForceNew("app_id"); err != nil {
			return err
		}
		if err := d.ForceNew("app_config"); err != nil {
			return err
		}
	}

	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateAzureAppAccessPrivilegeInput{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	if v, ok := d.GetOk("app_id"); ok {
		input.AppID = v.(string)
	}

	if v, ok := d.GetOk("app_config"); ok {
		input.AppConfig = expandAppConfig(v.([]any))
	}

	privilege, err := client.CreateAzureAppAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetAzureAppAccessPrivilege(ctx, c, id)
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
		"app_id":      privilege.AppID,
		"app_config":  flattenAppConfig(privilege.AppConfig),
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

	input := &client.UpdateAzureAppAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("app_id") {
		v := d.Get("app_id").(string)
		input.AppID = &v
	}
	if d.HasChange("app_config") {
		if v, ok := d.GetOk("app_config"); ok {
			input.AppConfig = expandAppConfig(v.([]any))
		}
	}

	_, err := client.UpdateAzureAppAccessPrivilege(ctx, c, id, input)
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

func expandAppConfig(raw []any) *client.AzureAppConfig {
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]any)
	config := &client.AzureAppConfig{
		DisplayName: m["display_name"].(string),
	}
	if v, ok := m["roles"]; ok {
		rolesList := v.([]any)
		roles := make([]client.AzureAppRole, len(rolesList))
		for i, r := range rolesList {
			rm := r.(map[string]any)
			roles[i] = client.AzureAppRole{
				Name:  rm["name"].(string),
				Scope: rm["scope"].(string),
			}
		}
		config.Roles = roles
	}
	return config
}

func flattenAppConfig(config *client.AzureAppConfig) []any {
	if config == nil {
		return nil
	}
	roles := make([]any, len(config.Roles))
	for i, r := range config.Roles {
		roles[i] = map[string]any{
			"name":  r.Name,
			"scope": r.Scope,
		}
	}
	return []any{
		map[string]any{
			"display_name": config.DisplayName,
			"roles":        roles,
		},
	}
}
