package gitlab_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc          = "The unique identifier of the GitLab access privilege"
	nameDesc        = "The name of the GitLab access privilege"
	descriptionDesc = "The description of the GitLab access privilege"
	scopesDesc      = "The list of GitLab API token scopes"
	accessLevelDesc = "The GitLab access level (Guest, Reporter, Developer, Maintainer, or Owner)"
	typeDesc        = "The type of access privilege"
)

var validScopes = []string{
	"api",
	"read_api",
	"read_repository",
	"write_repository",
	"read_registry",
	"write_registry",
	"read_virtual_registry",
	"write_virtual_registry",
	"create_runner",
	"manage_runner",
	"ai_features",
	"k8s_proxy",
	"self_rotate",
}

var validAccessLevels = []string{
	"Guest",
	"Reporter",
	"Developer",
	"Maintainer",
	"Owner",
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
	s["scopes"] = &schema.Schema{
		Description: scopesDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(validScopes, false),
		},
	}
	s["access_level"] = &schema.Schema{
		Description:  accessLevelDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(validAccessLevels, false),
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
		"scopes": {
			Description: scopesDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"access_level": {
			Description: accessLevelDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
