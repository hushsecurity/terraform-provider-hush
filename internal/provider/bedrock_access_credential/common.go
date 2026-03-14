package bedrock_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                   = "The unique identifier of the Bedrock access credential"
	nameDesc                 = "The name of the Bedrock access credential"
	descriptionDesc          = "The description of the Bedrock access credential"
	deploymentIDsDesc        = "List of deployment IDs that can access this credential"
	regionDesc               = "The AWS region for Bedrock (e.g., us-east-1)"
	accessKeyIDDesc          = "The AWS access key ID. Must be set together with secret_access_key or secret_access_key_wo. Omit for provider credentials mode."
	secretAccessKeyDesc      = "The AWS secret access key"
	secretAccessKeyWODesc    = "The AWS secret access key (write-only). This is a write-only attribute that is more secure than `secret_access_key` because Terraform will not store this value in the state file."
	secretAccessKeyWOVerDesc = "Used to trigger updates for `secret_access_key_wo`. This value should be changed when the secret access key content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	hasProviderCredsDesc     = "Whether the credential uses AWS provider credentials (no explicit access key)"
	typeDesc                 = "The type of access credential"
	kindDesc                 = "The kind of access credential"
)

var awsRegionRegexp = regexp.MustCompile(`^[a-z]{2,}(-[a-z]+)+-\d+$`)

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
	s["region"] = &schema.Schema{
		Description:  regionDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringMatch(awsRegionRegexp, "must be a valid AWS region (e.g., us-east-1)"),
	}
	s["access_key_id"] = &schema.Schema{
		Description: accessKeyIDDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["secret_access_key"] = &schema.Schema{
		Description:   secretAccessKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"secret_access_key_wo"},
	}
	s["secret_access_key_wo"] = &schema.Schema{
		Description:   secretAccessKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"secret_access_key"},
		RequiredWith:  []string{"secret_access_key_wo_version"},
	}
	s["secret_access_key_wo_version"] = &schema.Schema{
		Description:  secretAccessKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"secret_access_key_wo"},
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
		"region": {
			Description: regionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"access_key_id": {
			Description: accessKeyIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"has_provider_credentials": {
			Description: hasProviderCredsDesc,
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
