package deployment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc              = "The unique identifier of the deployment"
	nameDesc            = "The name of the deployment"
	descriptionDesc     = "The description of the deployment"
	envTypeDesc         = "The environment type for the deployment (dev, prod)"
	kindDesc            = "The deployment kind (k8s, ecs, serverless)"
	statusDesc          = "The current status of the deployment"
	tokenDesc           = "The deployment token for authentication"
	passwordDesc        = "The deployment password for authentication"
	imagePullSecretDesc = "The image pull secret for accessing private container images"
)

func DeploymentResourceSchema() map[string]*schema.Schema {
	s := DeploymentDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description: nameDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["description"] = &schema.Schema{
		Description: descriptionDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["env_type"] = &schema.Schema{
		Description: envTypeDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "dev",
		ValidateFunc: validation.StringInSlice([]string{
			"dev",
			"prod",
		}, false),
	}
	s["kind"] = &schema.Schema{
		Description: kindDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringInSlice([]string{
			"k8s",
			"ecs",
			"serverless",
		}, false),
	}

	s["token"] = &schema.Schema{
		Description: tokenDesc,
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}
	s["password"] = &schema.Schema{
		Description: passwordDesc,
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}
	s["image_pull_secret"] = &schema.Schema{
		Description: imagePullSecretDesc,
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}

	return s
}

func DeploymentDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: idDesc,
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: nameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"env_type": {
			Description: envTypeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

// Helper Functions

func deploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	var deployment *client.Deployment
	var err error
	var deploymentID string

	if id := d.Id(); id != "" {
		deploymentID = id
	} else if id, exists := d.GetOk("id"); exists {
		deploymentID = id.(string)
	} else {
		return diag.Errorf("deployment ID is required")
	}

	deployment, err = client.GetDeployment(ctx, c, deploymentID)
	if err != nil {
		// Handle 404 errors gracefully by removing from state
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		} else {
			return diag.FromErr(err)
		}
	}

	if deployment == nil {
		d.SetId("")
		return nil
	}

	if d.Id() == "" {
		d.SetId(deployment.ID)
	}

	if diags := setDeploymentFields(d, deployment); diags.HasError() {
		return diags
	}

	return nil
}

func setDeploymentFields(d *schema.ResourceData, deployment *client.Deployment) diag.Diagnostics {
	fields := map[string]interface{}{
		"name":        deployment.Name,
		"description": deployment.Description,
		"env_type":    deployment.EnvType,
		"kind":        deployment.Kind,
		"status":      deployment.Status,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}
