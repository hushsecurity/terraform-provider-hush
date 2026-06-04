package bitbucket_integration

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
	idDesc             = "The unique identifier of the Bitbucket integration"
	nameDesc           = "The name of the Bitbucket integration"
	descriptionDesc    = "The description of the Bitbucket integration"
	workspaceSlugDesc  = "The Bitbucket workspace slug (e.g., my-workspace)"
	tokenDesc          = "The access token for Bitbucket authentication"
	tokenWODesc        = "The access token for Bitbucket authentication (write-only). This is more secure than `token` because Terraform will not store this value in the state file. Either `token` or `token_wo` must be specified."
	tokenWOVersionDesc = "Used to trigger updates for `token_wo`. This value should be changed when the token changes. Can be any value (e.g., a timestamp, version number, or hash)."
	statusDesc         = "The current status of the integration"
	statusMessageDesc  = "The status message providing additional details about the integration status"
)

var workspaceSlugPattern = regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`)

func BitbucketIntegrationResourceSchema() map[string]*schema.Schema {
	s := BitbucketIntegrationDataSourceSchema()

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
	s["workspace_slug"] = &schema.Schema{
		Description: workspaceSlugDesc,
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 200),
			validation.StringMatch(workspaceSlugPattern, "must contain only alphanumeric characters, hyphens, underscores, and periods"),
		),
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
	return s
}

func BitbucketIntegrationDataSourceSchema() map[string]*schema.Schema {
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
		"workspace_slug": {
			Description: workspaceSlugDesc,
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

func bitbucketIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.BitbucketIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetBitbucketIntegration(ctx, c, id)
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
		integration, err = client.GetBitbucketIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no Bitbucket integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetBitbucketIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup Bitbucket integration by name '%s': %w", integrationName, lookupErr))
		}

		switch len(integrations) {
		case 0:
			return diag.Errorf("no Bitbucket integration found with name: %s", integrationName)
		case 1:
			integration, err = client.GetBitbucketIntegration(ctx, c, integrations[0].ID)
			if err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("multiple Bitbucket integrations found with name '%s'. Use the integration ID instead for exact matching", integrationName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(integration.ID)
	}

	if diags := setBitbucketIntegrationFields(d, integration); diags.HasError() {
		return diags
	}

	return nil
}

func setBitbucketIntegrationFields(d *schema.ResourceData, integration *client.BitbucketIntegration) diag.Diagnostics {
	fields := map[string]any{
		"name":           integration.Name,
		"description":    integration.Description,
		"workspace_slug": integration.WorkspaceSlug,
		"status":         integration.Status,
		"status_message": integration.StatusMessage,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}
