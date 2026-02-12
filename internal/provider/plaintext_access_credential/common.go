package plaintext_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc              = "The unique identifier of the plaintext access credential"
	nameDesc            = "The name of the plaintext access credential"
	descriptionDesc     = "The description of the plaintext access credential"
	deploymentIDsDesc   = "List of deployment IDs that can access this credential"
	secretDesc          = "The secret value for the plaintext credential"
	secretWODesc        = "The secret value for the plaintext credential (write-only). This is a write-only attribute that is more secure than `secret` because Terraform will not store this value in the state file. Either `secret` or `secret_wo` must be specified."
	secretWOVersionDesc = "Used to trigger updates for `secret_wo`. This value should be changed when the secret content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	typeDesc            = "The type of access credential (always PLAINTEXT for this resource)"
)

func PlaintextAccessCredentialResourceSchema() map[string]*schema.Schema {
	s := PlaintextAccessCredentialDataSourceSchema()

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
	s["secret"] = &schema.Schema{
		Description:   secretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ForceNew:      true, // Secret cannot be updated, requires recreation
		ConflictsWith: []string{"secret_wo"},
		ExactlyOneOf:  []string{"secret", "secret_wo"},
	}
	s["secret_wo"] = &schema.Schema{
		Description:   secretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"secret"},
		ExactlyOneOf:  []string{"secret", "secret_wo"},
		RequiredWith:  []string{"secret_wo_version"},
	}
	s["secret_wo_version"] = &schema.Schema{
		Description:  secretWOVersionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"secret_wo"},
	}

	return s
}

func PlaintextAccessCredentialDataSourceSchema() map[string]*schema.Schema {
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
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
