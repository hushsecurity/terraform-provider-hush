package gcp_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Manages a Hush Security GCP integration for scanning GCP projects for secrets and sensitive data.\n\n" +
	"Set `service_account_email` to complete the integration with the GCP service account created by the onboarding module."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: gcpIntegrationCreate,
		ReadContext:   gcpIntegrationResourceRead,
		UpdateContext: gcpIntegrationUpdate,
		DeleteContext: gcpIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: GCPIntegrationResourceSchema(),
	}
}

func gcpIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateGCPIntegrationInput{
		Name: d.Get("name").(string),
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	if v, ok := d.GetOk("projects"); ok {
		input.Projects = expandGCPProjects(v.([]any))
	}
	if v, ok := d.GetOk("features"); ok {
		input.Features = expandGCPFeatures(v.([]any))
	}

	resp, err := client.CreateGCPIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	// If service_account_email is provided, complete the integration
	if saEmail := d.Get("service_account_email").(string); saEmail != "" {
		completeInput := &client.CompleteGCPIntegrationInput{
			ServiceAccountEmail: saEmail,
		}
		_, err := client.CompleteGCPIntegration(ctx, c, d.Id(), completeInput)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return gcpIntegrationResourceRead(ctx, d, m)
}

func gcpIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle completion if service_account_email is newly set
	if d.HasChange("service_account_email") {
		old, new := d.GetChange("service_account_email")
		if old.(string) == "" && new.(string) != "" {
			completeInput := &client.CompleteGCPIntegrationInput{
				ServiceAccountEmail: new.(string),
			}
			_, err := client.CompleteGCPIntegration(ctx, c, d.Id(), completeInput)
			if err != nil {
				errResponse, ok := err.(*client.APIError)
				if ok && errResponse.StatusCode == http.StatusNotFound {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}
		}
	}

	// Handle metadata and nested block updates
	input := &client.UpdateGCPIntegrationInput{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
	if d.HasChange("projects") {
		if v, ok := d.GetOk("projects"); ok {
			input.Projects = expandGCPProjects(v.([]any))
		} else {
			input.Projects = []client.GCPProjectInput{}
		}
		hasChanges = true
	}
	if d.HasChange("features") {
		if v, ok := d.GetOk("features"); ok {
			input.Features = expandGCPFeatures(v.([]any))
		} else {
			input.Features = []client.GCPFeatureInput{}
		}
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateGCPIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return gcpIntegrationResourceRead(ctx, d, m)
}

func gcpIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteGCPIntegration(ctx, c, d.Id())
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

// gcpIntegrationResourceRead reads the integration and filters features/projects
// to only include those the user configured, preventing drift from API defaults.
func gcpIntegrationResourceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	// Capture configured feature/project names BEFORE the read overwrites state
	configuredFeatureNames := getConfiguredFeatureNames(d)
	configuredProjectIDs := getConfiguredProjectIDs(d)

	diags := gcpIntegrationRead(ctx, d, m)
	if diags.HasError() || d.Id() == "" {
		return diags
	}

	// Filter features to only those the user configured (API returns all, including disabled defaults)
	if len(configuredFeatureNames) > 0 {
		allFeatures := d.Get("features").([]any)
		filtered := make([]map[string]any, 0, len(configuredFeatureNames))

		// Preserve config ordering
		for _, name := range configuredFeatureNames {
			for _, af := range allFeatures {
				afMap := af.(map[string]any)
				if afMap["name"].(string) == name {
					filtered = append(filtered, afMap)
					break
				}
			}
		}

		if len(filtered) > 0 {
			d.Set("features", filtered) //nolint:errcheck // already validated in gcpIntegrationRead
		}
	}

	// Filter projects to only those the user configured
	if len(configuredProjectIDs) > 0 {
		allProjects := d.Get("projects").([]any)
		filtered := make([]map[string]any, 0, len(configuredProjectIDs))

		// Preserve config ordering
		for _, pid := range configuredProjectIDs {
			for _, ap := range allProjects {
				apMap := ap.(map[string]any)
				if apMap["project_id"].(string) == pid {
					filtered = append(filtered, apMap)
					break
				}
			}
		}

		if len(filtered) > 0 {
			d.Set("projects", filtered) //nolint:errcheck // already validated in gcpIntegrationRead
		}
	}

	return diags
}

// getConfiguredFeatureNames returns the feature names from the current state/config
// before the read function overwrites them. On import (no prior state), returns nil.
func getConfiguredFeatureNames(d *schema.ResourceData) []string {
	raw := d.Get("features").([]any)
	if len(raw) == 0 {
		return nil
	}
	names := make([]string, 0, len(raw))
	for _, f := range raw {
		m := f.(map[string]any)
		if name, ok := m["name"].(string); ok && name != "" {
			names = append(names, name)
		}
	}
	return names
}

// getConfiguredProjectIDs returns the project IDs from the current state/config
// before the read function overwrites them. On import (no prior state), returns nil.
func getConfiguredProjectIDs(d *schema.ResourceData) []string {
	raw := d.Get("projects").([]any)
	if len(raw) == 0 {
		return nil
	}
	ids := make([]string, 0, len(raw))
	for _, p := range raw {
		m := p.(map[string]any)
		if pid, ok := m["project_id"].(string); ok && pid != "" {
			ids = append(ids, pid)
		}
	}
	return ids
}

func expandGCPProjects(raw []any) []client.GCPProjectInput {
	projects := make([]client.GCPProjectInput, len(raw))
	for i, v := range raw {
		m := v.(map[string]any)
		projects[i] = client.GCPProjectInput{
			ProjectID: m["project_id"].(string),
			Enabled:   m["enabled"].(bool),
		}
	}
	return projects
}

func expandGCPFeatures(raw []any) []client.GCPFeatureInput {
	features := make([]client.GCPFeatureInput, len(raw))
	for i, v := range raw {
		m := v.(map[string]any)
		features[i] = client.GCPFeatureInput{
			Name:    m["name"].(string),
			Enabled: m["enabled"].(bool),
		}
	}
	return features
}
