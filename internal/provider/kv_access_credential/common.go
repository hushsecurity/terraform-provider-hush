package kv_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the KV access credential"
	nameDesc          = "The name of the KV access credential"
	descriptionDesc   = "The description of the KV access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential"
	itemsDesc         = "List of key-value pairs for the credential"
	keyDesc           = "The key name for the environment variable"
	valueDesc         = "The value for the key-value pair"
	keysDesc          = "List of keys available in this credential (computed)"
	typeDesc          = "The type of access credential (always KV for this resource)"
)

func KVAccessCredentialResourceSchema() map[string]*schema.Schema {
	s := KVAccessCredentialDataSourceSchema()

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
		MinItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["items"] = &schema.Schema{
		Description: itemsDesc,
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		ForceNew:    true, // Items cannot be updated, requires recreation
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Description: keyDesc,
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 255),
						validation.StringDoesNotMatch(regexp.MustCompile(`^_?_?hush`), "Keys cannot start with '_hush' or '__hush'"),
					),
				},
				"value": {
					Description: valueDesc,
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
				},
			},
		},
	}

	return s
}

func KVAccessCredentialDataSourceSchema() map[string]*schema.Schema {
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
		"keys": {
			Description: keysDesc,
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
