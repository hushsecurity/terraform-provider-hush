package mongodb_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage MongoDB dynamic access credentials in the Hush Security platform.",
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

	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	} else if v, ok := d.GetOk("password_wo"); ok {
		password = v.(string)
	}

	input := &client.CreateMongoDBAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		DBName:        d.Get("db_name").(string),
		Host:          d.Get("host").(string),
		Port:          d.Get("port").(int),
		Username:      d.Get("username").(string),
		Password:      password,
		AuthSource:    d.Get("auth_source").(string),
		TLS:           d.Get("tls").(bool),
	}

	if v, ok := d.GetOk("tls_ca"); ok {
		input.TLSCA = v.(string)
	}

	credential, err := client.CreateMongoDBAccessCredential(ctx, c, input)
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

	credential, err := client.GetMongoDBAccessCredential(ctx, c, id)
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
		"db_name":        credential.DBName,
		"host":           credential.Host,
		"port":           credential.Port,
		"username":       credential.Username,
		"auth_source":    credential.AuthSource,
		"tls":            credential.TLS,
		"tls_ca":         credential.TLSCA,
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

	input := &client.UpdateMongoDBAccessCredentialInput{}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}
	if d.HasChange("description") {
		v := d.Get("description").(string)
		input.Description = &v
	}
	if d.HasChange("deployment_ids") {
		list := d.Get("deployment_ids").([]any)
		ids := make([]string, len(list))
		for i, item := range list {
			ids[i] = item.(string)
		}
		input.DeploymentIDs = &ids
	}
	if d.HasChange("db_name") {
		v := d.Get("db_name").(string)
		input.DBName = &v
	}
	if d.HasChange("host") {
		v := d.Get("host").(string)
		input.Host = &v
	}
	if d.HasChange("port") {
		v := d.Get("port").(int)
		input.Port = &v
	}
	if d.HasChange("username") {
		v := d.Get("username").(string)
		input.Username = &v
	}
	if d.HasChange("auth_source") {
		v := d.Get("auth_source").(string)
		input.AuthSource = &v
	}
	if d.HasChange("tls") {
		v := d.Get("tls").(bool)
		input.TLS = &v
	}
	if d.HasChange("tls_ca") {
		v := d.Get("tls_ca").(string)
		input.TLSCA = &v
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

	_, err := client.UpdateMongoDBAccessCredential(ctx, c, id, input)
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
