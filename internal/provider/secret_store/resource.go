package secret_store

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Manages a Hush Security secret store, describing where the access-manager materializes secrets for a set of deployments."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: resourceCreate,
		ReadContext:   secretStoreRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:        SecretStoreResourceSchema(),
		CustomizeDiff: customizeDiff,
	}
}

// customizeDiff applies the rules the schema cannot: which fields a vault auth
// method requires. Doing it here rather than at apply is the point -- the API
// refuses these too, and a customer should hear it from the plan.
func customizeDiff(
	_ context.Context, d *schema.ResourceDiff, _ any,
) error {
	v, ok := d.GetOk("hc_vault")
	if !ok {
		return nil
	}
	block := v.([]any)[0].(map[string]any)
	auth, ok := block["auth"].([]any)
	if !ok || len(auth) == 0 || auth[0] == nil {
		return nil
	}
	return validateVaultAuth(auth[0].(map[string]any))
}

// validateVaultAuth requires the field the named method uses and refuses the
// fields belonging to the others. A field of another method is refused rather
// than ignored: a store's config is immutable, so one created with a role it
// never uses cannot be corrected, and the mistake it usually represents is
// naming the wrong method.
func validateVaultAuth(auth map[string]any) error {
	method, _ := auth["method"].(string)
	role, _ := auth["role"].(string)
	mount, _ := auth["mount"].(string)

	if method == client.SecretStoreVaultAuthToken {
		if role != "" {
			return fmt.Errorf("auth method %q does not use role", method)
		}
		if mount != "" {
			return fmt.Errorf("auth method %q does not use mount", method)
		}
		return nil
	}
	if role == "" {
		return fmt.Errorf("auth method %q needs a role", method)
	}
	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	config, err := expandConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	input := &client.CreateSecretStoreInput{
		Name:          d.Get("name").(string),
		DeploymentIDs: expandStringList(d.Get("deployment_ids").([]any)),
		Config:        *config,
	}
	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	store, err := client.CreateSecretStore(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(store.ID)

	return secretStoreRead(ctx, d, m)
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateSecretStoreInput{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
	if d.HasChange("deployment_ids") {
		ids := expandStringList(d.Get("deployment_ids").([]any))
		input.DeploymentIDs = &ids
		hasChanges = true
	}

	if hasChanges {
		if _, err := client.UpdateSecretStore(ctx, c, d.Id(), input); err != nil {
			if errResponse, ok := err.(*client.APIError); ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return secretStoreRead(ctx, d, m)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if err := client.DeleteSecretStore(ctx, c, d.Id()); err != nil {
		if errResponse, ok := err.(*client.APIError); ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// expandConfig reads the single present config block and builds the client config
// with the matching kind. ExactlyOneOf guarantees exactly one block is set.
func expandConfig(d *schema.ResourceData) (*client.SecretStoreConfig, error) {
	if v, ok := d.GetOk("aws_sm"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:     client.SecretStoreKindAWSSM,
			Prefix:   block["prefix"].(string),
			Region:   block["region"].(string),
			KmsKeyID: block["kms_key_id"].(string),
		}, nil
	}
	if v, ok := d.GetOk("aws_ssm"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:     client.SecretStoreKindAWSSSM,
			Prefix:   block["prefix"].(string),
			Region:   block["region"].(string),
			KmsKeyID: block["kms_key_id"].(string),
		}, nil
	}
	if v, ok := d.GetOk("gcp_sm"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:      client.SecretStoreKindGCPSM,
			Prefix:    block["prefix"].(string),
			ProjectID: block["project_id"].(string),
		}, nil
	}
	if v, ok := d.GetOk("k8s_secrets"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:      client.SecretStoreKindK8sSecrets,
			Prefix:    block["prefix"].(string),
			Namespace: block["namespace"].(string),
		}, nil
	}
	if v, ok := d.GetOk("hc_vault"); ok {
		block := v.([]any)[0].(map[string]any)
		// the auth block is Required with MaxItems 1, so exactly one is here
		auth := block["auth"].([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:           client.SecretStoreKindHCVault,
			Prefix:         block["prefix"].(string),
			Address:        block["address"].(string),
			Mount:          block["mount"].(string),
			VaultNamespace: block["vault_namespace"].(string),
			CaCert:         block["ca_cert"].(string),
			Auth: &client.SecretStoreVaultAuth{
				Method: auth["method"].(string),
				Mount:  auth["mount"].(string),
				Role:   auth["role"].(string),
			},
		}, nil
	}
	return nil, fmt.Errorf("one of the config blocks (%s) must be set",
		strings.Join(configBlockNames, ", "))
}

func expandStringList(items []any) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}
