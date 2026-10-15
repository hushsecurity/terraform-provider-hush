package secret_store

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Manages a Hush Security secret store, describing where the access-manager materializes secrets for a set of deployments."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: resourceCreate,
		ReadContext:   secretStoreRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:        SecretStoreResourceSchema(),
		CustomizeDiff: customizeDiff,
	}
}

// customizeDiff applies the rules the schema cannot: which fields an auth
// method requires. Doing it here rather than at apply is the point -- the API
// refuses these too, and a customer should hear it from the plan.
func customizeDiff(
	_ context.Context, d *schema.ResourceDiff, _ any,
) error {
	if hasAuthBlock(d, "hc_vault") {
		return validateVaultAuth(d)
	}
	if hasAuthBlock(d, "az_key_vault") {
		if err := validateAzureAuth(d); err != nil {
			return err
		}
		return validateAzureCloudHost(d)
	}
	return nil
}

// validateAzureCloudHost pairs cloud with the host suffix it implies, which a
// ValidateFunc cannot do: it sees one attribute. Judged only once both are
// known, for the reason validateAzureAuth gives.
func validateAzureCloudHost(d *schema.ResourceDiff) error {
	const base = "az_key_vault.0."
	if !d.NewValueKnown(base+"vault_url") || !d.NewValueKnown(base+"cloud") {
		return nil
	}
	rawURL := d.Get(base + "vault_url").(string)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return nil // vault_url's own validator reports it
	}
	host := parsed.Hostname()
	cloud := azureCloudOrPublic(d.Get(base + "cloud").(string))
	if suffix, ok := vaultHostSuffixByCloud[cloud]; ok &&
		strings.HasSuffix(host, suffix) {
		return nil
	}
	// the usual mistake is leaving cloud unset on a sovereign vault, so name
	// the cloud the url belongs to
	for other, suffix := range vaultHostSuffixByCloud {
		if strings.HasSuffix(host, suffix) {
			return fmt.Errorf(
				"az_key_vault.vault_url is a %s vault; set cloud to %q, got cloud %q and %q",
				other, other, cloud, rawURL)
		}
	}
	return fmt.Errorf(
		"az_key_vault.vault_url must end with %s for cloud %q, got %q",
		vaultHostSuffixByCloud[cloud], cloud, rawURL)
}

func hasAuthBlock(d *schema.ResourceDiff, kind string) bool {
	v, ok := d.GetOk(kind)
	if !ok {
		return false
	}
	block := v.([]any)[0].(map[string]any)
	auth, ok := block["auth"].([]any)
	return ok && len(auth) != 0 && auth[0] != nil
}

// validateAzureAuth refuses the fields a method does not use rather than
// ignoring them: the access manager refuses them too, and a store's config
// cannot be edited afterwards.
//
// Judged only once known, for the reason validateVaultAuth gives: a value from
// an azuread_application is unknown at plan time.
func validateAzureAuth(d *schema.ResourceDiff) error {
	const base = "az_key_vault.0.auth.0."

	if !d.NewValueKnown(base + "method") {
		return nil
	}
	method := d.Get(base + "method").(string)
	clientIDKnown := d.NewValueKnown(base + "client_id")
	clientID := d.Get(base + "client_id").(string)

	if method == client.SecretStoreAzureAuthClientSecret {
		// the default method reads AZURE_TENANT_ID instead; this one has no
		// such fallback
		if d.NewValueKnown(base+"tenant_id") &&
			d.Get(base+"tenant_id").(string) == "" {
			return fmt.Errorf("auth method %q needs tenant_id", method)
		}
		if clientIDKnown && clientID == "" {
			return fmt.Errorf(
				"auth method %q needs client_id, the service principal's application id",
				method)
		}
		return nil
	}
	if clientIDKnown && clientID != "" {
		return fmt.Errorf("auth method %q does not use client_id", method)
	}
	if d.NewValueKnown(base+"tenant_id") &&
		d.Get(base+"tenant_id").(string) != "" {
		return fmt.Errorf("auth method %q does not use tenant_id", method)
	}
	return nil
}

