package gitlab_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the GitLab access credential"
	nameDesc          = "The name of the GitLab access credential"
	descriptionDesc   = "The description of the GitLab access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential"
	tokenDesc         = "The GitLab API token"
	tokenWODesc       = "The GitLab API token (write-only). More secure than `token` because Terraform will not store this value in the state file."
	tokenWOVerDesc    = "Used to trigger updates for `token_wo`. Change when the token changes."
	baseURLDesc       = "The GitLab instance URL (default: https://gitlab.com)"
	resourceTypeDesc  = "The type of GitLab resource to manage (group or project)"
	resourceIDDesc    = "The GitLab group or project ID"
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
	s["token"] = &schema.Schema{
		Description:   tokenDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"token_wo"},
		ExactlyOneOf:  []string{"token", "token_wo"},
	}
	s["token_wo"] = &schema.Schema{
		Description:   tokenWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"token"},
		ExactlyOneOf:  []string{"token", "token_wo"},
		RequiredWith:  []string{"token_wo_version"},
	}
	s["token_wo_version"] = &schema.Schema{
		Description:  tokenWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"token_wo"},
	}
	s["base_url"] = &schema.Schema{
		Description:  baseURLDesc,
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "https://gitlab.com",
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
	}
	s["resource_type"] = &schema.Schema{
		Description:  resourceTypeDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"group", "project"}, false),
	}
	s["resource_id"] = &schema.Schema{
		Description: resourceIDDesc,
		Type:        schema.TypeString,
		Required:    true,
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
		"base_url": {
			Description: baseURLDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"resource_type": {
			Description: resourceTypeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"resource_id": {
			Description: resourceIDDesc,
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
