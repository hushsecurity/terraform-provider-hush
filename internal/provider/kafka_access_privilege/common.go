package kafka_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc             = "The unique identifier of the Kafka access privilege"
	nameDesc           = "The name of the Kafka access privilege"
	descriptionDesc    = "The description of the Kafka access privilege"
	aclsDesc           = "The Kafka ACL entries granted by this privilege"
	resourceTypeDesc   = "The Kafka resource type the ACL applies to (e.g., Topic, Group, Cluster, TransactionalId)"
	resourceNameDesc   = "The name of the Kafka resource the ACL applies to (use `*` for all)"
	patternTypeDesc    = "How resource_name is matched: LITERAL (exact) or PREFIXED (name prefix)"
	operationDesc      = "The Kafka operation the ACL applies to (e.g., Read, Write, Create, All)"
	permissionTypeDesc = "Whether the ACL grants (ALLOW) or denies (DENY) the operation"
	hostDesc           = "The host the ACL applies to (defaults to `*`, all hosts)"
	typeDesc           = "The type of access privilege"
)

var (
	validPatternTypes    = []string{"LITERAL", "PREFIXED"}
	validPermissionTypes = []string{"ALLOW", "DENY"}
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
	s["acls"] = &schema.Schema{
		Description: aclsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"resource_type": {
					Description:  resourceTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				"resource_name": {
					Description:  resourceNameDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				"pattern_type": {
					Description:  patternTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(validPatternTypes, false),
				},
				"operation": {
					Description:  operationDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				"permission_type": {
					Description:  permissionTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(validPermissionTypes, false),
				},
				"host": {
					Description: hostDesc,
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "*",
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
		"acls": {
			Description: aclsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"resource_type": {
						Description: resourceTypeDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"resource_name": {
						Description: resourceNameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"pattern_type": {
						Description: patternTypeDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"operation": {
						Description: operationDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"permission_type": {
						Description: permissionTypeDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"host": {
						Description: hostDesc,
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

func expandACLs(list []any) []client.KafkaAclEntry {
	acls := make([]client.KafkaAclEntry, len(list))
	for i, item := range list {
		m := item.(map[string]any)
		acls[i] = client.KafkaAclEntry{
			ResourceType:   m["resource_type"].(string),
			ResourceName:   m["resource_name"].(string),
			PatternType:    m["pattern_type"].(string),
			Operation:      m["operation"].(string),
			PermissionType: m["permission_type"].(string),
			Host:           m["host"].(string),
		}
	}
	return acls
}

func flattenACLs(acls []client.KafkaAclEntry) []any {
	result := make([]any, len(acls))
	for i, a := range acls {
		result[i] = map[string]any{
			"resource_type":   a.ResourceType,
			"resource_name":   a.ResourceName,
			"pattern_type":    a.PatternType,
			"operation":       a.Operation,
			"permission_type": a.PermissionType,
			"host":            a.Host,
		}
	}
	return result
}
