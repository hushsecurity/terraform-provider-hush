package bedrock_access_credential

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/credutil"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Bedrock dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: credutil.ForbidDeploymentIDsChange,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func getSecretAccessKey(d *schema.ResourceData) *string {
	s := writeonly.GetString(d, "secret_access_key", "secret_access_key_wo")
	if s == "" {
		return nil
	}
	return &s
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateBedrockAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		SecretStoreID: d.Get("secret_store_id").(string),
		Region:        d.Get("region").(string),
	}

	if v, ok := d.GetOk("access_key_id"); ok {
		s := v.(string)
		input.AccessKeyID = &s
	}

	input.SecretAccessKey = getSecretAccessKey(d)

	// Validate that access_key_id and secret_access_key are set/unset together
	hasKeyID := input.AccessKeyID != nil
	hasSecret := input.SecretAccessKey != nil
	if hasKeyID != hasSecret {
		return diag.Errorf("access_key_id and secret_access_key (or secret_access_key_wo) must both be set or both be omitted")
	}

	credential, err := client.CreateBedrockAccessCredential(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	return resourceRead(ctx, d, meta)
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	id := d.Id()
	if id == "" {
		if v, ok := d.GetOk("id"); ok {
			id = v.(string)
		}
	}

	if id == "" {
		return diag.Errorf("id is required")
	}

	credential, err := client.GetBedrockAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":                     credential.Name,
		"description":              credential.Description,
		"deployment_ids":           credential.DeploymentIDs,
		"region":                   credential.Region,
		"has_provider_credentials": credential.HasProviderCredentials,
		"type":                     string(credential.Type),
		"kind":                     credential.Kind,
		"secret_store_id":          credential.SecretStoreID,
	}

	if credential.AccessKeyID != nil {
		fields["access_key_id"] = *credential.AccessKeyID
	} else {
		fields["access_key_id"] = ""
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	input := &client.UpdateBedrockAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("secret_store_id") {
		v := d.Get("secret_store_id").(string)
		input.SecretStoreID = client.NewSecretStoreIDUpdate(v)
	}
	if d.HasChange("region") {
		v := d.Get("region").(string)
		input.Region = &v
	}

	awsKeysChanged := d.HasChange("access_key_id") ||
		d.HasChange("secret_access_key") ||
		d.HasChange("secret_access_key_wo") ||
		d.HasChange("secret_access_key_wo_version")

	if awsKeysChanged {
		if v, ok := d.GetOk("access_key_id"); ok {
			s := v.(string)
			input.AccessKeyID = &s
		}
		input.SecretAccessKey = getSecretAccessKey(d)

		// Validate that access_key_id and secret_access_key are set/unset together
		hasKeyID := input.AccessKeyID != nil
		hasSecret := input.SecretAccessKey != nil
		if hasKeyID != hasSecret {
			return diag.Errorf("access_key_id and secret_access_key (or secret_access_key_wo) must both be set or both be omitted")
		}
	}

	if _, err := client.UpdateBedrockAccessCredential(ctx, c, id, input); err != nil {
		return diag.FromErr(err)
	}

	// If AWS keys were removed, clear access_key_id to avoid drift detection from empty string
	if awsKeysChanged && input.AccessKeyID == nil {
		if err := d.Set("access_key_id", ""); err != nil {
			return diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("failed to set access_key_id: %s", err),
			}}
		}
	}

	return resourceRead(ctx, d, meta)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	err := client.DeleteAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
