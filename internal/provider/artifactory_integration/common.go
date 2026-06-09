package artifactory_integration

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc                 = "The unique identifier of the Artifactory integration"
	nameDesc               = "The name of the Artifactory integration"
	descriptionDesc        = "The description of the Artifactory integration"
	orgURLDesc             = "The Artifactory organization URL (e.g., https://mycompany.jfrog.io)"
	tokenDesc              = "The access token for Artifactory authentication"
	tokenWODesc            = "The access token for Artifactory authentication (write-only). This is more secure than `token` because Terraform will not store this value in the state file. Either `token` or `token_wo` must be specified."
	tokenWOVersionDesc     = "Used to trigger updates for `token_wo`. This value should be changed when the token changes. Can be any value (e.g., a timestamp, version number, or hash)."
	onpremDeploymentIDDesc = "The ID of the on-premises deployment to associate with this integration"
	statusDesc             = "The current status of the integration"
	statusMessageDesc      = "The status message providing additional details about the integration status"
	webhookProvisionedDesc = "Whether the webhook has been provisioned for this integration"
)

func ArtifactoryIntegrationResourceSchema() map[string]*schema.Schema {
	s := ArtifactoryIntegrationDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 60),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 200),
	}
	s["org_url"] = &schema.Schema{
		Description:  orgURLDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
	}
	s["token"] = &schema.Schema{
		Description:   tokenDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"token_wo"},
		ExactlyOneOf:  []string{"token", "token_wo"},
	}
	s["token_wo"] = &schema.Schema{
		Description:   tokenWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"token"},
		ExactlyOneOf:  []string{"token", "token_wo"},
		RequiredWith:  []string{"token_wo_version"},
	}
	s["token_wo_version"] = &schema.Schema{
		Description:  tokenWOVersionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"token_wo"},
	}
	s["onprem_deployment_id"] = &schema.Schema{
		Description:  onpremDeploymentIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "onprem_deployment_id must start with 'dep-'"),
	}
	return s
}

func ArtifactoryIntegrationDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description:   idDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"id"},
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"org_url": {
			Description: orgURLDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"onprem_deployment_id": {
			Description: onpremDeploymentIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status_message": {
			Description: statusMessageDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"webhook_provisioned": {
			Description: webhookProvisionedDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}
}

func artifactoryIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.ArtifactoryIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetArtifactoryIntegration(ctx, c, id)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	} else if id, exists := d.GetOk("id"); exists {
		integrationID := id.(string)
		integration, err = client.GetArtifactoryIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no Artifactory integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetArtifactoryIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup Artifactory integration by name '%s': %w", integrationName, lookupErr))
		}

		switch len(integrations) {
		case 0:
			return diag.Errorf("no Artifactory integration found with name: %s", integrationName)
		case 1:
			integration, err = client.GetArtifactoryIntegration(ctx, c, integrations[0].ID)
			if err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("multiple Artifactory integrations found with name '%s'. Use the integration ID instead for exact matching", integrationName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(integration.ID)
	}

	if diags := setArtifactoryIntegrationFields(d, integration); diags.HasError() {
		return diags
	}

	return nil
}

func setArtifactoryIntegrationFields(d *schema.ResourceData, integration *client.ArtifactoryIntegration) diag.Diagnostics {
	fields := map[string]any{
		"name":                 integration.Name,
		"description":          integration.Description,
		"org_url":              integration.OrgURL,
		"onprem_deployment_id": integration.OnpremDeploymentID,
		"status":               integration.Status,
		"status_message":       integration.StatusMessage,
		"webhook_provisioned":  integration.WebhookProvisioned,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}
