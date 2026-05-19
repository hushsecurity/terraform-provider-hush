package gitlab_integration

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
	idDesc                 = "The unique identifier of the GitLab integration"
	nameDesc               = "The name of the GitLab integration"
	descriptionDesc        = "The description of the GitLab integration"
	tokenDesc              = "The GitLab personal access token or group access token"
	tokenWODesc            = "The GitLab token (write-only). This is more secure than `token` because Terraform will not store this value in the state file. Either `token` or `token_wo` must be specified."
	tokenWOVersionDesc     = "Used to trigger updates for `token_wo`. This value should be changed when the token changes. Can be any value (e.g., a timestamp, version number, or hash)."
	groupIDDesc            = "The GitLab group ID to scan. Mutually exclusive with `project_id`."
	projectIDDesc          = "The GitLab project ID to scan. Mutually exclusive with `group_id`."
	groupDesc              = "The resolved GitLab group name (computed)"
	visibilitiesDesc       = "List of repository visibilities to scan. Valid values: `private`, `public`, `internal`."
	baseURLDesc            = "The base URL of the GitLab instance. Defaults to `https://gitlab.com`."
	selectedReposDesc      = "List of specific repository names to scan. If not specified, all repositories in the group/project are scanned."
	enablePRScansDesc      = "Whether to enable pull/merge request scanning for this integration"
	botNameDesc            = "The bot name used for GitLab integration (computed by the API)"
	onpremDeploymentIDDesc = "The ID of the on-premises deployment to associate with this integration"
	statusDesc             = "The current status of the integration"
	statusMessageDesc      = "Additional details about the integration status"
	typeDesc               = "The type of integration (always 'gitlab' for this resource)"
	createdAtDesc          = "The timestamp when the integration was created"
	modifiedAtDesc         = "The timestamp when the integration was last modified"
)

func GitlabIntegrationResourceSchema() map[string]*schema.Schema {
	s := GitlabIntegrationDataSourceSchema()

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
	s["group_id"] = &schema.Schema{
		Description:   groupIDDesc,
		Type:          schema.TypeInt,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"project_id"},
		AtLeastOneOf:  []string{"group_id", "project_id"},
	}
	s["project_id"] = &schema.Schema{
		Description:   projectIDDesc,
		Type:          schema.TypeInt,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"group_id"},
		AtLeastOneOf:  []string{"group_id", "project_id"},
	}
	s["visibilities"] = &schema.Schema{
		Description: visibilitiesDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"private", "public", "internal"}, false),
		},
	}
	s["base_url"] = &schema.Schema{
		Description: baseURLDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	}
	s["selected_repos"] = &schema.Schema{
		Description: selectedReposDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	s["enable_pr_scans"] = &schema.Schema{
		Description: enablePRScansDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	}
	s["onprem_deployment_id"] = &schema.Schema{
		Description:  onpremDeploymentIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "onprem_deployment_id must start with 'dep-'"),
	}

	return s
}

func GitlabIntegrationDataSourceSchema() map[string]*schema.Schema {
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
		"group_id": {
			Description: groupIDDesc,
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"project_id": {
			Description: projectIDDesc,
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"group": {
			Description: groupDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"visibilities": {
			Description: visibilitiesDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"base_url": {
			Description: baseURLDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"selected_repos": {
			Description: selectedReposDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"enable_pr_scans": {
			Description: enablePRScansDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"bot_name": {
			Description: botNameDesc,
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
	}
}

func gitlabIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.GitlabIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetGitlabIntegration(ctx, c, id)
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
		integration, err = client.GetGitlabIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no GitLab integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetGitlabIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup GitLab integration by name '%s': %w", integrationName, lookupErr))
		}

		switch len(integrations) {
		case 0:
			return diag.Errorf("no GitLab integration found with name: %s", integrationName)
		case 1:
			// List response only has base fields; fetch full type-specific details
			integration, err = client.GetGitlabIntegration(ctx, c, integrations[0].ID)
			if err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("multiple GitLab integrations found with name '%s'. Use the integration ID instead for exact matching", integrationName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(integration.ID)
	}

	if diags := setGitlabIntegrationFields(d, integration); diags.HasError() {
		return diags
	}

	return nil
}

func setGitlabIntegrationFields(d *schema.ResourceData, integration *client.GitlabIntegration) diag.Diagnostics {
	fields := map[string]any{
		"name":                 integration.Name,
		"description":          integration.Description,
		"group":                integration.Group,
		"base_url":             integration.BaseURL,
		"bot_name":             integration.BotName,
		"onprem_deployment_id": integration.OnpremDeploymentID,
		"status":               integration.Status,
		"status_message":       integration.StatusMessage,
		"type":                 integration.Type,
		"created_at":           integration.CreatedAt,
		"modified_at":          integration.ModifiedAt,
	}

	if integration.EnablePRScans != nil {
		fields["enable_pr_scans"] = *integration.EnablePRScans
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if integration.GroupID != nil {
		if err := d.Set("group_id", *integration.GroupID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set group_id: %w", err))
		}
	}
	if integration.ProjectID != nil {
		if err := d.Set("project_id", *integration.ProjectID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set project_id: %w", err))
		}
	}
	if integration.Visibilities != nil {
		if err := d.Set("visibilities", integration.Visibilities); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set visibilities: %w", err))
		}
	}
	if integration.SelectedRepos != nil {
		if err := d.Set("selected_repos", integration.SelectedRepos); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set selected_repos: %w", err))
		}
	}

	return nil
}
