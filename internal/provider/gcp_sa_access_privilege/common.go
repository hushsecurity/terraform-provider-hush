package gcp_sa_access_privilege

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc          = "The unique identifier of the GCP SA access privilege"
	nameDesc        = "The name of the GCP SA access privilege"
	descriptionDesc = "The description of the GCP SA access privilege"
	projectIDDesc   = "The GCP project ID"
	saEmailDesc     = "The service account email. Mutually exclusive with sa_config."
	saConfigDesc    = "The service account configuration. Mutually exclusive with sa_email."
	displayNameDesc = "The display name of the service account"
	rolesDesc       = "The list of GCP IAM roles"
	typeDesc        = "The type of access privilege"
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
	s["project_id"] = &schema.Schema{
		Description:  projectIDDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]+$`), "project_id must contain only lowercase letters, numbers, and hyphens"),
	}
	s["sa_email"] = &schema.Schema{
		Description:   saEmailDesc,
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"sa_config"},
		ExactlyOneOf:  []string{"sa_email", "sa_config"},
	}
	s["sa_config"] = &schema.Schema{
		Description:   saConfigDesc,
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sa_email"},
		ExactlyOneOf:  []string{"sa_email", "sa_config"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_name": {
					Description:  displayNameDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				"roles": {
					Description: rolesDesc,
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
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
		"project_id": {
			Description: projectIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"sa_email": {
			Description: saEmailDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"sa_config": {
			Description: saConfigDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Description: displayNameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"roles": {
						Description: rolesDesc,
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
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
