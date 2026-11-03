package pinot_access_privilege

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc            = "The unique identifier of the Pinot access privilege"
	nameDesc          = "The name of the Pinot access privilege"
	descriptionDesc   = "The description of the Pinot access privilege"
	grantDesc         = "The tables the generated user can query through the Pinot broker. It gets no controller access. An empty grant means all tables"
	tablesDesc        = "The tables granted, named without their `_OFFLINE`/`_REALTIME` suffix. Omit for all tables; an empty list is refused, since Pinot reads it as all tables too"
	excludeTablesDesc = "The tables denied, even when `tables` is omitted, named without their `_OFFLINE`/`_REALTIME` suffix. Needs Pinot 1.4.0 or later: an older Pinot would ignore it and grant every table, so the policy fails there. An empty list is refused"
	typeDesc          = "The type of access privilege"
)

// validateTable mirrors midgard: Pinot matches raw table names, so a suffixed
// one would never match.
func validateTable(v any, k string) ([]string, []error) {
	t := v.(string)
	if t == "" {
		return nil, []error{fmt.Errorf("%s cannot contain an empty table name", k)}
	}
	for _, suffix := range []string{"_OFFLINE", "_REALTIME"} {
		if strings.HasSuffix(t, suffix) {
			return nil, []error{fmt.Errorf(
				"%s: table %q must be named without its %s suffix", k, t, suffix)}
		}
	}
	return nil, nil
}

func stringList(desc string, validate func(any, string) ([]string, []error)) *schema.Schema {
	return &schema.Schema{
		Description: desc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validate,
		},
	}
}

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

	s["grant"] = &schema.Schema{
		Description: grantDesc,
		Type:        schema.TypeList,
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tables":         stringList(tablesDesc, validateTable),
				"exclude_tables": stringList(excludeTablesDesc, validateTable),
			},
		},
	}

	return s
}

func computedList(desc string) *schema.Schema {
	return &schema.Schema{
		Description: desc,
		Type:        schema.TypeList,
		Computed:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
	}
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
		"grant": {
			Description: grantDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"tables":         computedList(tablesDesc),
					"exclude_tables": computedList(excludeTablesDesc),
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

// validateGrant checks at plan what the schema cannot.
//
// An empty list is never sent, since the SDK cannot tell it from an absent
// one: an empty tables or exclude_tables would silently grant all tables.
func validateGrant(_ context.Context, d *schema.ResourceDiff, _ any) error {
	grant := d.GetRawConfig().GetAttr("grant")
	if grant.IsNull() || !grant.IsKnown() || grant.LengthInt() == 0 {
		return nil
	}
	g := grant.Index(cty.NumberIntVal(0))
	if g.IsNull() || !g.IsKnown() {
		return nil
	}
	for _, attr := range []string{"tables", "exclude_tables"} {
		v := g.GetAttr(attr)
		if !v.IsNull() && v.IsKnown() && v.LengthInt() == 0 {
			return fmt.Errorf("grant.%s cannot be an empty list; omit it instead", attr)
		}
	}
	return nil
}

func toStrings(v any) []string {
	list, _ := v.([]any)
	out := make([]string, 0, len(list))
	for _, item := range list {
		out = append(out, item.(string))
	}
	return out
}

func expandGrant(list []any) client.PinotGrant {
	if len(list) == 0 || list[0] == nil {
		return client.PinotGrant{}
	}
	m := list[0].(map[string]any)
	return client.PinotGrant{
		Tables:        toStrings(m["tables"]),
		ExcludeTables: toStrings(m["exclude_tables"]),
	}
}

func flattenGrant(grant client.PinotGrant) []any {
	orEmpty := func(s []string) []string {
		if s == nil {
			return []string{}
		}
		return s
	}
	return []any{map[string]any{
		"tables":         orEmpty(grant.Tables),
		"exclude_tables": orEmpty(grant.ExcludeTables),
	}}
}
