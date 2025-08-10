package deployment

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Deployment resource for managing Hush Security deployments"

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: deploymentCreate,
		ReadContext:   deploymentRead,
		UpdateContext: deploymentUpdate,
		DeleteContext: deploymentDelete,
		Schema:        DeploymentResourceSchema(),
	}
}

func deploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateDeploymentInput{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		EnvType:     d.Get("env_type").(string),
		Kind:        d.Get("kind").(string),
	}

	resp, err := client.CreateDeploymentWithCredentials(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Created deployment", map[string]interface{}{
		"deployment_id": resp.ID,
		"name":          resp.Name,
	})

	if resp.ID == "" {
		return diag.Errorf("API returned empty ID for new deployment")
	}

	d.SetId(resp.ID)

	if diags := setCredentialFields(d, resp); diags.HasError() {
		return diags
	}

	if diags := setDeploymentFields(d, &resp.Deployment); diags.HasError() {
		return diags
	}

	return nil
}

func deploymentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateDeploymentInput{}
	hasChanges := false

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
	if d.HasChange("env_type") {
		envType := d.Get("env_type").(string)
		input.EnvType = &envType
		hasChanges = true
	}
	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("kind") {
		kind := d.Get("kind").(string)
		input.Kind = &kind
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	updated, err := client.UpdateDeployment(ctx, c, d.Id(), input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update deployment: %w", err))
	}

	tflog.Debug(ctx, "Updated deployment", map[string]interface{}{
		"deployment_id": d.Id(),
	})

	if diags := setDeploymentFields(d, updated); diags.HasError() {
		return diags
	}

	return nil
}

func deploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteDeployment(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func setCredentialFields(d *schema.ResourceData, resp *client.DeploymentCredentialsResponse) diag.Diagnostics {
	credentials := map[string]interface{}{
		"token":             resp.Token,
		"password":          resp.Password,
		"image_pull_secret": resp.ImagePullSecret,
	}

	for field, value := range credentials {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}
