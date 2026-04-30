package aws_access_key_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                   = "The unique identifier of the AWS access key access credential"
	nameDesc                 = "The name of the AWS access key access credential"
	descriptionDesc          = "The description of the AWS access key access credential"
	deploymentIDsDesc        = "List of deployment IDs that can access this credential"
	accessKeyIDValueDesc     = "The AWS access key ID"
	secretAccessKeyDesc      = "The AWS secret access key"
	secretAccessKeyWODesc    = "The AWS secret access key (write-only). This is a write-only attribute that is more secure than `secret_access_key` because Terraform will not store this value in the state file. Either `secret_access_key` or `secret_access_key_wo` must be specified."
	secretAccessKeyWOVerDesc = "Used to trigger updates for `secret_access_key_wo`. This value should be changed when the secret access key content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	permissionBoundaryDesc   = "Whether the linked Access Privilege policy should be attached to the dynamically created IAM user as a permission boundary instead of as a managed policy. When enabled, the Access Privilege must contain exactly one policy."
	typeDesc                 = "The type of access credential"
	kindDesc                 = "The kind of access credential"
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
	s["access_key_id_value"] = &schema.Schema{
		Description:  accessKeyIDValueDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^AKIA[0-9A-Z]{16}$`), "must be a valid AWS access key ID starting with 'AKIA'"),
	}
	s["secret_access_key"] = &schema.Schema{
		Description:   secretAccessKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"secret_access_key_wo"},
		RequiredWith:  []string{"access_key_id_value"},
	}
	s["secret_access_key_wo"] = &schema.Schema{
		Description:   secretAccessKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"secret_access_key"},
		RequiredWith:  []string{"secret_access_key_wo_version", "access_key_id_value"},
	}
	s["secret_access_key_wo_version"] = &schema.Schema{
		Description:  secretAccessKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"secret_access_key_wo"},
	}
	s["permission_boundary"] = &schema.Schema{
		Description: permissionBoundaryDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
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
		"access_key_id_value": {
			Description: accessKeyIDValueDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"permission_boundary": {
			Description: permissionBoundaryDesc,
			Type:        schema.TypeBool,
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
