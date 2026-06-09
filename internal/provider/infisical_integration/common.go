package infisical_integration

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
	idDesc                    = "The unique identifier of the Infisical integration"
	nameDesc                  = "The name of the Infisical integration"
	descriptionDesc           = "The description of the Infisical integration"
	baseURLDesc               = "The Infisical base URL (e.g., https://app.infisical.com)"
	clientIDDesc              = "The client ID for Infisical authentication"
	clientSecretDesc          = "The client secret for Infisical authentication"
	clientSecretWODesc        = "The client secret for Infisical authentication (write-only). This is more secure than `client_secret` because Terraform will not store this value in the state file. Either `client_secret` or `client_secret_wo` must be specified."
	clientSecretWOVersionDesc = "Used to trigger updates for `client_secret_wo`. This value should be changed when the client secret changes. Can be any value (e.g., a timestamp, version number, or hash)."
	onpremDeploymentIDDesc    = "The ID of the on-premises deployment to associate with this integration"
	statusDesc                = "The current status of the integration"
	statusMessageDesc         = "The status message providing additional details about the integration status"
)

func InfisicalIntegrationResourceSchema() map[string]*schema.Schema {
	s := InfisicalIntegrationDataSourceSchema()

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
	s["base_url"] = &schema.Schema{
		Description:  baseURLDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
	}
	s["client_id"] = &schema.Schema{
		Description: clientIDDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["client_secret"] = &schema.Schema{
		Description:   clientSecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"client_secret_wo"},
		ExactlyOneOf:  []string{"client_secret", "client_secret_wo"},
	}
	s["client_secret_wo"] = &schema.Schema{
		Description:   clientSecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"client_secret"},
		ExactlyOneOf:  []string{"client_secret", "client_secret_wo"},
		RequiredWith:  []string{"client_secret_wo_version"},
	}
	s["client_secret_wo_version"] = &schema.Schema{
		Description:  clientSecretWOVersionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"client_secret_wo"},
	}
	s["onprem_deployment_id"] = &schema.Schema{
		Description:  onpremDeploymentIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "onprem_deployment_id must start with 'dep-'"),
	}
	return s
}

func InfisicalIntegrationDataSourceSchema() map[string]*schema.Schema {
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
		"base_url": {
			Description: baseURLDesc,
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
	}
}

func infisicalIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.InfisicalIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetInfisicalIntegration(ctx, c, id)
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
		integration, err = client.GetInfisicalIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no Infisical integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetInfisicalIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup Infisical integration by name '%s': %w", integrationName, lookupErr))
		}

		switch len(integrations) {
		case 0:
			return diag.Errorf("no Infisical integration found with name: %s", integrationName)
		case 1:
			integration, err = client.GetInfisicalIntegration(ctx, c, integrations[0].ID)
			if err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("multiple Infisical integrations found with name '%s'. Use the integration ID instead for exact matching", integrationName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(integration.ID)
	}

	if diags := setInfisicalIntegrationFields(d, integration); diags.HasError() {
		return diags
	}

	return nil
}

func setInfisicalIntegrationFields(d *schema.ResourceData, integration *client.InfisicalIntegration) diag.Diagnostics {
	fields := map[string]any{
		"name":                 integration.Name,
		"description":          integration.Description,
		"base_url":             integration.BaseURL,
		"onprem_deployment_id": integration.OnpremDeploymentID,
		"status":               integration.Status,
		"status_message":       integration.StatusMessage,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}
