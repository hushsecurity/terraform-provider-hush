package deployment

import (
	"context"
	"fmt"
	"net/http"

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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: DeploymentResourceSchema(),
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

	d.SetId(resp.ID)

	// Set computed sensitive fields
	if err := d.Set("token", resp.Token); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set token: %w", err))
	}
	if err := d.Set("password", resp.Password); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set password: %w", err))
	}
	if err := d.Set("image_pull_secret", resp.ImagePullSecret); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set image_pull_secret: %w", err))
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

	_, err := client.UpdateDeployment(ctx, c, d.Id(), input)
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func deploymentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteDeployment(ctx, c, d.Id())
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
