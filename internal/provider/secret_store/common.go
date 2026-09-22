package secret_store

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc            = "The unique identifier of the secret store"
	nameDesc          = "The name of the secret store"
	descriptionDesc   = "The description of the secret store"
	deploymentIDsDesc = "List of deployment IDs this secret store is associated with"
	statusDesc        = "The aggregate status of the secret store across its deployments (pending, ready, warning, error)"
	statusDetailDesc  = "Detail of the worst deployment status"

	prefixBaseDesc   = "Namespace prefix for secrets in the backend store. Lowercase, and each segment must start and end with a letter or digit. \"-\" is allowed for every kind and \".\" may not repeat."
	prefix80Desc     = prefixBaseDesc + " At most 80 characters."
	awsSMPrefixDesc  = prefix80Desc + " AWS Secrets Manager also allows \"_\" \".\" \"+\" \"=\" \"@\" and \"/\", with \"/\" separating segments."
	awsSSMPrefixDesc = prefix80Desc + " AWS SSM Parameter Store also allows \"_\" \".\" and \"/\", with \"/\" separating segments, may not start with \"aws\" or \"ssm\", and is limited to 9 \"/\"-separated segments."
	gcpSMPrefixDesc  = prefix80Desc + " GCP Secret Manager also allows \"_\", with \"-\" separating segments."
	k8sPrefixDesc    = prefixBaseDesc + " A Kubernetes Secret name also allows \".\", with \"-\" separating segments, and no other punctuation."

	hcVaultPrefixDesc = prefix80Desc + " A Vault KV v2 path also allows _ . and \"/\" between segments."
	// 32, not 80: a Key Vault secret name is capped at 127 and the namespace
	// and key take the rest. Mirrors am's KindLimits.MaxUserPrefix().
	azureKvPrefixDesc = prefixBaseDesc + " A Key Vault secret name allows no other punctuation, with \"-\" separating segments, and at most 32 characters -- the backend caps a name at 127 and the rest is taken by what the access manager appends."

	regionDesc    = "The cloud region of the backend store"
	kmsKeyIDDesc  = "The KMS key used to encrypt secrets (optional)"
	projectIDDesc = "The GCP project that hosts the backend store"
	namespaceDesc = "The Kubernetes namespace for the secrets (defaults to the access-manager install namespace when omitted)"

	awsSMDesc  = "Configuration for an AWS Secrets Manager backend. Immutable: changing it forces a new secret store."
	awsSSMDesc = "Configuration for an AWS SSM Parameter Store backend. Immutable: changing it forces a new secret store."
	gcpSMDesc  = "Configuration for a GCP Secret Manager backend. Immutable: changing it forces a new secret store."
	k8sDesc    = "Configuration for a Kubernetes Secrets backend. Immutable: changing it forces a new secret store."
	vaultDesc  = "Configuration for a HashiCorp Vault backend, on its KV v2 secrets engine. Immutable: changing it forces a new secret store."
	azureDesc  = "Configuration for an Azure Key Vault backend. Immutable: changing it forces a new secret store."

	vaultURLDesc = "The https address of the Key Vault, e.g. \"https://acme.vault.azure.net\"."
	cloudDesc    = "The Azure cloud the vault and its Entra ID authority live in: \"public\" (the default), \"china\" or \"usgov\". It selects the authority as well as the data-plane audience, so it cannot be inferred from the URL."

	azureAuthDesc       = "How the access manager authenticates to Key Vault"
	azureAuthMethodDesc = "The auth method: \"default\" (the access manager's own identity -- on AKS, workload identity) or \"client_secret\" (a service principal). Defaults to \"default\"."
	azureTenantDesc     = "The Entra ID tenant. Required by both methods."
	azureClientIDDesc   = "The service principal's application id, for the \"client_secret\" method. The \"default\" method takes none: the SDK has no field for one and reads AZURE_CLIENT_ID from the environment, which the workload identity webhook sets, so one given here is refused rather than silently ignored."

	addressDesc = "The https address of the Vault server. Plaintext is refused: every request carries the Vault token in a header, and a login posts the access manager's service-account token in a body."
	mountDesc   = "The KV v2 mount the secrets live under (defaults to \"secret\" when omitted)"
	vaultNsDesc = "A Vault Enterprise namespace to address the secrets in (optional). Unrelated to the prefix, which names the secrets themselves; Vault OSS ignores it."
	caCertDesc  = "The PEM certificate the Vault server's certificate is verified against (optional). Needed only when it does not chain to a root the access manager already trusts."

	authDesc       = "How the access manager authenticates to Vault"
	authMethodDesc = "The Vault auth method: \"kubernetes\" (the access manager presents its pod's service-account token and the cluster's TokenReview api vouches for it), \"jwt\" (the same token, validated against the cluster's JWKS, for a Vault that cannot reach the api server), or \"token\" (a token the deployment already holds). Defaults to \"kubernetes\"."
	authMountDesc  = "The path the auth method is mounted at (defaults to the method's own name when omitted). Not used by the \"token\" method, which does not log in."
	authRoleDesc   = "The Vault role the access manager's service account is bound to. Required by the \"kubernetes\" and \"jwt\" methods, and not used by \"token\"."
)

