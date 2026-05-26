package temporal_cloud_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the Temporal Cloud access privilege"
	nameDesc        = "The name of the Temporal Cloud access privilege"
	descriptionDesc = "The description of the Temporal Cloud access privilege"
	grantsDesc      = "The list of namespace permission grants"
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
	s["grants"] = &schema.Schema{
		Description: grantsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"namespace": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
					Description:  "The Temporal Cloud namespace",
				},
				"permission": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"read", "write", "admin"}, false),
					Description:  "The permission level (read, write, or admin)",
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
		"grants": {
			Description: grantsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"namespace": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The Temporal Cloud namespace",
					},
					"permission": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The permission level",
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

func expandGrants(list []any) []client.TemporalCloudGrant {
	grants := make([]client.TemporalCloudGrant, len(list))
	for i, v := range list {
		m := v.(map[string]any)
		grants[i] = client.TemporalCloudGrant{
			Namespace:  m["namespace"].(string),
			Permission: m["permission"].(string),
		}
	}
	return grants
}

func flattenGrants(grants []client.TemporalCloudGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		result[i] = map[string]any{
			"namespace":  g.Namespace,
			"permission": g.Permission,
		}
	}
	return result
}
