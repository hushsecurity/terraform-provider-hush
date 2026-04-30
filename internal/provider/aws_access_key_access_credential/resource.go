package aws_access_key_access_credential

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage AWS access key dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: validateKeyPairing,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func validateKeyPairing(_ context.Context, d *schema.ResourceDiff, _ any) error {
	_, hasKeyID := d.GetOk("access_key_id_value")
	_, hasSecret := d.GetOk("secret_access_key")
	_, hasSecretWO := d.GetOk("secret_access_key_wo")

	if hasKeyID && !hasSecret && !hasSecretWO && d.Id() == "" {
		return fmt.Errorf("access_key_id_value requires one of secret_access_key or secret_access_key_wo")
	}
	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	secretAccessKey := getSecretAccessKey(d)

	input := &client.CreateAWSAccessKeyAccessCredentialInput{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		DeploymentIDs:      deploymentIDs,
		AccessKeyID:        d.Get("access_key_id_value").(string),
		SecretAccessKey:    secretAccessKey,
		PermissionBoundary: d.Get("permission_boundary").(bool),
	}

	credential, err := client.CreateAWSAccessKeyAccessCredential(ctx, c, input)
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

	credential, err := client.GetAWSAccessKeyAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":                credential.Name,
		"description":         credential.Description,
		"deployment_ids":      credential.DeploymentIDs,
		"access_key_id_value": credential.AccessKeyID,
		"permission_boundary": credential.PermissionBoundary,
		"type":                string(credential.Type),
		"kind":                credential.Kind,
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

	input := &client.UpdateAWSAccessKeyAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("access_key_id_value") {
		v := d.Get("access_key_id_value").(string)
		input.AccessKeyID = &v
	}
	if d.HasChange("secret_access_key") || d.HasChange("secret_access_key_wo") || d.HasChange("secret_access_key_wo_version") {
		secretAccessKey := getSecretAccessKey(d)
		input.SecretAccessKey = &secretAccessKey
	}
	if d.HasChange("permission_boundary") {
		v := d.Get("permission_boundary").(bool)
		input.PermissionBoundary = &v
	}

	_, err := client.UpdateAWSAccessKeyAccessCredential(ctx, c, id, input)
	if err != nil {
		return diag.FromErr(err)
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

// getSecretAccessKey reads the secret access key from either the regular
// attribute or the write-only attribute via GetRawConfig.
func getSecretAccessKey(d *schema.ResourceData) string {
	if v, ok := d.GetOk("secret_access_key"); ok {
		return v.(string)
	}
	rawConfig := d.GetRawConfig()
	if rawConfig.IsNull() {
		return ""
	}
	v := rawConfig.GetAttr("secret_access_key_wo")
	if v.IsNull() || !v.IsKnown() {
		return ""
	}
	return v.AsString()
}
