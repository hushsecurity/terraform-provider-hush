package rabbitmq_access_credential

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
		Description:   "Manage RabbitMQ dynamic access credentials in the Hush Security platform.",
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

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	password := writeonly.GetString(d, "password", "password_wo")

	input := &client.CreateRabbitmqAccessCredentialInput{
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		DeploymentIDs:  deploymentIDs,
		SecretStoreID:  d.Get("secret_store_id").(string),
		Host:           d.Get("host").(string),
		Port:           d.Get("port").(int),
		ManagementPort: d.Get("management_port").(int),
		Password:       password,
		Vhost:          d.Get("vhost").(string),
		TLS:            d.Get("tls").(bool),
		AutoRotateRoot: d.Get("auto_rotate_root").(bool),
	}

	if v, ok := d.GetOk("username"); ok {
		input.Username = v.(string)
	}
	if v, ok := d.GetOk("tls_ca"); ok {
		input.TLSCA = v.(string)
	}

	credential, err := client.CreateRabbitmqAccessCredential(ctx, c, input)
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

	credential, err := client.GetRabbitmqAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":             credential.Name,
		"description":      credential.Description,
		"deployment_ids":   credential.DeploymentIDs,
		"host":             credential.Host,
		"port":             credential.Port,
		"management_port":  credential.ManagementPort,
		"username":         credential.Username,
		"vhost":            credential.Vhost,
		"tls":              credential.TLS,
		"tls_ca":           credential.TLSCA,
		"auto_rotate_root": credential.AutoRotateRoot,
		"type":             string(credential.Type),
		"kind":             credential.Kind,
		"secret_store_id":  credential.SecretStoreID,
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

	input := &client.UpdateRabbitmqAccessCredentialInput{}

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
	if d.HasChange("host") {
		v := d.Get("host").(string)
		input.Host = &v
	}
	if d.HasChange("port") {
		v := d.Get("port").(int)
		input.Port = &v
	}
	if d.HasChange("management_port") {
		v := d.Get("management_port").(int)
		input.ManagementPort = &v
	}
	if d.HasChange("username") {
		v := d.Get("username").(string)
		input.Username = &v
	}
	if d.HasChange("vhost") {
		v := d.Get("vhost").(string)
		input.Vhost = &v
	}
	if d.HasChange("tls") {
		v := d.Get("tls").(bool)
		input.TLS = &v
	}
	if d.HasChange("auto_rotate_root") {
		v := d.Get("auto_rotate_root").(bool)
		input.AutoRotateRoot = &v
	}
	if d.HasChange("tls_ca") {
		v := d.Get("tls_ca").(string)
		input.TLSCA = &v
	}
	if d.HasChange("password") || d.HasChange("password_wo") || d.HasChange("password_wo_version") {
		password := writeonly.GetString(d, "password", "password_wo")
		input.Password = &password
	}

	_, err := client.UpdateRabbitmqAccessCredential(ctx, c, id, input)
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
