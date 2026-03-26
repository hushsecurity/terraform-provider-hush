package snowflake_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Snowflake dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	input := &client.CreateSnowflakeAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		Account:       d.Get("account").(string),
		Warehouse:     d.Get("warehouse").(string),
		Database:      d.Get("database").(string),
		Schema:        d.Get("schema").(string),
		Username:      d.Get("username").(string),
		AuthMethod:    d.Get("auth_method").(string),
	}

	if v, ok := d.GetOk("role"); ok {
		input.Role = v.(string)
	}

	if v, ok := d.GetOk("password"); ok {
		input.Password = v.(string)
	} else if v, ok := d.GetOk("password_wo"); ok {
		input.Password = v.(string)
	}

	if v, ok := d.GetOk("private_key"); ok {
		input.PrivateKey = v.(string)
	} else if v, ok := d.GetOk("private_key_wo"); ok {
		input.PrivateKey = v.(string)
	}

	credential, err := client.CreateSnowflakeAccessCredential(ctx, c, input)
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

	credential, err := client.GetSnowflakeAccessCredential(ctx, c, id)
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
		"account":        credential.Account,
		"warehouse":      credential.Warehouse,
		"database":       credential.Database,
		"schema":         credential.Schema,
		"role":           credential.Role,
		"username":       credential.Username,
		"auth_method":    credential.AuthMethod,
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

	input := &client.UpdateSnowflakeAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("account") {
		v := d.Get("account").(string)
		input.Account = &v
	}
	if d.HasChange("warehouse") {
		v := d.Get("warehouse").(string)
		input.Warehouse = &v
	}
	if d.HasChange("database") {
		v := d.Get("database").(string)
		input.Database = &v
	}
	if d.HasChange("schema") {
		v := d.Get("schema").(string)
		input.Schema = &v
	}
	if d.HasChange("role") {
		v := d.Get("role").(string)
		input.Role = &v
	}
	if d.HasChange("username") {
		v := d.Get("username").(string)
		input.Username = &v
	}
	if d.HasChange("auth_method") {
		v := d.Get("auth_method").(string)
		input.AuthMethod = &v
	}
	if d.HasChange("password") || d.HasChange("password_wo") || d.HasChange("password_wo_version") {
		var password string
		if v, ok := d.GetOk("password"); ok {
			password = v.(string)
		} else if v, ok := d.GetOk("password_wo"); ok {
			password = v.(string)
		}
		input.Password = &password
	}
	if d.HasChange("private_key") || d.HasChange("private_key_wo") || d.HasChange("private_key_wo_version") {
		var privateKey string
		if v, ok := d.GetOk("private_key"); ok {
			privateKey = v.(string)
		} else if v, ok := d.GetOk("private_key_wo"); ok {
			privateKey = v.(string)
		}
		input.PrivateKey = &privateKey
	}

	_, err := client.UpdateSnowflakeAccessCredential(ctx, c, id, input)
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
