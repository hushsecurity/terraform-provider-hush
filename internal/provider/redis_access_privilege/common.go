package redis_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the Redis access privilege"
	nameDesc        = "The name of the Redis access privilege"
	descriptionDesc = "The description of the Redis access privilege"
	grantsDesc      = "The list of Redis ACL grant entries"
	keysDesc        = "The key patterns this privilege applies to (e.g., \"*\", \"cache:*\")"
	channelsDesc    = "The Pub/Sub channel patterns this privilege applies to"
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
				"type": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"category", "command"}, false),
					Description:  "The type of grant entry (category or command)",
				},
				"action": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"include", "exclude"}, false),
					Description:  "The action for this grant entry (include or exclude)",
				},
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the Redis command or category (e.g., read, write, get, set)",
				},
			},
		},
	}
	s["keys"] = &schema.Schema{
		Description: keysDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	s["channels"] = &schema.Schema{
		Description: channelsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
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
					"type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The type of grant entry",
					},
					"action": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The action for this grant entry",
					},
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The name of the Redis command or category",
					},
				},
			},
		},
		"keys": {
			Description: keysDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"channels": {
			Description: channelsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func expandGrants(list []any) []client.RedisGrant {
	grants := make([]client.RedisGrant, len(list))
	for i, v := range list {
		m := v.(map[string]any)
		grants[i] = client.RedisGrant{
			Type:   m["type"].(string),
			Action: m["action"].(string),
			Name:   m["name"].(string),
		}
	}
	return grants
}

func flattenGrants(grants []client.RedisGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		result[i] = map[string]any{
			"type":   g.Type,
			"action": g.Action,
			"name":   g.Name,
		}
	}
	return result
}
