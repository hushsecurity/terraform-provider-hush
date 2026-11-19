package statsig_access_privilege

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                    = "The unique identifier of the Statsig access privilege"
	nameDesc                  = "The name of the Statsig access privilege"
	descriptionDesc           = "The description of the Statsig access privilege"
	keyTypeDesc               = "The kind of Statsig key to mint: `server` (a server secret key, the default) or `client` (a client API key)"
	environmentsDesc          = "The Statsig environments the key works in. Unset means every environment. Statsig refuses an environment the project does not have"
	targetAppIDDesc           = "The ID of the Statsig target app the key is limited to. Unset means every target app"
	secondaryTargetAppIDsDesc = "The IDs of further Statsig target apps the key is limited to"
	typeDesc                  = "The type of access privilege"
)

// midgard refuses an empty name and one with surrounding whitespace, as
// Python's strip() finds it: any Unicode space, as TrimSpace does.
var statsigName = validation.All(validation.StringIsNotEmpty, noSurroundingSpace)

func noSurroundingSpace(v any, k string) ([]string, []error) {
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == s {
		return nil, nil
	}
	return nil, []error{fmt.Errorf("%s must not start or end with whitespace", k)}
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
	s["key_type"] = &schema.Schema{
		Description:  keyTypeDesc,
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "server",
		ValidateFunc: validation.StringInSlice([]string{"server", "client"}, false),
	}
	s["environments"] = &schema.Schema{
		Description: environmentsDesc,
		Type:        schema.TypeSet,
		Optional:    true,
		MinItems:    1,
		Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: statsigName},
	}
	s["target_app_id"] = &schema.Schema{
		Description:  targetAppIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: statsigName,
	}
	s["secondary_target_app_ids"] = &schema.Schema{
		Description: secondaryTargetAppIDsDesc,
		Type:        schema.TypeSet,
		Optional:    true,
		MinItems:    1,
		Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: statsigName},
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
		"key_type": {
			Description: keyTypeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"environments": {
			Description: environmentsDesc,
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"target_app_id": {
			Description: targetAppIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"secondary_target_app_ids": {
			Description: secondaryTargetAppIDsDesc,
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
