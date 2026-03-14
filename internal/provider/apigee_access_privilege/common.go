package apigee_access_privilege

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc             = "The unique identifier of the Apigee access privilege"
	nameDesc           = "The name of the Apigee access privilege"
	descriptionDesc    = "The description of the Apigee access privilege"
	developerEmailDesc = "The developer email address for the Apigee app"
	projectIDDesc      = "The GCP project ID"
	apiProductsDesc    = "List of API product names"
	appNameDesc        = "The name of an existing Apigee developer app. Mutually exclusive with app_config."
	appConfigDesc      = "Configuration for creating a new Apigee developer app. Mutually exclusive with app_name."
	displayNameDesc    = "The display name for the new Apigee developer app"
	typeDesc           = "The type of access privilege"
)

var gcpProjectIDRegexp = regexp.MustCompile(`^[a-z0-9-]+$`)

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
	s["developer_email"] = &schema.Schema{
		Description: developerEmailDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["project_id"] = &schema.Schema{
		Description:  projectIDDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.All(validation.StringLenBetween(6, 30), validation.StringMatch(gcpProjectIDRegexp, "must contain only lowercase letters, numbers, and hyphens")),
	}
	s["api_products"] = &schema.Schema{
		Description: apiProductsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	s["app_name"] = &schema.Schema{
		Description:   appNameDesc,
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"app_config"},
		ExactlyOneOf:  []string{"app_name", "app_config"},
	}
	s["app_config"] = &schema.Schema{
		Description:   appConfigDesc,
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"app_name"},
		ExactlyOneOf:  []string{"app_name", "app_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_name": {
					Description: displayNameDesc,
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
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
		"developer_email": {
			Description: developerEmailDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"project_id": {
			Description: projectIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"api_products": {
			Description: apiProductsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"app_name": {
			Description: appNameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"app_config": {
			Description: appConfigDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Description: displayNameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func expandAppConfig(raw []any) *client.ApigeeAppConfig {
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	m := raw[0].(map[string]any)
	return &client.ApigeeAppConfig{
		DisplayName: m["display_name"].(string),
	}
}

func flattenAppConfig(config *client.ApigeeAppConfig) []any {
	if config == nil {
		return []any{}
	}
	return []any{
		map[string]any{
			"display_name": config.DisplayName,
		},
	}
}
