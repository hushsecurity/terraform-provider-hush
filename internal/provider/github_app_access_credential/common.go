package github_app_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc              = "The unique identifier of the GitHub App access credential"
	nameDesc            = "The name of the GitHub App access credential"
	descriptionDesc     = "The description of the GitHub App access credential"
	deploymentIDsDesc   = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	clientIDDesc        = "The GitHub App's client ID, shown on its settings page: 1-64 letters, digits and dots. Older `Iv1.<hex>` client IDs are accepted, as is the numeric App ID"
	installationIDDesc  = "The ID of the App's installation on the account whose repositories the tokens reach: a positive integer of at most 19 digits, without leading zeros"
	privateKeyDesc      = "The GitHub App's private key: the unencrypted RSA key in PEM format, as downloaded from the App's settings page"
	privateKeyWODesc    = "The GitHub App's private key (write-only): the unencrypted RSA key in PEM format. More secure than `private_key` because Terraform will not store this value in the state file."
	privateKeyWOVerDesc = "Used to trigger updates for `private_key_wo`. Change when the private key changes."
	typeDesc            = "The type of access credential"
	kindDesc            = "The kind of access credential"
	secretStoreIDDesc   = "The ID of the secret store where this credential is saved (optional)"
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
	s["secret_store_id"] = &schema.Schema{
		Description:  secretStoreIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^sst-`), "secret_store_id must start with 'sst-'"),
	}
	s["private_key"] = &schema.Schema{
		Description:   privateKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"private_key_wo"},
		ExactlyOneOf:  []string{"private_key", "private_key_wo"},
	}
	s["private_key_wo"] = &schema.Schema{
		Description:   privateKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"private_key"},
		ExactlyOneOf:  []string{"private_key", "private_key_wo"},
		RequiredWith:  []string{"private_key_wo_version"},
	}
	s["private_key_wo_version"] = &schema.Schema{
		Description:  privateKeyWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"private_key_wo"},
	}
	s["client_id"] = &schema.Schema{
		Description: clientIDDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringMatch(
			regexp.MustCompile(`^[A-Za-z0-9.]{1,64}$`),
			"client_id must be 1-64 letters, digits or dots"),
	}
	s["installation_id"] = &schema.Schema{
		Description: installationIDDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringMatch(
			regexp.MustCompile(`^[1-9][0-9]{0,18}$`),
			"installation_id must be a positive integer of at most 19 digits, without leading zeros"),
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
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"installation_id": {
			Description: installationIDDesc,
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
		"secret_store_id": {
			Description: secretStoreIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
