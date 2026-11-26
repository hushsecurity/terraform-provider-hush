package cloudinary_access_credential

import (
	"context"
	"errors"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/credutil"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Cloudinary access credentials in the Hush Security platform. With an Enterprise account management key, the access manager creates a product environment API key per policy version, gives it the privilege's role, and hands the workload `cloud_name`, `api_key`, `api_secret` and `cloudinary_url`.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: customdiff.All(validateAccountChangeDiff, credutil.ForbidDeploymentIDsChange),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func getAPIKey(d *schema.ResourceData) string {
	return writeonly.GetString(d, "api_key", "api_key_wo")
}

func getAPISecret(d *schema.ResourceData) string {
	return writeonly.GetString(d, "api_secret", "api_secret_wo")
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateCloudinaryAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		SecretStoreID: d.Get("secret_store_id").(string),
		AccountID:     d.Get("account_id").(string),
		APIKey:        getAPIKey(d),
		APISecret:     getAPISecret(d),
	}

	credential, err := client.CreateCloudinaryAccessCredential(ctx, c, input)
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

	credential, err := client.GetCloudinaryAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":            credential.Name,
		"description":     credential.Description,
		"deployment_ids":  credential.DeploymentIDs,
		"account_id":      credential.AccountID,
		"type":            string(credential.Type),
		"kind":            credential.Kind,
		"secret_store_id": credential.SecretStoreID,
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

	input := &client.UpdateCloudinaryAccessCredentialInput{}

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
	if d.HasChange("account_id") {
		v := d.Get("account_id").(string)
		input.AccountID = &v
	}
	// Only on a change: every key or secret sent issues new credentials to
	// every policy on the credential.
	if d.HasChange("api_key") || d.HasChange("api_key_wo_version") {
		v := getAPIKey(d)
		input.APIKey = &v
	}
	if d.HasChange("api_secret") || d.HasChange("api_secret_wo_version") {
		v := getAPISecret(d)
		input.APISecret = &v
	}

	_, err := client.UpdateCloudinaryAccessCredential(ctx, c, id, input)
	if err != nil {
		// A planned key in state would hide an unsent change and block the retry.
		if keepPriorState(err) {
			d.Partial(true)
		}
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, meta)
}

// keepPriorState reports whether a failed update may not have reached Hush.
// Only a failed status wait follows an accepted PATCH, so Hush has the key.
func keepPriorState(err error) bool {
	var waitErr *client.StatusWaitError
	return !errors.As(err, &waitErr)
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
