package mongodb_atlas_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the MongoDB Atlas access privilege"
	nameDesc        = "The name of the MongoDB Atlas access privilege"
	descriptionDesc = "The description of the MongoDB Atlas access privilege"
	grantsDesc      = "The list of privilege grants"
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
				"privileges": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The list of MongoDB Atlas privilege actions (e.g., FIND, INSERT, UPDATE), or \"all\" to grant every action for the resource type",
				},
				"resource_type": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"database", "collection"}, false),
					Description:  "The type of resource (database or collection)",
				},
				"resource_names": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The collection names the grant applies to (collection-level grants only)",
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
					"privileges": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "The list of MongoDB Atlas privilege actions",
					},
					"resource_type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The type of resource",
					},
					"resource_names": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "The collection names the grant applies to",
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

func expandGrants(list []any) []client.MongoDBAtlasGrant {
	grants := make([]client.MongoDBAtlasGrant, len(list))
	for i, v := range list {
		m := v.(map[string]any)
		grant := client.MongoDBAtlasGrant{
			ResourceType: m["resource_type"].(string),
		}
		if privs, ok := m["privileges"].([]any); ok {
			grant.Privileges = make([]string, len(privs))
			for j, p := range privs {
				grant.Privileges[j] = p.(string)
			}
		}
		if names, ok := m["resource_names"].([]any); ok && len(names) > 0 {
			grant.ResourceNames = make([]string, len(names))
			for j, n := range names {
				grant.ResourceNames[j] = n.(string)
			}
		}
		grants[i] = grant
	}
	return grants
}

func flattenGrants(grants []client.MongoDBAtlasGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		privileges := g.Privileges
		if privileges == nil {
			privileges = []string{}
		}
		resourceNames := g.ResourceNames
		if resourceNames == nil {
			resourceNames = []string{}
		}
		m := map[string]any{
			"privileges":     privileges,
			"resource_type":  g.ResourceType,
			"resource_names": resourceNames,
		}
		result[i] = m
	}
	return result
}
