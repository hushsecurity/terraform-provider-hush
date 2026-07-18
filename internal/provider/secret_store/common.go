package secret_store

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

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

	prefixDesc    = "Namespace prefix for secrets in the backend store (1-10 chars, lowercase, starting with a letter)"
	regionDesc    = "The cloud region of the backend store"
	kmsKeyIDDesc  = "The KMS key used to encrypt secrets (optional)"
	projectIDDesc = "The GCP project that hosts the backend store"
	namespaceDesc = "The Kubernetes namespace for the secrets (defaults to the access-manager install namespace when omitted)"

	awsSMDesc  = "Configuration for an AWS Secrets Manager backend. Immutable: changing it forces a new secret store."
	awsSSMDesc = "Configuration for an AWS SSM Parameter Store backend. Immutable: changing it forces a new secret store."
	gcpSMDesc  = "Configuration for a GCP Secret Manager backend. Immutable: changing it forces a new secret store."
	k8sDesc    = "Configuration for a Kubernetes Secrets backend. Immutable: changing it forces a new secret store."
)

var configBlockNames = []string{"aws_sm", "aws_ssm", "gcp_sm", "k8s_secrets"}

var prefixValidation = validation.StringMatch(
	regexp.MustCompile(`^[a-z][a-z0-9]{0,9}$`),
	"prefix must be 1-10 characters, lowercase, and start with a letter",
)

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

	s["aws_sm"] = resourceConfigBlock(awsSMDesc, awsConfigResource())
	s["aws_ssm"] = resourceConfigBlock(awsSSMDesc, awsConfigResource())
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

func awsConfigResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  prefixDesc,
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
				Description:  prefixDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: prefixValidation,
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
				Description:  prefixDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: prefixValidation,
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
			"prefix":     {Description: prefixDesc, Type: schema.TypeString, Computed: true},
			"region":     {Description: regionDesc, Type: schema.TypeString, Computed: true},
			"kms_key_id": {Description: kmsKeyIDDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

func gcpConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":     {Description: prefixDesc, Type: schema.TypeString, Computed: true},
			"project_id": {Description: projectIDDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

func k8sConfigDataSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"prefix":    {Description: prefixDesc, Type: schema.TypeString, Computed: true},
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