// 15 Parameter Store hierarchy levels less the six a remote key appends.
const awsSSMMaxSegments = 9

var configBlockNames = []string{
	"aws_sm", "aws_ssm", "gcp_sm", "k8s_secrets", "hc_vault", "azure_kv",
}

// The prefix maxima, mirroring am's KindLimits.MaxUserPrefix(). It stopped
// being one figure with azure_kv, whose backend caps a secret name at 127 and
// so leaves 32 once the namespace and a maximum key are accounted for. Checked
// here rather than left to the API because a tightening is otherwise missed
// until apply.
const (
	maxPrefixLength        = 80
	azureKvMaxPrefixLength = 32
)

// Charset and shape only, plus the length above. Formerly charset only, when
// the API enforces it, so it is described in the attribute rather than checked
// here. Each pattern allows runs of lowercase alphanumerics separated by
// punctuation, so a prefix never starts or ends with punctuation.
var (
	awsSMPrefixValidation = prefixCharsetValidation("-_.+=@", "/",
		"lowercase alphanumerics separated by - _ . + = @ or /")
	awsSSMPrefixValidation = validation.All(
		prefixCharsetValidation("-_.", "/",
			"lowercase alphanumerics separated by - _ . or /"),
		reservedPrefixValidation("aws", "ssm"),
		segmentCountValidation("/", awsSSMMaxSegments),
	)
	gcpSMPrefixValidation = prefixCharsetValidation("_", "-",
		"lowercase alphanumerics separated by - or _")
	k8sPrefixValidation = prefixCharsetValidation(".", "-",
		"lowercase alphanumerics separated by single - or .")
	// A KV v2 path. Nothing is reserved: the silo writes under
	// <mount>/data/<prefix>/..., where "data" and "metadata" are not special.
	hcVaultPrefixValidation = prefixCharsetValidation("-_.", "/",
		"lowercase alphanumerics separated by - _ . or /")
	// A Key Vault secret name admits [a-zA-Z0-9-] alone, so "-" separates and
	// nothing else is allowed inside a segment.
	azureKvPrefixValidation = validation.All(
		prefixCharsetValidation("", "-",
			"lowercase alphanumerics separated by -"),
		validation.StringLenBetween(1, azureKvMaxPrefixLength),
	)
)

// httpsURLValidation refuses at plan time what the API refuses on the request:
// a backend URL that is not https. Mirrors midgard's
// validate_secret_store_config, which applies it to Vault's address and Key
// Vault's vault_url alike; Go and python lower the scheme the same way, so one
// typed in capitals is accepted by both.
func httpsURLValidation(i any, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("%s: expected a string", k)}
	}
	parsed, err := url.Parse(v)
	if err != nil {
		return nil, []error{fmt.Errorf("%s: %q is not a URL: %w", k, v, err)}
	}
	if parsed.Scheme != "https" {
		return nil, []error{fmt.Errorf("%s: address must be an https URL, got %q", k, v)}
	}
	if parsed.Hostname() == "" {
		return nil, []error{fmt.Errorf("%s: address must have a host, got %q", k, v)}
	}
	return nil, nil
}

