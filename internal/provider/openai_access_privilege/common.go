package openai_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc             = "The unique identifier of the OpenAI access privilege"
	nameDesc           = "The name of the OpenAI access privilege"
	descriptionDesc    = "The description of the OpenAI access privilege"
	permissionTypeDesc = "The permission type (Owner, Viewer, Member, or Restricted)"
	permissionsDesc    = "The list of specific permissions (applicable when permission_type is Restricted)"
	typeDesc           = "The type of access privilege"
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
	s["permission_type"] = &schema.Schema{
		Description:  permissionTypeDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"Owner", "Viewer", "Member", "Restricted"}, false),
	}
	s["permissions"] = &schema.Schema{
		Description: permissionsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the permission",
				},
				"level": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The level of the permission",
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
		"permission_type": {
			Description: permissionTypeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"permissions": {
			Description: permissionsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The name of the permission",
					},
					"level": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The level of the permission",
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

func expandPermissions(list []any) []client.OpenAIPermission {
	permissions := make([]client.OpenAIPermission, len(list))
	for i, v := range list {
		m := v.(map[string]any)
		permissions[i] = client.OpenAIPermission{
			Name:  m["name"].(string),
			Level: m["level"].(string),
		}
	}
	return permissions
}

func flattenPermissions(permissions []client.OpenAIPermission) []any {
	result := make([]any, len(permissions))
	for i, p := range permissions {
		m := map[string]any{
			"name":  p.Name,
			"level": p.Level,
		}
		result[i] = m
	}
	return result
}
