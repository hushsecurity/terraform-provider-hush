package confluence_integration

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
	idDesc              = "The unique identifier of the Confluence integration"
	nameDesc            = "The name of the Confluence integration"
	descriptionDesc     = "The description of the Confluence integration"
	orgDomainDesc       = "The Confluence organization domain (e.g., mycompany.atlassian.net)"
	userDesc            = "The email address of the Confluence user for authentication"
	apiKeyDesc          = "The API key for Confluence authentication"
	apiKeyWODesc        = "The API key for Confluence authentication (write-only). This is more secure than `api_key` because Terraform will not store this value in the state file. Either `api_key` or `api_key_wo` must be specified."
	apiKeyWOVersionDesc = "Used to trigger updates for `api_key_wo`. This value should be changed when the API key changes. Can be any value (e.g., a timestamp, version number, or hash)."
	statusDesc          = "The current status of the integration"
)

func ConfluenceIntegrationResourceSchema() map[string]*schema.Schema {
	s := ConfluenceIntegrationDataSourceSchema()

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
	s["org_domain"] = &schema.Schema{
		Description: orgDomainDesc,
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}
	s["user"] = &schema.Schema{
		Description: userDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["api_key"] = &schema.Schema{
		Description:   apiKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"api_key_wo"},
		ExactlyOneOf:  []string{"api_key", "api_key_wo"},
	}
	s["api_key_wo"] = &schema.Schema{
		Description:   apiKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"api_key"},
		ExactlyOneOf:  []string{"api_key", "api_key_wo"},
		RequiredWith:  []string{"api_key_wo_version"},
	}
	s["api_key_wo_version"] = &schema.Schema{
		Description:  apiKeyWOVersionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"api_key_wo"},
	}
	return s
}

func ConfluenceIntegrationDataSourceSchema() map[string]*schema.Schema {
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
		"org_domain": {
			Description: orgDomainDesc,
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

func confluenceIntegrationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var integration *client.ConfluenceIntegration
	var err error

	if id := d.Id(); id != "" {
		integration, err = client.GetConfluenceIntegration(ctx, c, id)
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
		integration, err = client.GetConfluenceIntegration(ctx, c, integrationID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no Confluence integration found with ID: %s", integrationID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		integrationName := name.(string)
		integrations, lookupErr := client.GetConfluenceIntegrationsByName(ctx, c, integrationName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup Confluence integration by name '%s': %w", integrationName, lookupErr))
		}

		switch len(integrations) {
		case 0:
			return diag.Errorf("no Confluence integration found with name: %s", integrationName)
		case 1:
			integration, err = client.GetConfluenceIntegration(ctx, c, integrations[0].ID)
			if err != nil {
				return diag.FromErr(err)
			}
		default:
			return diag.Errorf("multiple Confluence integrations found with name '%s'. Use the integration ID instead for exact matching", integrationName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(integration.ID)
	}

	if diags := setConfluenceIntegrationFields(d, integration); diags.HasError() {
		return diags
	}

	return nil
}

func setConfluenceIntegrationFields(d *schema.ResourceData, integration *client.ConfluenceIntegration) diag.Diagnostics {
	fields := map[string]any{
		"name":        integration.Name,
		"description": integration.Description,
		"org_domain":  integration.OrgDomain,
		"status":      integration.Status,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	return nil
}
