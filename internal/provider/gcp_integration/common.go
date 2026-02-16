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
	idDesc                  = "The unique identifier of the GCP integration"
	nameDesc                = "The name of the GCP integration"
	descriptionDesc         = "The description of the GCP integration"
	statusDesc              = "The current status of the GCP integration (pending, ok)"
	typeDesc                = "The type of integration (always gcp)"
	serviceAccountEmailDesc = "The GCP service account email used for the integration. Providing this will complete the integration setup."
	onboardingScriptDesc    = "The generated onboarding shell script for setting up GCP IAM permissions"
)

var gcpProjectIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{4,28}[a-z0-9]$`)

func GCPIntegrationResourceSchema() map[string]*schema.Schema {
	s := GCPIntegrationDataSourceSchema()

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
	s["service_account_email"] = &schema.Schema{
		Description: serviceAccountEmailDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}

	s["project"] = &schema.Schema{
		Description: "GCP project to include in the integration",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"project_id": {
					Description: "The GCP project ID",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: validation.StringMatch(
						gcpProjectIDPattern,
						"must be a valid GCP project ID (6-30 lowercase letters, digits, or hyphens, starting with a letter and ending with a letter or digit)",
					),
				},
				"enabled": {
					Description: "Whether this project is enabled for scanning",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"display_name": {
					Description: "The display name of the GCP project (discovered after completion)",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"state": {
					Description: "The state of the project (requested, ACTIVE)",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"organization_id": {
					Description: "The GCP organization ID the project belongs to",
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}

	s["feature"] = &schema.Schema{
		Description: "Feature to enable for the GCP integration",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The feature name (iam, secret_manager, gcs_tf_state, artifact_registry)",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: validation.StringInSlice([]string{
						"iam",
						"secret_manager",
						"gcs_tf_state",
						"artifact_registry",
					}, false),
				},
				"enabled": {
					Description: "Whether this feature is enabled",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"state": {
					Description: "The state of the feature (ok, disabled, insufficient_permissions, error)",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"state_message": {
					Description: "Additional details about the feature state",
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
			ConflictsWith: []string{"id"},
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_account_email": {
			Description: serviceAccountEmailDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"project": {
			Description: "GCP projects in the integration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"project_id": {
						Description: "The GCP project ID",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"enabled": {
						Description: "Whether this project is enabled for scanning",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"display_name": {
						Description: "The display name of the GCP project",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state": {
						Description: "The state of the project",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"organization_id": {
						Description: "The GCP organization ID",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"feature": {
			Description: "Features in the GCP integration",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "The feature name",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"enabled": {
						Description: "Whether this feature is enabled",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"state": {
						Description: "The state of the feature",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"state_message": {
						Description: "Additional details about the feature state",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"onboarding_script": {
			Description: onboardingScriptDesc,
			Type:        schema.TypeString,
			Computed:    true,
			Sensitive:   true,
		},
	}
}

// Helper Functions

func gcpIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	var integ *client.GCPIntegration
	var err error

	if id := d.Id(); id != "" {
		integ, err = client.GetGCPIntegration(ctx, c, id)
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
	} else if id, exists := d.GetOk("id"); exists {
		// Lookup by ID provided in configuration
		integID := id.(string)
		integ, err = client.GetGCPIntegration(ctx, c, integID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no GCP integration found with ID: %s", integID)
			} else {
				return diag.FromErr(err)
			}
		}
	} else if name, exists := d.GetOk("name"); exists {
		// Lookup by name
		integName := name.(string)
		integrations, err := client.GetGCPIntegrationsByName(ctx, c, integName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup GCP integration by name '%s': %w", integName, err))
		}

		switch len(integrations) {
		case 0:
			return diag.Errorf("no GCP integration found with name: %s", integName)
		case 1:
			integ = &integrations[0]
		default:
			return diag.Errorf("multiple GCP integrations found with name '%s'. Integration names must be unique. Consider using the integration ID instead for exact matching", integName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(integ.ID)
	}

	// Fetch onboarding script
	script, err := client.GetGCPIntegrationOnboardingScript(ctx, c, integ.ID)
	if err != nil {
		// Non-fatal: log warning but don't fail the read
		// The script may not be available in all states
		errResponse, ok := err.(*client.APIError)
		if !ok || errResponse.StatusCode != http.StatusNotFound {
			return diag.FromErr(fmt.Errorf("failed to fetch onboarding script: %w", err))
		}
	} else {
		if err := d.Set("onboarding_script", script); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set onboarding_script: %w", err))
		}
	}

	if diags := setGCPIntegrationFields(d, integ); diags.HasError() {
		return diags
	}

	return nil
}

func setGCPIntegrationFields(d *schema.ResourceData, integ *client.GCPIntegration) diag.Diagnostics {
	fields := map[string]interface{}{
		"name":        integ.Name,
		"description": integ.Description,
		"status":      integ.Status,
		"type":        integ.Type,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	// Set service_account_email
	if integ.ServiceAccountEmail != nil {
		if err := d.Set("service_account_email", *integ.ServiceAccountEmail); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set service_account_email: %w", err))
		}
	}

	// Set projects
	if integ.Projects != nil {
		projects := make([]map[string]interface{}, len(integ.Projects))
		for i, p := range integ.Projects {
			project := map[string]interface{}{
				"project_id": p.ProjectID,
				"enabled":    p.Enabled,
				"state":      p.State,
			}
			if p.DisplayName != nil {
				project["display_name"] = *p.DisplayName
			} else {
				project["display_name"] = ""
			}
			if p.OrganizationID != nil {
				project["organization_id"] = *p.OrganizationID
			} else {
				project["organization_id"] = ""
			}
			projects[i] = project
		}
		if err := d.Set("project", projects); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set project: %w", err))
		}
	}

	// Set features
	if integ.Features != nil {
		features := make([]map[string]interface{}, len(integ.Features))
		for i, f := range integ.Features {
			feature := map[string]interface{}{
				"name":    f.Name,
				"enabled": f.Enabled,
				"state":   f.State,
			}
			if f.StateMessage != nil {
				feature["state_message"] = *f.StateMessage
			} else {
				feature["state_message"] = ""
			}
			features[i] = feature
		}
		if err := d.Set("feature", features); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set feature: %w", err))
		}
	}

	return nil
}

// extractProjectInputs extracts project block data from the schema into client input structs
func extractProjectInputs(d *schema.ResourceData) []client.GCPProjectInput {
	if v, ok := d.GetOk("project"); ok {
		projectList := v.([]interface{})
		if len(projectList) > 0 {
			projects := make([]client.GCPProjectInput, len(projectList))
			for i, p := range projectList {
				projectMap := p.(map[string]interface{})
				projects[i] = client.GCPProjectInput{
					ProjectID: projectMap["project_id"].(string),
					Enabled:   projectMap["enabled"].(bool),
				}
			}
			return projects
		}
	}
	return nil
}

// extractFeatureInputs extracts feature block data from the schema into client input structs
func extractFeatureInputs(d *schema.ResourceData) []client.GCPFeatureInput {
	if v, ok := d.GetOk("feature"); ok {
		featureList := v.([]interface{})
		if len(featureList) > 0 {
			features := make([]client.GCPFeatureInput, len(featureList))
			for i, f := range featureList {
				featureMap := f.(map[string]interface{})
				features[i] = client.GCPFeatureInput{
					Name:    featureMap["name"].(string),
					Enabled: featureMap["enabled"].(bool),
				}
			}
			return features
		}
	}
	return nil
}
