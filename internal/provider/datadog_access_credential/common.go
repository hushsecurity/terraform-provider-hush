package datadog_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the Datadog access credential"
	nameDesc          = "The name of the Datadog access credential"
	descriptionDesc   = "The description of the Datadog access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential"
	apiKeyDesc        = "The Datadog API key"
	apiKeyWODesc      = "The Datadog API key (write-only). This is a write-only attribute that is more secure than `api_key` because Terraform will not store this value in the state file. Either `api_key` or `api_key_wo` must be specified."
	apiKeyWOVerDesc   = "Used to trigger updates for `api_key_wo`. This value should be changed when the API key content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	appKeyDesc        = "The Datadog application key"
	appKeyWODesc      = "The Datadog application key (write-only). More secure than `app_key` because Terraform will not store this value in the state file."
	appKeyWOVerDesc   = "Used to trigger updates for `app_key_wo`. Change when the application key changes."
	siteDesc          = "The Datadog site (e.g., datadoghq.com, us3.datadoghq.com, us5.datadoghq.com, datadoghq.eu, ap1.datadoghq.com)"
	typeDesc          = "The type of access credential"
	kindDesc          = "The kind of access credential"
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
		MinItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["api_key"] = &schema.Schema{
		Description:   apiKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"api_key_wo"},
		ExactlyOneOf:  []string{"api_key", "api_key_wo"},
	}
	s["api_key_wo"] = &schema.Schema{
		Description:   apiKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"api_key"},
		ExactlyOneOf:  []string{"api_key", "api_key_wo"},
		RequiredWith:  []string{"api_key_wo_version"},
	}
	s["api_key_wo_version"] = &schema.Schema{
		Description:  apiKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"api_key_wo"},
	}
	s["app_key"] = &schema.Schema{
		Description:   appKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"app_key_wo"},
		ExactlyOneOf:  []string{"app_key", "app_key_wo"},
	}
	s["app_key_wo"] = &schema.Schema{
		Description:   appKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"app_key"},
		ExactlyOneOf:  []string{"app_key", "app_key_wo"},
		RequiredWith:  []string{"app_key_wo_version"},
	}
	s["app_key_wo_version"] = &schema.Schema{
		Description:  appKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"app_key_wo"},
	}
	s["site"] = &schema.Schema{
		Description: siteDesc,
		Type:        schema.TypeString,
		Optional:    true,
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
		"site": {
			Description: siteDesc,
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
