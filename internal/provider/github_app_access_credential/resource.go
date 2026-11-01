package github_app_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/credutil"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage GitHub App access credentials in the Hush Security platform. The access manager mints short-lived installation tokens from the App's private key.",
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

func getPrivateKey(d *schema.ResourceData) string {
	return writeonly.GetString(d, "private_key", "private_key_wo")
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateGithubAppAccessCredentialInput{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		DeploymentIDs:  deploymentIDs,
		SecretStoreID:  d.Get("secret_store_id").(string),
		ClientID:       d.Get("client_id").(string),
		InstallationID: d.Get("installation_id").(string),
		PrivateKey:     getPrivateKey(d),
	}

	credential, err := client.CreateGithubAppAccessCredential(ctx, c, input)
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

	credential, err := client.GetGithubAppAccessCredential(ctx, c, id)
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
		"client_id":       credential.ClientID,
		"installation_id": credential.InstallationID,
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

	input := &client.UpdateGithubAppAccessCredentialInput{}

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
	if d.HasChange("private_key") || d.HasChange("private_key_wo") || d.HasChange("private_key_wo_version") {
		v := getPrivateKey(d)
		input.PrivateKey = &v
	}
	if d.HasChange("client_id") {
		v := d.Get("client_id").(string)
		input.ClientID = &v
	}
	if d.HasChange("installation_id") {
		v := d.Get("installation_id").(string)
		input.InstallationID = &v
	}

	_, err := client.UpdateGithubAppAccessCredential(ctx, c, id, input)
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
