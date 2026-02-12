package gemini_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                     = "The unique identifier of the Gemini access credential"
	nameDesc                   = "The name of the Gemini access credential"
	descriptionDesc            = "The description of the Gemini access credential"
	deploymentIDsDesc          = "List of deployment IDs that can access this credential"
	serviceAccountKeyDesc      = "The GCP service account key JSON"
	serviceAccountKeyWODesc    = "The GCP service account key JSON (write-only). This is a write-only attribute that is more secure than `service_account_key` because Terraform will not store this value in the state file. Either `service_account_key` or `service_account_key_wo` must be specified."
	serviceAccountKeyWOVerDesc = "Used to trigger updates for `service_account_key_wo`. This value should be changed when the service account key content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	projectIDDesc              = "The GCP project ID"
	typeDesc                   = "The type of access credential"
	kindDesc                   = "The kind of access credential"
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
	s["service_account_key"] = &schema.Schema{
		Description:   serviceAccountKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"service_account_key_wo"},
		ExactlyOneOf:  []string{"service_account_key", "service_account_key_wo"},
	}
	s["service_account_key_wo"] = &schema.Schema{
		Description:   serviceAccountKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"service_account_key"},
		ExactlyOneOf:  []string{"service_account_key", "service_account_key_wo"},
		RequiredWith:  []string{"service_account_key_wo_version"},
	}
	s["service_account_key_wo_version"] = &schema.Schema{
		Description:  serviceAccountKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"service_account_key_wo"},
	}
	s["project_id"] = &schema.Schema{
		Description:  projectIDDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]+$`), "project_id must contain only lowercase letters, numbers, and hyphens"),
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
		"project_id": {
			Description: projectIDDesc,
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
