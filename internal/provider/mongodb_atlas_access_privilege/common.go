package mongodb_atlas_access_privilege

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc           = "The unique identifier of the MongoDB Atlas access privilege"
	nameDesc         = "The name of the MongoDB Atlas access privilege"
	descriptionDesc  = "The description of the MongoDB Atlas access privilege"
	grantsDesc       = "The list of privilege grants"
	builtinRolesDesc = "Atlas built-in roles granted on the admin database, for privileges no custom-role action carries, such as indexStats"
	needOneDesc      = ". Name at least one grant or built-in role."
	typeDesc         = "The type of access privilege"
)

var builtinRoles = []string{"clusterMonitor"}

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
		Description: grantsDesc + needOneDesc,
		Type:        schema.TypeList,
		Optional:    true,
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

	s["builtin_roles"] = &schema.Schema{
		Description: builtinRolesDesc + needOneDesc + " Supported: clusterMonitor.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(builtinRoles, false),
		},
	}

	return s
}

// customizeDiff mirrors the Hush API's rule that a privilege names at least
// one grant or built-in role. It reads the raw config, so a value unknown at
// plan time is never rejected.
func customizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	return checkGrantsOrBuiltinRoles(
		config.GetAttr("grants"), config.GetAttr("builtin_roles"))
}

func checkGrantsOrBuiltinRoles(grants, roles cty.Value) error {
	if named(grants) || named(roles) {
		return nil
	}
	return fmt.Errorf("name at least one grant or built-in role")
}

// named reports whether a list or set may hold an element. A value not
// wholly known counts: an unknown element is still an element.
func named(v cty.Value) bool {
	if v.IsNull() {
		return false
	}
	return !v.IsWhollyKnown() || v.LengthInt() > 0
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
		"builtin_roles": {
			Description: builtinRolesDesc,
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func expandBuiltinRoles(set *schema.Set) []string {
	roles := []string{}
	for _, r := range set.List() {
		roles = append(roles, r.(string))
	}
	return roles
}

func flattenBuiltinRoles(roles []string) []any {
	result := make([]any, len(roles))
	for i, r := range roles {
		result[i] = r
	}
	return result
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
