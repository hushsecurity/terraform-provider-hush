package redis_access_credential

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Redis dynamic access credentials in the Hush Security platform.",
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

	db := d.Get("database").(int)
	input := &client.CreateRedisAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		Host:          d.Get("host").(string),
		Port:          d.Get("port").(int),
		Password:      password,
		Database:      &db,
		TLS:           d.Get("tls").(bool),
		Engine:        d.Get("engine").(string),
	}

	if v, ok := d.GetOk("username"); ok {
		input.Username = v.(string)
	}

	if v, ok := d.GetOk("tls_ca"); ok {
		input.TLSCA = v.(string)
	}

	if v, ok := d.GetOk("cache_engine"); ok {
		input.CacheEngine = v.(string)
	}
	if v, ok := d.GetOk("region"); ok {
		input.Region = v.(string)
	}
	if v, ok := d.GetOk("user_group_id"); ok {
		input.UserGroupID = v.(string)
	}
	if v, ok := d.GetOk("access_key_id"); ok {
		input.AccessKeyID = v.(string)
	}
	if v, ok := d.GetOk("secret_access_key"); ok {
		input.SecretAccessKey = v.(string)
	} else if v, ok := d.GetOk("secret_access_key_wo"); ok {
		input.SecretAccessKey = v.(string)
	}

	credential, err := client.CreateRedisAccessCredential(ctx, c, input)
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

	credential, err := client.GetRedisAccessCredential(ctx, c, id)
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
		"host":           credential.Host,
		"port":           credential.Port,
		"username":       credential.Username,
		"database":       credential.Database,
		"tls":            credential.TLS,
		"tls_ca":         credential.TLSCA,
		"engine":         credential.Engine,
		"cache_engine":   credential.CacheEngine,
		"region":         credential.Region,
		"user_group_id":  credential.UserGroupID,
		"access_key_id":  credential.AccessKeyID,
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

	input := &client.UpdateRedisAccessCredentialInput{}

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
	if d.HasChange("database") {
		v := d.Get("database").(int)
		input.Database = &v
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
	if d.HasChange("engine") {
		v := d.Get("engine").(string)
		input.Engine = &v
	}
	if d.HasChange("cache_engine") {
		v := d.Get("cache_engine").(string)
		input.CacheEngine = &v
	}
	if d.HasChange("region") {
		v := d.Get("region").(string)
		input.Region = &v
	}
	if d.HasChange("user_group_id") {
		v := d.Get("user_group_id").(string)
		input.UserGroupID = &v
	}
	if d.HasChange("access_key_id") {
		v := d.Get("access_key_id").(string)
		input.AccessKeyID = &v
	}
	if d.HasChange("secret_access_key") || d.HasChange("secret_access_key_wo") || d.HasChange("secret_access_key_wo_version") {
		var secret string
		if v, ok := d.GetOk("secret_access_key"); ok {
			secret = v.(string)
		} else if v, ok := d.GetOk("secret_access_key_wo"); ok {
			secret = v.(string)
		}
		input.SecretAccessKey = &secret
	}

	_, err := client.UpdateRedisAccessCredential(ctx, c, id, input)
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