// validateVaultAuth requires the field the named method uses and refuses the
// fields belonging to the others. A field of another method is refused rather
// than ignored: a store's config is immutable, so one created with a role it
// never uses cannot be corrected, and the mistake it usually represents is
// naming the wrong method.
//
// A value another resource supplies is unknown at plan time and ResourceDiff
// yields "" for it, so each field is judged only once it is known -- the rule
// deployment/resource.go states. role taken from a vault_kubernetes_auth_
// backend_role is the ordinary way to write this, and it is set by apply.
func validateVaultAuth(d *schema.ResourceDiff) error {
	const base = "hc_vault.0.auth.0."

	if !d.NewValueKnown(base + "method") {
		return nil
	}
	method := d.Get(base + "method").(string)

	if method == client.SecretStoreVaultAuthToken {
		for _, field := range []string{"role", "mount"} {
			if !d.NewValueKnown(base + field) {
				continue
			}
			if d.Get(base+field).(string) != "" {
				return fmt.Errorf("auth method %q does not use %s", method, field)
			}
		}
		return nil
	}
	if d.NewValueKnown(base+"role") && d.Get(base+"role").(string) == "" {
		return fmt.Errorf("auth method %q needs a role", method)
	}
	return nil
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	config, err := expandConfig(d)
	if err != nil {
		return diag.FromErr(err)
	}

	input := &client.CreateSecretStoreInput{
		Name:          d.Get("name").(string),
		DeploymentIDs: expandStringList(d.Get("deployment_ids").([]any)),
		Config:        *config,
	}
	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	store, err := client.CreateSecretStore(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(store.ID)

	return secretStoreRead(ctx, d, m)
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateSecretStoreInput{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
	if d.HasChange("deployment_ids") {
		ids := expandStringList(d.Get("deployment_ids").([]any))
		input.DeploymentIDs = &ids
		hasChanges = true
	}

	if hasChanges {
		if _, err := client.UpdateSecretStore(ctx, c, d.Id(), input); err != nil {
			if errResponse, ok := err.(*client.APIError); ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return secretStoreRead(ctx, d, m)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if err := client.DeleteSecretStore(ctx, c, d.Id()); err != nil {
		if errResponse, ok := err.(*client.APIError); ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// expandConfig reads the single present config block and builds the client config
// with the matching kind. ExactlyOneOf guarantees exactly one block is set.
func expandConfig(d *schema.ResourceData) (*client.SecretStoreConfig, error) {
	if v, ok := d.GetOk("aws_sm"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:     client.SecretStoreKindAWSSM,
			Prefix:   block["prefix"].(string),
			Region:   block["region"].(string),
			KmsKeyID: block["kms_key_id"].(string),
		}, nil
	}
	if v, ok := d.GetOk("aws_ssm"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:     client.SecretStoreKindAWSSSM,
			Prefix:   block["prefix"].(string),
			Region:   block["region"].(string),
			KmsKeyID: block["kms_key_id"].(string),
		}, nil
	}
	if v, ok := d.GetOk("gcp_sm"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:      client.SecretStoreKindGCPSM,
			Prefix:    block["prefix"].(string),
			ProjectID: block["project_id"].(string),
		}, nil
	}
	if v, ok := d.GetOk("k8s_secrets"); ok {
		block := v.([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:      client.SecretStoreKindK8sSecrets,
			Prefix:    block["prefix"].(string),
			Namespace: block["namespace"].(string),
		}, nil
	}
	if v, ok := d.GetOk("hc_vault"); ok {
		block := v.([]any)[0].(map[string]any)
		// the auth block is Required with MaxItems 1, so exactly one is here
		auth := block["auth"].([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:           client.SecretStoreKindHCVault,
			Prefix:         block["prefix"].(string),
			Address:        block["address"].(string),
			Mount:          block["mount"].(string),
			VaultNamespace: block["vault_namespace"].(string),
			CaCert:         block["ca_cert"].(string),
			Auth: &client.SecretStoreAuth{
				Method: auth["method"].(string),
				Mount:  auth["mount"].(string),
				Role:   auth["role"].(string),
			},
		}, nil
	}
	if v, ok := d.GetOk("az_key_vault"); ok {
		block := v.([]any)[0].(map[string]any)
		// the auth block is Required with MaxItems 1, so exactly one is here
		auth := block["auth"].([]any)[0].(map[string]any)
		return &client.SecretStoreConfig{
			Kind:     client.SecretStoreKindAzKeyVault,
			Prefix:   block["prefix"].(string),
			VaultURL: block["vault_url"].(string),
			Cloud:    block["cloud"].(string),
			Auth: &client.SecretStoreAuth{
				Method:   auth["method"].(string),
				TenantID: auth["tenant_id"].(string),
				ClientID: auth["client_id"].(string),
			},
		}, nil
	}
	return nil, fmt.Errorf("one of the config blocks (%s) must be set",
		strings.Join(configBlockNames, ", "))
}

func expandStringList(items []any) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}
