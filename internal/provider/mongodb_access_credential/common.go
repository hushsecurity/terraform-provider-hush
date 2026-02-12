package mongodb_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc            = "The unique identifier of the MongoDB access credential"
	nameDesc          = "The name of the MongoDB access credential"
	descriptionDesc   = "The description of the MongoDB access credential"
	deploymentIDsDesc = "List of deployment IDs that can access this credential"
	dbNameDesc        = "The name of the MongoDB database"
	hostDesc          = "The hostname or IP address of the MongoDB server"
	portDesc          = "The port number of the MongoDB server (default: 27017)"
	usernameDesc      = "The username for the MongoDB connection"
	passwordDesc      = "The password for the MongoDB connection"
	passwordWODesc    = "The password for the MongoDB connection (write-only). This is a write-only attribute that is more secure than `password` because Terraform will not store this value in the state file. Either `password` or `password_wo` must be specified."
	passwordWOVerDesc = "Used to trigger updates for `password_wo`. This value should be changed when the password content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	authSourceDesc    = "The authentication source database (default: admin)"
	tlsDesc           = "Whether to use TLS for the MongoDB connection (default: false)"
	tlsCADesc         = "The TLS CA certificate for the MongoDB connection"
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
	s["port"] = &schema.Schema{
		Description: portDesc,
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     27017,
	}
	s["username"] = &schema.Schema{
		Description: usernameDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["password"] = &schema.Schema{
		Description:   passwordDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"password_wo"},
		ExactlyOneOf:  []string{"password", "password_wo"},
	}
	s["password_wo"] = &schema.Schema{
		Description:   passwordWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"password"},
		ExactlyOneOf:  []string{"password", "password_wo"},
		RequiredWith:  []string{"password_wo_version"},
	}
	s["password_wo_version"] = &schema.Schema{
		Description:  passwordWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"password_wo"},
	}
	s["auth_source"] = &schema.Schema{
		Description: authSourceDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "admin",
	}
	s["tls"] = &schema.Schema{
		Description: tlsDesc,
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
	}
	s["tls_ca"] = &schema.Schema{
		Description: tlsCADesc,
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
		"port": {
			Description: portDesc,
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"username": {
			Description: usernameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"auth_source": {
			Description: authSourceDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tls": {
			Description: tlsDesc,
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"tls_ca": {
			Description: tlsCADesc,
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
