package azure_app_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc          = "The unique identifier of the Azure app access privilege"
	nameDesc        = "The name of the Azure app access privilege"
	descriptionDesc = "The description of the Azure app access privilege"
	appIDDesc       = "The Azure application ID (UUID). Mutually exclusive with app_config."
	appConfigDesc   = "The Azure application configuration. Mutually exclusive with app_id."
	displayNameDesc = "The display name of the Azure application"
	rolesDesc       = "The list of Azure roles"
	roleNameDesc    = "The role name"
	roleScopeDesc   = "The role scope"
	typeDesc        = "The type of access privilege"
)

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
	s["app_id"] = &schema.Schema{
		Description:   appIDDesc,
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"app_config"},
		ExactlyOneOf:  []string{"app_id", "app_config"},
	}
	s["app_config"] = &schema.Schema{
		Description:   appConfigDesc,
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"app_id"},
		ExactlyOneOf:  []string{"app_id", "app_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_name": {
					Description:  displayNameDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				"roles": {
					Description: rolesDesc,
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Description: roleNameDesc,
								Type:        schema.TypeString,
								Required:    true,
							},
							"scope": {
								Description: roleScopeDesc,
								Type:        schema.TypeString,
								Required:    true,
							},
						},
					},
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
		"app_id": {
			Description: appIDDesc,
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
					"roles": {
						Description: rolesDesc,
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: roleNameDesc,
									Type:        schema.TypeString,
									Computed:    true,
								},
								"scope": {
									Description: roleScopeDesc,
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
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
