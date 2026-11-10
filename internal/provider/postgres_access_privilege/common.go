package postgres_access_privilege

import (
	"context"
	"errors"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc          = "The unique identifier of the PostgreSQL access privilege"
	nameDesc        = "The name of the PostgreSQL access privilege"
	descriptionDesc = "The description of the PostgreSQL access privilege"
	grantsDesc      = "The list of privilege grants"
	memberOfDesc    = "Roles the user is made a member of, with INHERIT and SET, so it can transfer what it creates to them (ALTER ... OWNER TO). The root credential's user needs ADMIN OPTION on each role, or SUPERUSER for a superuser role. Requires access-manager chart 0.34.0 or later on the deployments of policies using the privilege."
	typeDesc        = "The type of access privilege"
)

var errNothingGranted = errors.New("at least one of grants or member_of is required")

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
	s["member_of"] = &schema.Schema{
		Description: memberOfDesc,
		Type:        schema.TypeSet,
		Optional:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 63),
				validation.StringNotInSlice([]string{"public"}, false),
			),
		},
	}
	s["grants"] = &schema.Schema{
		Description: grantsDesc + ". At least one of grants or member_of is required.",
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
					Description: "The list of PostgreSQL privileges (e.g., SELECT, INSERT, UPDATE, DELETE)",
				},
				"object_type": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"TABLE", "SEQUENCE", "FUNCTION", "SCHEMA", "DATABASE"}, false),
					Description:  "The type of database object (TABLE, SEQUENCE, FUNCTION, SCHEMA, DATABASE)",
				},
				"object_names": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The names of the database objects",
				},
				"column_names": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The names of the columns (for column-level privileges)",
				},
				"all_in_schema": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Grant on all objects of the given type in the specified schema",
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
						Description: "The list of PostgreSQL privileges",
					},
					"object_type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The type of database object",
					},
					"object_names": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "The names of the database objects",
					},
					"column_names": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "The names of the columns",
					},
					"all_in_schema": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Grant on all objects of the given type in the specified schema",
					},
				},
			},
		},
		"member_of": {
			Description: memberOfDesc,
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

// requireGrantsOrMemberOf mirrors midgard, which refuses a privilege that
// grants nothing. It reads the raw config, where a value unknown at plan time
// is still present; the resolved diff drops an unknown set element.
func requireGrantsOrMemberOf(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	if grantsSomething(config.GetAttr("grants")) || grantsSomething(config.GetAttr("member_of")) {
		return nil
	}
	return errNothingGranted
}

// grantsSomething counts a value not wholly known as granting, leaving it to
// the API at apply.
func grantsSomething(v cty.Value) bool {
	if v.IsNull() {
		return false
	}
	if !v.IsWhollyKnown() {
		return true
	}
	return v.LengthInt() > 0
}

func expandMemberOf(set *schema.Set) []string {
	roles := make([]string, 0, set.Len())
	for _, r := range set.List() {
		roles = append(roles, r.(string))
	}
	return roles
}

func expandGrants(list []any) []client.PostgresGrant {
	grants := make([]client.PostgresGrant, len(list))
	for i, v := range list {
		m := v.(map[string]any)
		grant := client.PostgresGrant{
			ObjectType: m["object_type"].(string),
		}
		if privs, ok := m["privileges"].([]any); ok {
			grant.Privileges = make([]string, len(privs))
			for j, p := range privs {
				grant.Privileges[j] = p.(string)
			}
		}
		if names, ok := m["object_names"].([]any); ok && len(names) > 0 {
			grant.ObjectNames = make([]string, len(names))
			for j, n := range names {
				grant.ObjectNames[j] = n.(string)
			}
		}
		if cols, ok := m["column_names"].([]any); ok && len(cols) > 0 {
			grant.ColumnNames = make([]string, len(cols))
			for j, c := range cols {
				grant.ColumnNames[j] = c.(string)
			}
		}
		if ais, ok := m["all_in_schema"].(bool); ok {
			grant.AllInSchema = ais
		}
		grants[i] = grant
	}
	return grants
}

func flattenGrants(grants []client.PostgresGrant) []any {
	result := make([]any, len(grants))
	for i, g := range grants {
		privileges := g.Privileges
		if privileges == nil {
			privileges = []string{}
		}
		objectNames := g.ObjectNames
		if objectNames == nil {
			objectNames = []string{}
		}
		columnNames := g.ColumnNames
		if columnNames == nil {
			columnNames = []string{}
		}
		m := map[string]any{
			"privileges":    privileges,
			"object_type":   g.ObjectType,
			"object_names":  objectNames,
			"column_names":  columnNames,
			"all_in_schema": g.AllInSchema,
		}
		result[i] = m
	}
	return result
}
