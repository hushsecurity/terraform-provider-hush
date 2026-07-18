package kafka_access_credential

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/credutil"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Kafka dynamic access credentials in the Hush Security platform.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		CustomizeDiff: customizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ResourceSchema(),
	}
}

// customizeDiff rejects deployment_ids changes after creation and enforces the
// per-engine field rules (mirroring midgard's _validate_engine_fields): every
// engine requires all its fields except the optional tls/tls_ca, and forbids the
// other engine's fields.
func customizeDiff(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if err := credutil.ForbidDeploymentIDsChange(ctx, d, meta); err != nil {
		return err
	}
	return validateEngineFields(d)
}

func validateEngineFields(d *schema.ResourceDiff) error {
	engine := d.Get("engine").(string)

	var required, forbidden []string
	switch engine {
	case engineNative:
		required = []string{"bootstrap_servers", "username", "sasl_mechanism", "password"}
		forbidden = []string{"project", "service_name", "token"}
	case engineAiven:
		required = []string{"project", "service_name", "token"}
		forbidden = []string{"bootstrap_servers", "username", "sasl_mechanism", "tls", "tls_ca", "password"}
	default:
		return nil
	}

	var missing []string
	for _, f := range required {
		if !attrSet(d, f) {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("engine %q requires: %s", engine, strings.Join(missing, ", "))
	}

	var present []string
	for _, f := range forbidden {
		if attrSet(d, f) {
			present = append(present, f)
		}
	}
	if len(present) > 0 {
		return fmt.Errorf("engine %q does not allow: %s", engine, strings.Join(present, ", "))
	}

	return nil
}

// attrSet reports whether attr is configured. Each secret may be supplied via
// its plain attribute or its write-only counterpart, so either counts as set.
func attrSet(d *schema.ResourceDiff, attr string) bool {
	switch attr {
	case "password":
		return rawSet(d, "password") || rawSet(d, "password_wo")
	case "token":
		return rawSet(d, "token") || rawSet(d, "token_wo")
	default:
		return rawSet(d, attr)
	}
}

// rawSet reports whether attr is configured in raw config. An unknown value (a
// reference resolved at apply, e.g. random_password.x.result) counts as set and
// is validated by the backend; null does not, so schema defaults stay unset.
func rawSet(d *schema.ResourceDiff, attr string) bool {
	rc := d.GetRawConfig()
	if rc.IsNull() {
		return false
	}
	v := rc.GetAttr(attr)
	if v.IsNull() {
		return false
	}
	if !v.IsKnown() {
		return true
	}
	if v.Type() == cty.String {
		return v.AsString() != ""
	}
	return true
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*client.Client)

	deploymentIDs := make([]string, 0)
	if v, ok := d.GetOk("deployment_ids"); ok {
		for _, item := range v.([]any) {
			deploymentIDs = append(deploymentIDs, item.(string))
		}
	}

	engine := d.Get("engine").(string)
	input := &client.CreateKafkaAccessCredentialInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		DeploymentIDs: deploymentIDs,
		SecretStoreID: d.Get("secret_store_id").(string),
		Engine:        engine,
	}

	switch engine {
	case engineNative:
		input.BootstrapServers = d.Get("bootstrap_servers").(string)
		input.Username = d.Get("username").(string)
		input.SaslMechanism = d.Get("sasl_mechanism").(string)
		input.TLS = d.Get("tls").(bool)
		if v, ok := d.GetOk("tls_ca"); ok {
			input.TLSCA = v.(string)
		}
		input.Password = writeonly.GetString(d, "password", "password_wo")
	case engineAiven:
		input.Project = d.Get("project").(string)
		input.ServiceName = d.Get("service_name").(string)
		input.Token = writeonly.GetString(d, "token", "token_wo")
	}

	credential, err := client.CreateKafkaAccessCredential(ctx, c, input)
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

	credential, err := client.GetKafkaAccessCredential(ctx, c, id)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(credential.ID)

	fields := map[string]any{
		"name":              credential.Name,
		"description":       credential.Description,
		"deployment_ids":    credential.DeploymentIDs,
		"engine":            credential.Engine,
		"bootstrap_servers": credential.BootstrapServers,
		"username":          credential.Username,
		"sasl_mechanism":    credential.SaslMechanism,
		"tls":               credential.TLS,
		"tls_ca":            credential.TLSCA,
		"project":           credential.Project,
		"service_name":      credential.ServiceName,
		"type":              string(credential.Type),
		"kind":              credential.Kind,
		"secret_store_id":   credential.SecretStoreID,
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

	input := &client.UpdateKafkaAccessCredentialInput{}

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
		input.SecretStoreID = &v
	}
	if d.HasChange("bootstrap_servers") {
		v := d.Get("bootstrap_servers").(string)
		input.BootstrapServers = &v
	}
	if d.HasChange("username") {
		v := d.Get("username").(string)
		input.Username = &v
	}
	if d.HasChange("sasl_mechanism") {
		v := d.Get("sasl_mechanism").(string)
		input.SaslMechanism = &v
	}
	if d.HasChange("tls") {
		v := d.Get("tls").(bool)
		input.TLS = &v
	}
	if d.HasChange("tls_ca") {
		v := d.Get("tls_ca").(string)
		input.TLSCA = &v
	}
	if d.HasChange("project") {
		v := d.Get("project").(string)
		input.Project = &v
	}
	if d.HasChange("service_name") {
		v := d.Get("service_name").(string)
		input.ServiceName = &v
	}
	if d.HasChange("password") || d.HasChange("password_wo") || d.HasChange("password_wo_version") {
		password := writeonly.GetString(d, "password", "password_wo")
		input.Password = &password
	}
	if d.HasChange("token") || d.HasChange("token_wo") || d.HasChange("token_wo_version") {
		token := writeonly.GetString(d, "token", "token_wo")
		input.Token = &token
	}

	_, err := client.UpdateKafkaAccessCredential(ctx, c, id, input)
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
