package rabbitmq_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the RabbitMQ access privilege"
	nameDesc        = "The name of the RabbitMQ access privilege"
	descriptionDesc = "The description of the RabbitMQ access privilege"
	permissionsDesc = "The RabbitMQ permission entries"
	vhostDesc       = "The RabbitMQ virtual host"
	configureDesc   = "The configure permission pattern (regex)"
	writeDesc       = "The write permission pattern (regex)"
	readDesc        = "The read permission pattern (regex)"
	tagsDesc        = "The list of RabbitMQ user tags"
	typeDesc        = "The type of access privilege"
)

var validTags = []string{"administrator", "monitoring", "policymaker", "management", "impersonator", "none"}

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
	s["permissions"] = &schema.Schema{
		Description: permissionsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"vhost": {
					Description: vhostDesc,
					Type:        schema.TypeString,
					Required:    true,
				},
				"configure": {
					Description: configureDesc,
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
				},
				"write": {
					Description: writeDesc,
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
				},
				"read": {
					Description: readDesc,
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
				},
			},
		},
	}
	s["tags"] = &schema.Schema{
		Description: tagsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(validTags, false),
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
		"permissions": {
			Description: permissionsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"vhost": {
						Description: vhostDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"configure": {
						Description: configureDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"write": {
						Description: writeDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"read": {
						Description: readDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"tags": {
			Description: tagsDesc,
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

func expandPermissions(list []any) []client.RabbitmqPermissionEntry {
	permissions := make([]client.RabbitmqPermissionEntry, len(list))
	for i, item := range list {
		m := item.(map[string]any)
		permissions[i] = client.RabbitmqPermissionEntry{
			Vhost:     m["vhost"].(string),
			Configure: m["configure"].(string),
			Write:     m["write"].(string),
			Read:      m["read"].(string),
		}
	}
	return permissions
}

func flattenPermissions(permissions []client.RabbitmqPermissionEntry) []any {
	result := make([]any, len(permissions))
	for i, p := range permissions {
		result[i] = map[string]any{
			"vhost":     p.Vhost,
			"configure": p.Configure,
			"write":     p.Write,
			"read":      p.Read,
		}
	}
	return result
}
