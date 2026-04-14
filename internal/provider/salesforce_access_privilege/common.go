package salesforce_access_privilege

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc          = "The unique identifier of the Salesforce access privilege"
	nameDesc        = "The name of the Salesforce access privilege"
	descriptionDesc = "The description of the Salesforce access privilege"
	runAsUserDesc   = "The Salesforce user to impersonate (email address)"
	scopesDesc      = "The list of Salesforce OAuth2 scopes"
	typeDesc        = "The type of access privilege"
)

var validScopes = []string{
	"Api",
	"Web",
	"Full",
	"RefreshToken",
	"OpenID",
	"Profile",
	"Email",
	"Address",
	"Phone",
	"OfflineAccess",
	"CustomPermissions",
	"Lightning",
	"Content",
	"Chatter",
	"Wave",
	"Eclair",
	"Pardot",
	"CDPIngest",
	"CDPProfile",
	"CDPQuery",
	"Chatbot",
	"CDPSegment",
	"CDPIdentityResolution",
	"CDPCalculatedInsight",
	"SFApiPlatform",
	"Interaction",
	"CDP",
	"EinsteinGPT",
	"PwdlessLogin",
	"ForgotPassword",
	"UserRegistration",
	"MCP",
	"SCRT",
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
	s["run_as_user"] = &schema.Schema{
		Description:  runAsUserDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
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
		"run_as_user": {
			Description: runAsUserDesc,
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
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
