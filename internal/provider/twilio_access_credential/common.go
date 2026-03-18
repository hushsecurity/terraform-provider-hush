package twilio_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                = "The unique identifier of the Twilio access credential"
	nameDesc              = "The name of the Twilio access credential"
	descriptionDesc       = "The description of the Twilio access credential"
	deploymentIDsDesc     = "List of deployment IDs that can access this credential"
	accountSIDDesc        = "The Twilio Account SID"
	apiKeySIDDesc         = "The Twilio API Key SID"
	apiKeySecretDesc      = "The Twilio API Key Secret"
	apiKeySecretWODesc    = "The Twilio API Key Secret (write-only). More secure than `api_key_secret` because Terraform will not store this value in the state file."
	apiKeySecretWOVerDesc = "Used to trigger updates for `api_key_secret_wo`. Change when the secret changes."
	typeDesc              = "The type of access credential"
	kindDesc              = "The kind of access credential"
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
	s["deployment_ids"] = &schema.Schema{
		Description: deploymentIDsDesc,
		Type:        schema.TypeList,
		Required:    true,
		ForceNew:    true,
		MinItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["account_sid"] = &schema.Schema{
		Description: accountSIDDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["api_key_sid"] = &schema.Schema{
		Description: apiKeySIDDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["api_key_secret"] = &schema.Schema{
		Description:   apiKeySecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"api_key_secret_wo"},
		ExactlyOneOf:  []string{"api_key_secret", "api_key_secret_wo"},
	}
	s["api_key_secret_wo"] = &schema.Schema{
		Description:   apiKeySecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"api_key_secret"},
		ExactlyOneOf:  []string{"api_key_secret", "api_key_secret_wo"},
		RequiredWith:  []string{"api_key_secret_wo_version"},
	}
	s["api_key_secret_wo_version"] = &schema.Schema{
		Description:  apiKeySecretWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"api_key_secret_wo"},
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
		"deployment_ids": {
			Description: deploymentIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"account_sid": {
			Description: accountSIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"api_key_sid": {
			Description: apiKeySIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
