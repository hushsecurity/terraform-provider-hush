package mongodb_atlas_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc              = "The unique identifier of the MongoDB Atlas access credential"
	nameDesc            = "The name of the MongoDB Atlas access credential"
	descriptionDesc     = "The description of the MongoDB Atlas access credential"
	deploymentIDsDesc   = "List of deployment IDs that can access this credential"
	groupIDDesc         = "The MongoDB Atlas project (group) ID"
	dbNameDesc          = "The name of the MongoDB Atlas database"
	hostDesc            = "The hostname of the MongoDB Atlas cluster"
	clientIDDesc        = "The MongoDB Atlas service account client ID (used together with `client_secret`)"
	clientSecretDesc    = "The MongoDB Atlas service account client secret"
	clientSecretWODesc  = "The MongoDB Atlas service account client secret (write-only). This is a write-only attribute that is more secure than `client_secret` because Terraform will not store this value in the state file."
	clientSecretWOVDesc = "Used to trigger updates for `client_secret_wo`. This value should be changed when the client secret content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	publicKeyDesc       = "The MongoDB Atlas API public key (used together with `private_key`)"
	privateKeyDesc      = "The MongoDB Atlas API private key"
	privateKeyWODesc    = "The MongoDB Atlas API private key (write-only). This is a write-only attribute that is more secure than `private_key` because Terraform will not store this value in the state file."
	privateKeyWOVDesc   = "Used to trigger updates for `private_key_wo`. This value should be changed when the private key content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	typeDesc            = "The type of access credential"
	kindDesc            = "The kind of access credential"
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
		Description: deploymentIDsDesc + ". Changing this after creation is not supported; the credential must be deleted and recreated.",
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		MaxItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["group_id"] = &schema.Schema{
		Description: groupIDDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["db_name"] = &schema.Schema{
		Description: dbNameDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["host"] = &schema.Schema{
		Description: hostDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["client_id"] = &schema.Schema{
		Description:   clientIDDesc,
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"public_key", "private_key", "private_key_wo"},
	}
	s["client_secret"] = &schema.Schema{
		Description:   clientSecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"client_secret_wo", "public_key", "private_key", "private_key_wo"},
		RequiredWith:  []string{"client_id"},
	}
	s["client_secret_wo"] = &schema.Schema{
		Description:   clientSecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"client_secret", "public_key", "private_key", "private_key_wo"},
		RequiredWith:  []string{"client_secret_wo_version", "client_id"},
	}
	s["client_secret_wo_version"] = &schema.Schema{
		Description:  clientSecretWOVDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"client_secret_wo"},
	}
	s["public_key"] = &schema.Schema{
		Description:   publicKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"client_id", "client_secret", "client_secret_wo"},
	}
	s["private_key"] = &schema.Schema{
		Description:   privateKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"private_key_wo", "client_id", "client_secret", "client_secret_wo"},
		RequiredWith:  []string{"public_key"},
	}
	s["private_key_wo"] = &schema.Schema{
		Description:   privateKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"private_key", "client_id", "client_secret", "client_secret_wo"},
		RequiredWith:  []string{"private_key_wo_version", "public_key"},
	}
	s["private_key_wo_version"] = &schema.Schema{
		Description:  privateKeyWOVDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"private_key_wo"},
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
		"group_id": {
			Description: groupIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"db_name": {
			Description: dbNameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"host": {
			Description: hostDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"public_key": {
			Description: publicKeyDesc,
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
