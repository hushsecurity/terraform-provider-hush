package secret_store

import (
	"context"
	"fmt"
	"net/http"

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
		Schema: SecretStoreResourceSchema(),
	}
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
	return nil, fmt.Errorf("one of the config blocks (aws_sm, aws_ssm, gcp_sm, k8s_secrets) must be set")
}

func expandStringList(items []any) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}
