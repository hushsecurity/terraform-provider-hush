package snowflake_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the Snowflake access credential"
	nameDesc          = "The name of the Snowflake access credential"
	descriptionDesc   = "The description of the Snowflake access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential"
	accountDesc       = "The Snowflake account identifier (e.g., MYORG-MYACCOUNT)"
	warehouseDesc     = "The Snowflake warehouse name"
	databaseDesc      = "The Snowflake database name"
	schemaDesc        = "The Snowflake schema name (default: PUBLIC)"
	roleDesc          = "The Snowflake role name for the root connection"
	usernameDesc      = "The username for the Snowflake connection"
	passwordDesc      = "The password for the Snowflake connection (required when auth_method is 'password')"
	passwordWODesc    = "The password for the Snowflake connection (write-only). This is a write-only attribute that is more secure than `password` because Terraform will not store this value in the state file."
	passwordWOVerDesc = "Used to trigger updates for `password_wo`. This value should be changed when the password content changes."
	privateKeyDesc    = "The PEM-encoded private key for the Snowflake connection (required when auth_method is 'key-pair')"
	privateKeyWODesc  = "The PEM-encoded private key for the Snowflake connection (write-only). This is a write-only attribute that is more secure than `private_key` because Terraform will not store this value in the state file."
	privateKeyWOVDesc = "Used to trigger updates for `private_key_wo`. This value should be changed when the private key content changes."
	authMethodDesc    = "The authentication method for the Snowflake connection (password or key-pair)"
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
	s["account"] = &schema.Schema{
		Description:  accountDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["warehouse"] = &schema.Schema{
		Description:  warehouseDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["database"] = &schema.Schema{
		Description:  databaseDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["schema"] = &schema.Schema{
		Description:  schemaDesc,
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "PUBLIC",
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["role"] = &schema.Schema{
		Description:  roleDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["username"] = &schema.Schema{
		Description:  usernameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
	}
	s["password"] = &schema.Schema{
		Description:   passwordDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"password_wo"},
	}
	s["password_wo"] = &schema.Schema{
		Description:   passwordWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"password"},
		RequiredWith:  []string{"password_wo_version"},
	}
	s["password_wo_version"] = &schema.Schema{
		Description:  passwordWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"password_wo"},
	}
	s["private_key"] = &schema.Schema{
		Description:   privateKeyDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"private_key_wo"},
	}
	s["private_key_wo"] = &schema.Schema{
		Description:   privateKeyWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"private_key"},
		RequiredWith:  []string{"private_key_wo_version"},
	}
	s["private_key_wo_version"] = &schema.Schema{
		Description:  privateKeyWOVDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"private_key_wo"},
	}
	s["auth_method"] = &schema.Schema{
		Description:  authMethodDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"password", "key-pair"}, false),
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
		"account": {
			Description: accountDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"warehouse": {
			Description: warehouseDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"database": {
			Description: databaseDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"schema": {
			Description: schemaDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role": {
			Description: roleDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"username": {
			Description: usernameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"auth_method": {
			Description: authMethodDesc,
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