// mountValidation keeps a mount inside the request path the access manager
// builds it into -- "<mount>/data/<prefix>/..." for the KV mount and
// "auth/<mount>/login" for the auth one -- so a mount cannot walk out of it.
// Mirrors midgard's VAULT_MOUNT.
var mountValidation = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9_-]+(/[A-Za-z0-9_-]+)*$`),
	"mount must be letters, digits, '-' and '_', with '/' between segments")

// The separator is kept single: it delimits the segments the API counts, and a
// doubled one would leave an empty segment -- Parameter Store rejects that
// outright. Other punctuation may be adjacent, where the backend attaches no
// meaning to it. "." is the exception, for every kind rather than per kind:
// "a..b" is an empty label in a DNS subdomain, which k8s_secrets refuses, and
// no other backend has a use for it. Checked before the pattern so it names
// itself, and only where "." is in the charset, so a kind without it reports
// the charset instead.
func prefixCharsetValidation(
	punctuation, separator, describe string,
) func(any, string) ([]string, []error) {
	// azure_kv allows no punctuation but the separator, and an empty character
	// class is not a valid expression, so the alternation collapses to it
	group := regexp.QuoteMeta(separator)
	if punctuation != "" {
		group = "[" + regexp.QuoteMeta(punctuation) + "]+|" + group
	}
	pattern := regexp.MustCompile(`^[a-z0-9]+((` + group + `)[a-z0-9]+)*$`)
	match := validation.StringMatch(pattern, "prefix must be "+describe)
	checkDot := strings.Contains(punctuation, ".")
	return func(i any, k string) ([]string, []error) {
		if v, ok := i.(string); ok && checkDot && strings.Contains(v, "..") {
			return nil, []error{fmt.Errorf("%s: prefix must not repeat '.'", k)}
		}
		return match(i, k)
	}
}

// AWS SSM will not accept a name beginning with "aws" or "ssm", and the prefix
// is a remote key's first path element, so such a store fails every write.
func reservedPrefixValidation(
	reserved ...string,
) func(any, string) ([]string, []error) {
	return func(i any, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("%s: expected a string", k)}
		}
		for _, r := range reserved {
			if strings.HasPrefix(strings.ToLower(v), r) {
				return nil, []error{fmt.Errorf(
					"%s: prefix must not start with %q, which the backend reserves",
					k, r)}
			}
		}
		return nil, nil
	}
}

// Unlike length, this is a restriction, so it belongs at plan time.
func segmentCountValidation(
	separator string, max int,
) func(any, string) ([]string, []error) {
	return func(i any, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("%s: expected a string", k)}
		}
		if n := len(strings.Split(v, separator)); n > max {
			return nil, []error{fmt.Errorf(
				"%s: prefix must have at most %d %q-separated segments, got %d",
				k, max, separator, n)}
		}
		return nil, nil
	}
}

func SecretStoreResourceSchema() map[string]*schema.Schema {
	s := SecretStoreDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 60),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 200),
	}
	s["deployment_ids"] = &schema.Schema{
		Description: deploymentIDsDesc,
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}

	s["aws_sm"] = resourceConfigBlock(awsSMDesc,
		awsConfigResource(awsSMPrefixDesc, awsSMPrefixValidation))
	s["aws_ssm"] = resourceConfigBlock(awsSSMDesc,
		awsConfigResource(awsSSMPrefixDesc, awsSSMPrefixValidation))
	s["gcp_sm"] = resourceConfigBlock(gcpSMDesc, gcpConfigResource())
	s["k8s_secrets"] = resourceConfigBlock(k8sDesc, k8sConfigResource())
	s["hc_vault"] = resourceConfigBlock(vaultDesc, hcVaultConfigResource())
	s["azure_kv"] = resourceConfigBlock(azureDesc, azureKvConfigResource())

	return s
}

func SecretStoreDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description:   idDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"id"},
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
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status_detail": {
			Description: statusDetailDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"aws_sm":      dataSourceConfigBlock(awsSMDesc, awsConfigDataSource()),
		"aws_ssm":     dataSourceConfigBlock(awsSSMDesc, awsConfigDataSource()),
		"gcp_sm":      dataSourceConfigBlock(gcpSMDesc, gcpConfigDataSource()),
		"k8s_secrets": dataSourceConfigBlock(k8sDesc, k8sConfigDataSource()),
		"hc_vault":    dataSourceConfigBlock(vaultDesc, hcVaultConfigDataSource()),
		"azure_kv":    dataSourceConfigBlock(azureDesc, azureKvConfigDataSource()),
	}
}

func resourceConfigBlock(description string, elem *schema.Resource) *schema.Schema {
	return &schema.Schema{
		Description:  description,
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		MaxItems:     1,
		Elem:         elem,
		ExactlyOneOf: configBlockNames,
	}
}

func dataSourceConfigBlock(description string, elem *schema.Resource) *schema.Schema {
	return &schema.Schema{
		Description: description,
		Type:        schema.TypeList,
		Computed:    true,
		Elem:        elem,
	}
}

func awsConfigResource(
	prefixDescription string,
	prefixValidation func(any, string) ([]string, []error),
) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  prefixDescription,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: prefixValidation,
			},
			"region": {
				Description:  regionDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"kms_key_id": {
				Description: kmsKeyIDDesc,
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func gcpConfigResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  gcpSMPrefixDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: gcpSMPrefixValidation,
			},
			"project_id": {
				Description:  projectIDDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func k8sConfigResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  k8sPrefixDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: k8sPrefixValidation,
			},
			"namespace": {
				Description: namespaceDesc,
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func hcVaultConfigResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  hcVaultPrefixDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: hcVaultPrefixValidation,
			},
			"address": {
				Description:  addressDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: httpsURLValidation,
			},
			"mount": {
				Description:  mountDesc,
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: mountValidation,
			},
			"vault_namespace": {
				Description: vaultNsDesc,
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"ca_cert": {
				Description: caCertDesc,
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"auth": {
				Description: authDesc,
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Description: authMethodDesc,
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Default:     client.SecretStoreVaultAuthKubernetes,
							ValidateFunc: validation.StringInSlice(
								client.SecretStoreVaultAuthMethods, false),
						},
						"mount": {
							Description:  authMountDesc,
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: mountValidation,
						},
						// Required by method, which the schema cannot express,
						// so both of these are optional here and the diff
						// requires the one the method uses.
						"role": {
							Description:  authRoleDesc,
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
		},
	}
}

func azureKvConfigResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  azureKvPrefixDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azureKvPrefixValidation,
			},
			"vault_url": {
				Description:  vaultURLDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: httpsURLValidation,
			},
			"cloud": {
				Description: cloudDesc,
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice(
					[]string{"public", "china", "usgov"}, false),
			},
			"auth": {
				Description: azureAuthDesc,
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Description: azureAuthMethodDesc,
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Default:     client.SecretStoreAzureAuthDefault,
							ValidateFunc: validation.StringInSlice(
								client.SecretStoreAzureAuthMethods, false),
						},
						"tenant_id": {
							Description:  azureTenantDesc,
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						// required by method, which the schema cannot express,
						// so the diff requires it where it belongs
						"client_id": {
							Description:  azureClientIDDesc,
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
		},
	}
}

func azureKvConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":    {Description: prefixBaseDesc, Type: schema.TypeString, Computed: true},
			"vault_url": {Description: vaultURLDesc, Type: schema.TypeString, Computed: true},
			"cloud":     {Description: cloudDesc, Type: schema.TypeString, Computed: true},
			"auth": {
				Description: azureAuthDesc,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method":    {Description: azureAuthMethodDesc, Type: schema.TypeString, Computed: true},
						"tenant_id": {Description: azureTenantDesc, Type: schema.TypeString, Computed: true},
						"client_id": {Description: azureClientIDDesc, Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func awsConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":     {Description: prefixBaseDesc, Type: schema.TypeString, Computed: true},
			"region":     {Description: regionDesc, Type: schema.TypeString, Computed: true},
			"kms_key_id": {Description: kmsKeyIDDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

func gcpConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":     {Description: prefixBaseDesc, Type: schema.TypeString, Computed: true},
			"project_id": {Description: projectIDDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

func k8sConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":    {Description: prefixBaseDesc, Type: schema.TypeString, Computed: true},
			"namespace": {Description: namespaceDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

func hcVaultConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":          {Description: prefixBaseDesc, Type: schema.TypeString, Computed: true},
			"address":         {Description: addressDesc, Type: schema.TypeString, Computed: true},
			"mount":           {Description: mountDesc, Type: schema.TypeString, Computed: true},
			"vault_namespace": {Description: vaultNsDesc, Type: schema.TypeString, Computed: true},
			"ca_cert":         {Description: caCertDesc, Type: schema.TypeString, Computed: true},
			"auth": {
				Description: authDesc,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {Description: authMethodDesc, Type: schema.TypeString, Computed: true},
						"mount":  {Description: authMountDesc, Type: schema.TypeString, Computed: true},
						"role":   {Description: authRoleDesc, Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func secretStoreRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	var store *client.SecretStore
	var err error

	if id := d.Id(); id != "" {
		store, err = client.GetSecretStore(ctx, c, id)
		if err != nil {
			if errResponse, ok := err.(*client.APIError); ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	} else if id, exists := d.GetOk("id"); exists {
		storeID := id.(string)
		store, err = client.GetSecretStore(ctx, c, storeID)
		if err != nil {
			if errResponse, ok := err.(*client.APIError); ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no secret store found with ID: %s", storeID)
			}
			return diag.FromErr(err)
		}
	} else if name, exists := d.GetOk("name"); exists {
		storeName := name.(string)
		stores, lookupErr := client.GetSecretStoresByName(ctx, c, storeName)
		if lookupErr != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup secret store by name '%s': %w", storeName, lookupErr))
		}
		if len(stores) == 0 {
			return diag.Errorf("no secret store found with name: %s", storeName)
		}
		if len(stores) > 1 {
			return diag.Errorf("multiple secret stores found with name: %s, please use id instead", storeName)
		}
		store = &stores[0]
	} else {
		return diag.Errorf("one of `id` or `name` must be specified")
	}

	d.SetId(store.ID)

	fields := map[string]any{
		"name":           store.Name,
		"description":    store.Description,
		"deployment_ids": store.DeploymentIDs,
		"status":         store.Status,
		"status_detail":  store.StatusDetail,
	}
	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if err := setConfigBlocks(d, &store.Config); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// setConfigBlocks writes the one config block matching the store's kind and clears
// the others, so state reflects exactly the active backend configuration.
func setConfigBlocks(d *schema.ResourceData, config *client.SecretStoreConfig) error {
	block := map[string]any{"prefix": config.Prefix}
	switch config.Kind {
	case client.SecretStoreKindAWSSM, client.SecretStoreKindAWSSSM:
		block["region"] = config.Region
		block["kms_key_id"] = config.KmsKeyID
	case client.SecretStoreKindGCPSM:
		block["project_id"] = config.ProjectID
	case client.SecretStoreKindK8sSecrets:
		block["namespace"] = config.Namespace
	case client.SecretStoreKindAzureKv:
		block["vault_url"] = config.VaultURL
		block["cloud"] = config.Cloud
		auth := map[string]any{}
		if config.Auth != nil {
			auth["method"] = config.Auth.Method
			auth["tenant_id"] = config.Auth.TenantID
			auth["client_id"] = config.Auth.ClientID
		}
		block["auth"] = []map[string]any{auth}
	case client.SecretStoreKindHCVault:
		block["address"] = config.Address
		block["mount"] = config.Mount
		block["vault_namespace"] = config.VaultNamespace
		block["ca_cert"] = config.CaCert
		auth := map[string]any{}
		if config.Auth != nil {
			auth["method"] = config.Auth.Method
			auth["mount"] = config.Auth.Mount
			auth["role"] = config.Auth.Role
		}
		block["auth"] = []map[string]any{auth}
	default:
		return fmt.Errorf("unknown secret store kind: %s", config.Kind)
	}

	for _, name := range configBlockNames {
		value := []map[string]any{}
		if name == config.Kind {
			value = []map[string]any{block}
		}
		if err := d.Set(name, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}
