package kafka_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	engineNative = "native"
	engineAiven  = "aiven"
)

var (
	validEngines        = []string{engineNative, engineAiven}
	validSaslMechanisms = []string{"PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"}
)

const (
	idDesc               = "The unique identifier of the Kafka access credential"
	nameDesc             = "The name of the Kafka access credential"
	descriptionDesc      = "The description of the Kafka access credential"
	deploymentIDsDesc    = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	engineDesc           = "The Kafka engine: `native` for a self-managed/standard Kafka cluster, or `aiven` for an Aiven-managed service. Immutable; changing it forces replacement."
	bootstrapServersDesc = "Comma-separated list of Kafka bootstrap brokers (host:port,host:port). Required when `engine` is `native`."
	usernameDesc         = "The SASL username for the root Kafka connection. Required when `engine` is `native`."
	saslMechanismDesc    = "The SASL mechanism for the Kafka connection (PLAIN, SCRAM-SHA-256, or SCRAM-SHA-512). Required when `engine` is `native`."
	passwordDesc         = "The SASL password for the Kafka connection (required when `engine` is `native`)."
	passwordWODesc       = "The SASL password for the Kafka connection (write-only). This is a write-only attribute that is more secure than `password` because Terraform will not store this value in the state file. Used when `engine` is `native`."
	passwordWOVerDesc    = "Used to trigger updates for `password_wo`. This value should be changed when the password content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	tlsDesc              = "Whether to use TLS when connecting to the Kafka brokers. Only valid when `engine` is `native`."
	tlsCADesc            = "The TLS CA certificate for the Kafka connection. Only valid when `engine` is `native`."
	projectDesc          = "The Aiven project that owns the Kafka service. Required when `engine` is `aiven`."
	serviceNameDesc      = "The Aiven Kafka service name. Required when `engine` is `aiven`."
	tokenDesc            = "The Aiven API token used to manage the service (required when `engine` is `aiven`)."
	tokenWODesc          = "The Aiven API token (write-only). This is a write-only attribute that is more secure than `token` because Terraform will not store this value in the state file. Used when `engine` is `aiven`."
	tokenWOVerDesc       = "Used to trigger updates for `token_wo`. This value should be changed when the token content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	typeDesc             = "The type of access credential"
	kindDesc             = "The kind of access credential"
	secretStoreIDDesc    = "The ID of the secret store where this credential is saved (optional)"
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
	s["engine"] = &schema.Schema{
		Description:  engineDesc,
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(validEngines, false),
	}
	// Native-engine fields. Optionality is enforced per-engine in CustomizeDiff
	// (validateEngineFields) because the SDK schema cannot express
	// conditional-required attributes.
	s["bootstrap_servers"] = &schema.Schema{
		Description: bootstrapServersDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["username"] = &schema.Schema{
		Description: usernameDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["sasl_mechanism"] = &schema.Schema{
		Description:  saslMechanismDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(validSaslMechanisms, false),
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
	// Aiven-engine fields.
	s["project"] = &schema.Schema{
		Description: projectDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["service_name"] = &schema.Schema{
		Description: serviceNameDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["token"] = &schema.Schema{
		Description:   tokenDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"token_wo"},
	}
	s["token_wo"] = &schema.Schema{
		Description:   tokenWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"token"},
		RequiredWith:  []string{"token_wo_version"},
	}
	s["token_wo_version"] = &schema.Schema{
		Description:  tokenWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"token_wo"},
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
		"engine": {
			Description: engineDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"bootstrap_servers": {
			Description: bootstrapServersDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"username": {
			Description: usernameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"sasl_mechanism": {
			Description: saslMechanismDesc,
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
		"project": {
			Description: projectDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_name": {
			Description: serviceNameDesc,
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
