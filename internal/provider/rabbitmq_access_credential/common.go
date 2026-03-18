package rabbitmq_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc             = "The unique identifier of the RabbitMQ access credential"
	nameDesc           = "The name of the RabbitMQ access credential"
	descriptionDesc    = "The description of the RabbitMQ access credential"
	deploymentIDsDesc  = "List of deployment IDs that can access this credential"
	hostDesc           = "The RabbitMQ host"
	portDesc           = "The RabbitMQ AMQP port (default: 5672)"
	managementPortDesc = "The RabbitMQ management API port (default: 15672)"
	usernameDesc       = "The RabbitMQ username"
	passwordDesc       = "The RabbitMQ password"
	passwordWODesc     = "The RabbitMQ password (write-only). This is a write-only attribute that is more secure than `password` because Terraform will not store this value in the state file."
	passwordWOVerDesc  = "Used to trigger updates for `password_wo`. This value should be changed when the password content changes."
	vhostDesc          = "The RabbitMQ virtual host (default: /)"
	tlsDesc            = "Whether to use TLS"
	tlsCADesc          = "The TLS CA certificate"
	typeDesc           = "The type of access credential"
	kindDesc           = "The kind of access credential"
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
	s["host"] = &schema.Schema{
		Description: hostDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["port"] = &schema.Schema{
		Description:  portDesc,
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      5672,
		ValidateFunc: validation.IntBetween(1, 65535),
	}
	s["management_port"] = &schema.Schema{
		Description:  managementPortDesc,
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      15672,
		ValidateFunc: validation.IntBetween(1, 65535),
	}
	s["username"] = &schema.Schema{
		Description: usernameDesc,
		Type:        schema.TypeString,
		Optional:    true,
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
	s["vhost"] = &schema.Schema{
		Description: vhostDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "/",
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
		Sensitive:   true,
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
		"management_port": {
			Description: managementPortDesc,
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"username": {
			Description: usernameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"vhost": {
			Description: vhostDesc,
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
			Sensitive:   true,
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
