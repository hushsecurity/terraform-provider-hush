package plaintext_access_credential

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
		Description:   "Manage plaintext access credentials in the Hush Security platform. Plaintext credentials store a single secret value that can be delivered as an environment variable to specified deployments.",
		CreateContext: plaintextAccessCredentialCreate,
		ReadContext:   plaintextAccessCredentialRead,
		UpdateContext: plaintextAccessCredentialUpdate,
		DeleteContext: plaintextAccessCredentialDelete,
		CustomizeDiff: credutil.ForbidDeploymentIDsChange,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: PlaintextAccessCredentialResourceSchema(),
	}
}

func plaintextAccessCredentialCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		list := v.([]any)
		for _, item := range list {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	secret := writeonly.GetString(d, "secret", "secret_wo")

	input := &client.CreatePlaintextAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		SecretStoreID: d.Get("secret_store_id").(string),
		Secret:        secret,
	}

	credential, err := client.CreatePlaintextAccessCredential(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	return plaintextAccessCredentialRead(ctx, d, meta)
}

func plaintextAccessCredentialRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	credential, err := client.GetPlaintextAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	if err := d.Set("name", credential.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", credential.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("deployment_ids", credential.DeploymentIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", string(credential.Type)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("secret_store_id", credential.SecretStoreID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func plaintextAccessCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)
	id := d.Id()

	input := &client.UpdatePlaintextAccessCredentialInput{}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		input.Description = &description
	}

	if d.HasChange("secret_store_id") {
		secretStoreID := d.Get("secret_store_id").(string)
		input.SecretStoreID = client.NewSecretStoreIDUpdate(secretStoreID)
	}

	if d.HasChange("secret") || d.HasChange("secret_wo") || d.HasChange("secret_wo_version") {
		secret := writeonly.GetString(d, "secret", "secret_wo")
		input.Secret = &secret
	}

	_, err := client.UpdatePlaintextAccessCredential(ctx, c, id, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return plaintextAccessCredentialRead(ctx, d, meta)
}

func plaintextAccessCredentialDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
