package github_app_access_privilege

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage GitHub App access privileges in the Hush Security platform: the repositories and permissions installation tokens are narrowed to.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: customizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

// getRepositories always returns a list: a nil slice would marshal to null,
// which the API rejects.
func getRepositories(d *schema.ResourceData) []string {
	list := d.Get("repositories").([]any)
	repos := make([]string, len(list))
	for i, item := range list {
		repos[i] = item.(string)
	}
	return repos
}

func getPermissions(d *schema.ResourceData) map[string]string {
	raw := d.Get("permissions").(map[string]any)
	perms := make(map[string]string, len(raw))
	for k, v := range raw {
		perms[k] = v.(string)
	}
	return perms
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	input := &client.CreateGithubAppAccessPrivilegeInput{
		Name:         d.Get("name").(string),
		Repositories: getRepositories(d),
		Permissions:  getPermissions(d),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}

	privilege, err := client.CreateGithubAppAccessPrivilege(ctx, c, input)
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

	privilege, err := client.GetGithubAppAccessPrivilege(ctx, c, id)
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
		"type":        privilege.Type,
		"permissions": privilege.Permissions,
	}

	if privilege.Repositories != nil {
		fields["repositories"] = privilege.Repositories
	} else {
		fields["repositories"] = []string{}
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

	input := &client.UpdateGithubAppAccessPrivilegeInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("repositories") {
		repos := getRepositories(d)
		input.Repositories = &repos
	}
	if d.HasChange("permissions") {
		perms := getPermissions(d)
		input.Permissions = &perms
	}

	_, err := client.UpdateGithubAppAccessPrivilege(ctx, c, id, input)
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
