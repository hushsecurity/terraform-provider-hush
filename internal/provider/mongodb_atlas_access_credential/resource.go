package mongodb_atlas_access_credential

import (
	"context"
	"fmt"
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
		Description:   "Manage MongoDB Atlas dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: customdiff.All(validateAtlasAuth, credutil.ForbidDeploymentIDsChange),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

// validateAtlasAuth enforces the backend rule that exactly one authentication
// method is configured: a service account (client_id + client_secret) or an
// API key (public_key + private_key). It only runs on create, because the
// secrets are write-only and not visible on update.
func validateAtlasAuth(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Id() != "" {
		return nil
	}

	_, hasClientID := d.GetOk("client_id")
	_, hasClientSecretPlain := d.GetOk("client_secret")
	hasClientSecret := hasClientSecretPlain || writeonly.IsSet(d, "client_secret_wo")
	_, hasPublicKey := d.GetOk("public_key")
	_, hasPrivateKeyPlain := d.GetOk("private_key")
	hasPrivateKey := hasPrivateKeyPlain || writeonly.IsSet(d, "private_key_wo")

	if hasClientID != hasClientSecret {
		return fmt.Errorf("client_id and client_secret must both be set")
	}
	if hasPublicKey != hasPrivateKey {
		return fmt.Errorf("public_key and private_key must both be set")
	}
	if hasClientID == hasPublicKey {
		return fmt.Errorf("use either client_id + client_secret or public_key + private_key")
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

	input := &client.CreateMongoDBAtlasAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		GroupID:       d.Get("group_id").(string),
		DBName:        d.Get("db_name").(string),
		Host:          d.Get("host").(string),
		ClientID:      d.Get("client_id").(string),
		ClientSecret:  writeonly.GetString(d, "client_secret", "client_secret_wo"),
		PublicKey:     d.Get("public_key").(string),
		PrivateKey:    writeonly.GetString(d, "private_key", "private_key_wo"),
	}

	credential, err := client.CreateMongoDBAtlasAccessCredential(ctx, c, input)
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

	credential, err := client.GetMongoDBAtlasAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":           credential.Name,
		"description":    credential.Description,
		"deployment_ids": credential.DeploymentIDs,
		"group_id":       credential.GroupID,
		"db_name":        credential.DBName,
		"host":           credential.Host,
		"client_id":      credential.ClientID,
		"public_key":     credential.PublicKey,
		"type":           string(credential.Type),
		"kind":           credential.Kind,
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

	input := &client.UpdateMongoDBAtlasAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("group_id") {
		v := d.Get("group_id").(string)
		input.GroupID = &v
	}
	if d.HasChange("db_name") {
		v := d.Get("db_name").(string)
		input.DBName = &v
	}
	if d.HasChange("host") {
		v := d.Get("host").(string)
		input.Host = &v
	}
	if d.HasChange("client_id") {
		v := d.Get("client_id").(string)
		input.ClientID = &v
	}
	if d.HasChange("client_secret") || d.HasChange("client_secret_wo") || d.HasChange("client_secret_wo_version") {
		v := writeonly.GetString(d, "client_secret", "client_secret_wo")
		input.ClientSecret = &v
	}
	if d.HasChange("public_key") {
		v := d.Get("public_key").(string)
		input.PublicKey = &v
	}
	if d.HasChange("private_key") || d.HasChange("private_key_wo") || d.HasChange("private_key_wo_version") {
		v := writeonly.GetString(d, "private_key", "private_key_wo")
		input.PrivateKey = &v
	}

	_, err := client.UpdateMongoDBAtlasAccessCredential(ctx, c, id, input)
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
