package gcp_integration

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
	idDesc                    = "The unique identifier of the GCP integration"
	nameDesc                  = "The name of the GCP integration"
	descriptionDesc           = "The description of the GCP integration"
	onpremDeploymentIDDesc    = "The ID of the on-premises deployment to associate with this integration"
	serviceAccountEmailDesc   = "The GCP service account email from the onboarding script. Set this after running the onboarding script to complete the integration setup."
	statusDesc                = "The current status of the integration"
	statusMessageDesc         = "Additional details about the integration status"
	statusAtDesc              = "The timestamp of the last status change"
	typeDesc                  = "The type of integration (always 'gcp' for this resource)"
	createdAtDesc             = "The timestamp when the integration was created"
	modifiedAtDesc            = "The timestamp when the integration was last modified"
	nextRescanAtDesc          = "The timestamp of the next scheduled rescan"
	nextFullScanAtDesc        = "The timestamp of the next full scan"
	nextPeriodicChecksAtDesc  = "The timestamp of the next periodic checks"
	nextUpdateResourcesAtDesc = "The timestamp of the next resource update"
	projectsDesc              = "List of GCP projects to scan"
	projectIDDesc             = "The GCP project ID"
	projectEnabledDesc        = "Whether scanning is enabled for this project"
	projectDisplayNameDesc    = "The display name of the GCP project"
	projectStateDesc          = "The current state of the project within the integration"
	projectStateMessageDesc   = "Additional details about the project state"
	projectOrganizationIDDesc = "The GCP organization ID associated with the project"
	featuresDesc              = "List of GCP features to enable"
	featureNameDesc           = "The feature name (secret_manager, gcs_tf_state, artifact_registry, iam)"
	featureEnabledDesc        = "Whether this feature is enabled"
	featureStateDesc          = "The current state of the feature"
	featureStateMessageDesc   = "Additional details about the feature state"
)

var gcpFeatureNames = []string{"secret_manager", "gcs_tf_state", "artifact_registry", "iam"}

func GCPIntegrationResourceSchema() map[string]*schema.Schema {
	s := GCPIntegrationDataSourceSchema()

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
	s["onprem_deployment_id"] = &schema.Schema{
		Description:  onpremDeploymentIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "onprem_deployment_id must start with 'dep-'"),
	}
	s["service_account_email"] = &schema.Schema{
		Description: serviceAccountEmailDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	}
	s["projects"] = &schema.Schema{
		Description: projectsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"project_id": {
					Description: projectIDDesc,
					Type:        schema.TypeString,
					Required:    true,
				},
				"enabled": {
					Description: projectEnabledDesc,
					Type:        schema.TypeBool,
					Required:    true,
				},
				"display_name": {
					Description: projectDisplayNameDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
				"state": {
					Description: projectStateDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
				"state_message": {
					Description: projectStateMessageDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
				"organization_id": {
					Description: projectOrganizationIDDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}
	s["features"] = &schema.Schema{
		Description: featuresDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description:  featureNameDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(gcpFeatureNames, false),
				},
				"enabled": {
					Description: featureEnabledDesc,
					Type:        schema.TypeBool,
					Required:    true,
				},
				"state": {
					Description: featureStateDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
				"state_message": {
					Description: featureStateMessageDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}

	return s
}

func GCPIntegrationDataSourceSchema() map[string]*schema.Schema {
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
		"onprem_deployment_id": {
			Description: onpremDeploymentIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_account_email": {
			Description: serviceAccountEmailDesc,
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
		"status_at": {
			Description: statusAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"created_at": {
			Description: createdAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"modified_at": {
			Description: modifiedAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"next_rescan_at": {
			Description: nextRescanAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"next_full_scan_at": {
			Description: nextFullScanAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"next_periodic_checks_at": {
			Description: nextPeriodicChecksAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"next_update_resources_at": {
			Description: nextUpdateResourcesAtDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"projects": {
			Description: projectsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"project_id": {
						Description: projectIDDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"enabled": {
						Description: projectEnabledDesc,
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"display_name": {
						Description: projectDisplayNameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state": {
						Description: projectStateDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state_message": {
						Description: projectStateMessageDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"organization_id": {
						Description: projectOrganizationIDDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"features": {
			Description: featuresDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: featureNameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"enabled": {
						Description: featureEnabledDesc,
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"state": {
						Description: featureStateDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state_message": {
						Description: featureStateMessageDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
	}
}

func gcpIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.GCPIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetGCPIntegration(ctx, c, id)
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
		integration, err = client.GetGCPIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no GCP integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetGCPIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup GCP integration by name '%s': %w", integrationName, lookupErr))
		}
		if len(integrations) == 0 {
			return diag.Errorf("no GCP integration found with name: %s", integrationName)
		}
		if len(integrations) > 1 {
			return diag.Errorf("multiple GCP integrations found with name: %s, please use id instead", integrationName)
		}
		// Get full details
		integration, err = client.GetGCPIntegration(ctx, c, integrations[0].ID)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		return diag.Errorf("one of `id` or `name` must be specified")
	}

	d.SetId(integration.ID)

	fields := map[string]any{
		"name":                     integration.Name,
		"description":              integration.Description,
		"onprem_deployment_id":     integration.OnpremDeploymentID,
		"service_account_email":    integration.ServiceAccountEmail,
		"status":                   integration.Status,
		"status_message":           integration.StatusMessage,
		"status_at":                integration.StatusAt,
		"type":                     integration.Type,
		"created_at":               integration.CreatedAt,
		"modified_at":              integration.ModifiedAt,
		"next_rescan_at":           integration.NextRescanAt,
		"next_full_scan_at":        integration.NextFullScanAt,
		"next_periodic_checks_at":  integration.NextPeriodicChecksAt,
		"next_update_resources_at": integration.NextUpdateResourcesAt,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if integration.Projects != nil {
		projects := make([]map[string]any, len(integration.Projects))
		for i, p := range integration.Projects {
			projects[i] = map[string]any{
				"project_id":      p.ProjectID,
				"enabled":         p.Enabled,
				"display_name":    p.DisplayName,
				"state":           p.State,
				"state_message":   p.StateMessage,
				"organization_id": p.OrganizationID,
			}
		}
		if err := d.Set("projects", projects); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set projects: %w", err))
		}
	}

	if integration.Features != nil {
		features := make([]map[string]any, len(integration.Features))
		for i, f := range integration.Features {
			features[i] = map[string]any{
				"name":          f.Name,
				"enabled":       f.Enabled,
				"state":         f.State,
				"state_message": f.StateMessage,
			}
		}
		if err := d.Set("features", features); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set features: %w", err))
		}
	}

	return nil
}
