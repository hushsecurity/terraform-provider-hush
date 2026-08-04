package redis_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	engineRedis             = "redis"
	engineElastiCache       = "elasticache"
	engineAiven             = "aiven"
	engineAzureManagedRedis = "azure_managed_redis"
)

var (
	validEngines      = []string{engineRedis, engineElastiCache, engineAiven, engineAzureManagedRedis}
	validCacheEngines = []string{"redis", "valkey"}
)

var (
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	// ARM naming rules for the two Azure locators, mirroring midgard's
	// AzureResourceGroup and AzureRedisClusterName.
	resourceGroupRegex = regexp.MustCompile(
		`^[-\p{L}\p{Nl}\p{M}\p{Nd}\p{Pc}\x{200C}\x{200D}.()]*[-\p{L}\p{Nl}\p{M}\p{Nd}\p{Pc}\x{200C}\x{200D}()]$`)
	clusterNameRegex = regexp.MustCompile(`^[A-Za-z0-9]+(-[A-Za-z0-9]+)*$`)
)

const (
	idDesc                 = "The unique identifier of the Redis access credential"
	nameDesc               = "The name of the Redis access credential"
	descriptionDesc        = "The description of the Redis access credential"
	deploymentIDsDesc      = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	hostDesc               = "The hostname or IP address of the Redis server. Required when `engine` is `redis` or `elasticache`; must not be set when `engine` is `aiven` or `azure_managed_redis` (Hush resolves the endpoint from the provider's API)."
	portDesc               = "The port number of the Redis server (default: 6379). Only valid when `engine` is `redis` or `elasticache`."
	usernameDesc           = "The username for the Redis connection (Redis 6+ ACL). Only valid when `engine` is `redis` or `elasticache`."
	passwordDesc           = "The password for the Redis connection. Required when `engine` is `redis`; must not be set for any other engine."
	passwordWODesc         = "The password for the Redis connection (write-only). This is a write-only attribute that is more secure than `password` because Terraform will not store this value in the state file. Required when `engine` is `redis`; must not be set for any other engine."
	passwordWOVerDesc      = "Used to trigger updates for `password_wo`. This value should be changed when the password content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	databaseDesc           = "The Redis database number (0-15, default: 0). Only valid when `engine` is `redis` or `elasticache`."
	tlsDesc                = "Whether to use TLS for the Redis connection. Only valid when `engine` is `redis` or `elasticache`."
	tlsCADesc              = "The TLS CA certificate for the Redis connection. Only valid when `engine` is `redis` or `elasticache`."
	engineDesc             = "The routing engine for this credential. `redis` connects directly to a Redis server using a password. `elasticache` provisions users via the AWS ElastiCache API. `aiven` provisions users via the Aiven API for an Aiven-managed Valkey service. `azure_managed_redis` provisions Entra ID service principals for an Azure Managed Redis cluster via the Azure APIs. Immutable; changing it forces replacement."
	cacheEngineDesc        = "The AWS ElastiCache cache engine. Required and only valid when `engine` is `elasticache`. One of `redis`, `valkey`. Not valid for the `aiven` or `azure_managed_redis` engines (Hush resolves the variant from the live service)."
	regionDesc             = "The AWS region of the ElastiCache cluster. Required and only valid when `engine` is `elasticache`."
	userGroupIDDesc        = "The ElastiCache user group ID to add provisioned users to. Required and only valid when `engine` is `elasticache`."
	accessKeyIDDesc        = "The AWS access key ID used to call the ElastiCache API. Only valid when `engine` is `elasticache`. Must be set together with `secret_access_key` (or `secret_access_key_wo`); omit both to use AWS workload identity federation (IRSA / instance profile / WIF)."
	secretAccessKeyDesc    = "The AWS secret access key used to call the ElastiCache API. Only valid when `engine` is `elasticache`. Must be set together with `access_key_id`. Omit both to use AWS workload identity federation."
	secretAccessKeyWODesc  = "The AWS secret access key (write-only). This is a write-only attribute that is more secure than `secret_access_key` because Terraform will not store this value in the state file."
	secretAccessKeyWOVDesc = "Used to trigger updates for `secret_access_key_wo`. This value should be changed when the secret content changes."
	projectDesc            = "The Aiven project that owns the Valkey service. Required when `engine` is `aiven`."
	serviceNameDesc        = "The Aiven Valkey service name. Required when `engine` is `aiven`."
	tokenDesc              = "The Aiven API token used to manage the service (required when `engine` is `aiven`)."
	tokenWODesc            = "The Aiven API token (write-only). This is a write-only attribute that is more secure than `token` because Terraform will not store this value in the state file. Used when `engine` is `aiven`."
	tokenWOVerDesc         = "Used to trigger updates for `token_wo`. This value should be changed when the token content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	tenantIDDesc           = "The Azure tenant ID (lowercase UUID) of the directory that owns the Managed Redis cluster. Required and only valid when `engine` is `azure_managed_redis`."
	clientIDDesc           = "The client ID of the Azure application Hush uses to manage the cluster's Entra ID users. Only valid when `engine` is `azure_managed_redis`. Must be set together with `client_secret` (or `client_secret_wo`); omit both to use the access-manager's default Azure credentials (managed identity / workload identity)."
	clientSecretDesc       = "The client secret of the Azure application identified by `client_id`. Only valid when `engine` is `azure_managed_redis`. Must be set together with `client_id`."
	clientSecretWODesc     = "The client secret of the Azure application identified by `client_id` (write-only). This is a write-only attribute that is more secure than `client_secret` because Terraform will not store this value in the state file."
	clientSecretWOVDesc    = "Used to trigger updates for `client_secret_wo`. This value should be changed when the client secret content changes. Can be any value (e.g., a timestamp, version number, or hash)."
	subscriptionIDDesc     = "The Azure subscription ID (lowercase UUID) that contains the Managed Redis cluster. Required and only valid when `engine` is `azure_managed_redis`."
	resourceGroupDesc      = "The Azure resource group that contains the Managed Redis cluster. Required and only valid when `engine` is `azure_managed_redis`."
	clusterNameDesc        = "The name of the Azure Managed Redis cluster. Required and only valid when `engine` is `azure_managed_redis`."
	typeDesc               = "The type of access credential"
	kindDesc               = "The kind of access credential"
	secretStoreIDDesc      = "The ID of the secret store where this credential is saved (optional)"
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
	// host is Optional (not Required) because the aiven engine must not set it;
	// per-engine requiredness is enforced in CustomizeDiff (validateEngineFields).
	s["host"] = &schema.Schema{
		Description: hostDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["port"] = &schema.Schema{
		Description: portDesc,
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     6379,
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
	s["database"] = &schema.Schema{
		Description:  databaseDesc,
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      0,
		ValidateFunc: validation.IntBetween(0, 15),
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
	s["engine"] = &schema.Schema{
		Description:  engineDesc,
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice(validEngines, false),
	}
	s["cache_engine"] = &schema.Schema{
		Description:  cacheEngineDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(validCacheEngines, false),
	}
	s["region"] = &schema.Schema{
		Description: regionDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["user_group_id"] = &schema.Schema{
		Description: userGroupIDDesc,
		Type:        schema.TypeString,
		Optional:    true,
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
		Description:  secretAccessKeyWOVDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"secret_access_key_wo"},
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
	// Azure Managed Redis engine fields. The client_id/client_secret pair is
	// enforced in CustomizeDiff, not via RequiredWith, so the write-only secret
	// can stand in for the plain one.
	s["tenant_id"] = &schema.Schema{
		Description:  tenantIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(uuidRegex, "tenant_id must be a lowercase UUID"),
	}
	s["client_id"] = &schema.Schema{
		Description:  clientIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
	}
	s["client_secret"] = &schema.Schema{
		Description:   clientSecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"client_secret_wo"},
	}
	s["client_secret_wo"] = &schema.Schema{
		Description:   clientSecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"client_secret"},
		RequiredWith:  []string{"client_secret_wo_version"},
	}
	s["client_secret_wo_version"] = &schema.Schema{
		Description:  clientSecretWOVDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"client_secret_wo"},
	}
	s["subscription_id"] = &schema.Schema{
		Description:  subscriptionIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(uuidRegex, "subscription_id must be a lowercase UUID"),
	}
	s["resource_group"] = &schema.Schema{
		Description: resourceGroupDesc,
		Type:        schema.TypeString,
		Optional:    true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 90),
			validation.StringMatch(resourceGroupRegex, "resource_group must contain only letters, digits, '-', '_', '.', '(' or ')' and must not end with a period"),
		),
	}
	s["cluster_name"] = &schema.Schema{
		Description: clusterNameDesc,
		Type:        schema.TypeString,
		Optional:    true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 60),
			validation.StringMatch(clusterNameRegex, "cluster_name must be alphanumeric segments separated by single hyphens"),
		),
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
		"username": {
			Description: usernameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"database": {
			Description: databaseDesc,
			Type:        schema.TypeInt,
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
		"engine": {
			Description: engineDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cache_engine": {
			Description: cacheEngineDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"region": {
			Description: regionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"user_group_id": {
			Description: userGroupIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"access_key_id": {
			Description: accessKeyIDDesc,
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
		"tenant_id": {
			Description: tenantIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"subscription_id": {
			Description: subscriptionIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"resource_group": {
			Description: resourceGroupDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cluster_name": {
			Description: clusterNameDesc,
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
