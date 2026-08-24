package secret_store

import (
	"context"
	"fmt"
	"net/http"
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

	prefixBaseDesc   = "Namespace prefix for secrets in the backend store. Lowercase, and each segment must start and end with a letter or digit. At most 80 characters, enforced by the API."
	awsSMPrefixDesc  = prefixBaseDesc + " AWS Secrets Manager also allows _ . + = @ and \"/\" between segments."
	awsSSMPrefixDesc = prefixBaseDesc + " AWS SSM Parameter Store also allows _ . and \"/\" between segments, may not start with \"aws\" or \"ssm\", and is limited to 9 \"/\"-separated segments."
	gcpSMPrefixDesc  = prefixBaseDesc + " GCP Secret Manager also allows _."
	k8sPrefixDesc    = prefixBaseDesc + " A Kubernetes Secret name also allows ., and no other punctuation."

	regionDesc    = "The cloud region of the backend store"
	kmsKeyIDDesc  = "The KMS key used to encrypt secrets (optional)"
	projectIDDesc = "The GCP project that hosts the backend store"
	namespaceDesc = "The Kubernetes namespace for the secrets (defaults to the access-manager install namespace when omitted)"

	awsSMDesc  = "Configuration for an AWS Secrets Manager backend. Immutable: changing it forces a new secret store."
	awsSSMDesc = "Configuration for an AWS SSM Parameter Store backend. Immutable: changing it forces a new secret store."
	gcpSMDesc  = "Configuration for a GCP Secret Manager backend. Immutable: changing it forces a new secret store."
	k8sDesc    = "Configuration for a Kubernetes Secrets backend. Immutable: changing it forces a new secret store."
)

// 15 Parameter Store hierarchy levels less the six a remote key appends.
const awsSSMMaxSegments = 9

var configBlockNames = []string{"aws_sm", "aws_ssm", "gcp_sm", "k8s_secrets"}

// Charset and shape only. The length maximum is one figure for every kind and
// the API enforces it, so it is described in the attribute rather than checked
// here. Each pattern allows runs of lowercase alphanumerics separated by single
// punctuation, so a prefix never starts or ends with punctuation and never
// repeats it.
var (
	awsSMPrefixValidation = prefixCharsetValidation("-_.+=@", "/", true,
		"lowercase alphanumerics separated by - _ . + = @ or /")
	awsSSMPrefixValidation = validation.All(
		prefixCharsetValidation("-_.", "/", true,
			"lowercase alphanumerics separated by - _ . or /"),
		reservedPrefixValidation("aws", "ssm"),
		segmentCountValidation("/", awsSSMMaxSegments),
	)
	gcpSMPrefixValidation = prefixCharsetValidation("_", "-", true,
		"lowercase alphanumerics separated by - or _")
	k8sPrefixValidation = prefixCharsetValidation(".", "-", false,
		"lowercase alphanumerics separated by single - or .")
)

// The separator is kept single whatever mayRepeat says: it delimits the
// segments the API counts, and a doubled one would leave an empty segment --
// Parameter Store rejects that outright. Other punctuation may be adjacent
// where the backend attaches no meaning to it. k8s_secrets is the
// exception, and keeps everything single: a Secret name is a DNS subdomain,
// where "." delimits labels and a label may not start or end with "-".
func prefixCharsetValidation(
	punctuation, separator string, mayRepeat bool, describe string,
) func(any, string) ([]string, []error) {
	group := regexp.QuoteMeta(punctuation + separator)
	if mayRepeat {
		group = "[" + regexp.QuoteMeta(punctuation) + "]+|" +
			regexp.QuoteMeta(separator)
	} else {
		group = "[" + group + "]"
	}
	pattern := regexp.MustCompile(`^[a-z0-9]+((` + group + `)[a-z0-9]+)*$`)
	return validation.StringMatch(pattern, "prefix must be "+describe)
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
