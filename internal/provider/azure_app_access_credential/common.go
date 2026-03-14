package azure_app_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc              = "The unique identifier of the Azure app access credential"
	nameDesc            = "The name of the Azure app access credential"
	descriptionDesc     = "The description of the Azure app access credential"
	deploymentIDsDesc   = "List of deployment IDs that can access this credential"
	tenantIDDesc        = "The Azure tenant ID (must be a valid UUID)"
	clientIDDesc        = "The Azure client ID (must be a valid UUID)"
	clientSecretDesc    = "The Azure client secret"
	clientSecretWODesc  = "The Azure client secret (write-only). This is a write-only attribute that is more secure than `client_secret` because Terraform will not store this value in the state file. Either `client_secret` or `client_secret_wo` must be specified."
	clientSecretWOVDesc = "Used to trigger updates for `client_secret_wo`. This value should be changed when the client secret content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	typeDesc            = "The type of access credential"
	kindDesc            = "The kind of access credential"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ResourceSchema() map[string]*schema.Schema {
	s := DataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 1000),
	}
	s["deployment_ids"] = &schema.Schema{
		Description: deploymentIDsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["tenant_id"] = &schema.Schema{
		Description:  tenantIDDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringMatch(uuidRegex, "tenant_id must be a valid UUID"),
	}
	s["client_id"] = &schema.Schema{
		Description:  clientIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(uuidRegex, "client_id must be a valid UUID"),
	}
	s["client_secret"] = &schema.Schema{
		Description:   clientSecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"client_secret_wo"},
		RequiredWith:  []string{"client_id"},
	}
	s["client_secret_wo"] = &schema.Schema{
		Description:   clientSecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"client_secret"},
		RequiredWith:  []string{"client_secret_wo_version", "client_id"},
	}
	s["client_secret_wo_version"] = &schema.Schema{
		Description:  clientSecretWOVDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"client_secret_wo"},
	}

	return s
}

func DataSourceSchema() map[string]*schema.Schema {
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
		"deployment_ids": {
			Description: deploymentIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"tenant_id": {
			Description: tenantIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
