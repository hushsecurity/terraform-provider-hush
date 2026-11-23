package cloudflare_r2_access_credential

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
		Description:   "Manage Cloudflare R2 access credentials in the Hush Security platform. With the root API token, the access manager issues an R2 token per policy version and hands the workload its S3 credentials: `access_key_id`, `secret_access_key` and `endpoint`.",
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

func getAPIToken(d *schema.ResourceData) string {
	return writeonly.GetString(d, "api_token", "api_token_wo")
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateCloudflareR2AccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		SecretStoreID: d.Get("secret_store_id").(string),
		AccountID:     d.Get("account_id").(string),
		APIToken:      getAPIToken(d),
	}

	credential, err := client.CreateCloudflareR2AccessCredential(ctx, c, input)
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

	credential, err := client.GetCloudflareR2AccessCredential(ctx, c, id)
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

	input := &client.UpdateCloudflareR2AccessCredentialInput{}

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
	// Only on a change: every token sent issues new credentials to every
	// policy on the credential.
	if d.HasChange("api_token") || d.HasChange("api_token_wo_version") {
		v := getAPIToken(d)
		input.APIToken = &v
	}

	_, err := client.UpdateCloudflareR2AccessCredential(ctx, c, id, input)
	if err != nil {
		// A planned token in state would hide an unsent change and block the retry.
		if keepPriorState(err) {
			d.Partial(true)
		}
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, meta)
}

// keepPriorState reports whether a failed update may not have reached Hush.
// Only a failed status wait follows an accepted PATCH, so Hush has the token.
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
