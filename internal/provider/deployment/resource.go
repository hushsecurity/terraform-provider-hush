package deployment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Deployment resource for managing Hush Security deployments"

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: deploymentCreate,
		ReadContext:   deploymentRead,
		UpdateContext: deploymentUpdate,
		DeleteContext: deploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: deploymentCustomizeDiff,
		Schema:        DeploymentResourceSchema(),
	}
}

// deploymentCustomizeDiff refuses a duplicate issuer at plan time. The API
// refuses one too, but selection matches on the issuer and takes the first
// entry, so a second entry for one issuer is unreachable and the audience or
// subjects it carried would be dropped without a word -- worth saying while the
// caller can still change it.
//
// Literals only. An issuer taken from another resource is unknown at plan time
// and ResourceDiff yields the zero value for it -- d.NewValueKnown reports the
// same thing without relying on that -- so it is skipped rather than reported,
// because refusing a configuration that is very likely fine is worse than
// leaving it to the API. For those the API stays the only check.
func deploymentCustomizeDiff(
	ctx context.Context, d *schema.ResourceDiff, m any,
) error {
	seen := make(map[string]struct{})
	for _, entry := range d.Get("oidc_provider").([]any) {
		fields, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		issuer, _ := fields["issuer"].(string)
		if issuer == "" {
			continue
		}
		if _, dup := seen[issuer]; dup {
			return fmt.Errorf(
				"oidc_provider: issuer %q is configured more than once", issuer)
		}
		seen[issuer] = struct{}{}
	}
	return nil
}

func deploymentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateDeploymentInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		EnvType:       d.Get("env_type").(string),
		Kind:          d.Get("kind").(string),
		OidcProviders: expandOidcProviders(d),
	}

	resp, err := client.CreateDeploymentWithCredentials(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	// Sync the returned deployment (including oidc_provider) into state.
	if diags := setDeploymentFields(d, &resp.Deployment); diags.HasError() {
		return diags
	}

	// Set computed sensitive fields
	if err := d.Set("token", resp.Token); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set token: %w", err))
	}
	if err := d.Set("password", resp.Password); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set password: %w", err))
	}
	if err := d.Set("image_pull_secret", resp.ImagePullSecret); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set image_pull_secret: %w", err))
	}

	return nil
}

func deploymentUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateDeploymentInput{}
	hasChanges := false

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
	if d.HasChange("env_type") {
		envType := d.Get("env_type").(string)
		input.EnvType = &envType
		hasChanges = true
	}
	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("kind") {
		kind := d.Get("kind").(string)
		input.Kind = &kind
		hasChanges = true
	}
	// Write the list and clear the singular field in the same request. The API
	// resolves each field from the change or, when absent, from the stored
	// deployment, and refuses a change that would leave it holding both -- so
	// sending the list alone would be measured against a stored singular and
	// refused. Clearing it here is also what migrates a deployment that still
	// holds the singular, without ever passing through a state that trusts no
	// issuer.
	if d.HasChange("oidc_provider") {
		input.OidcProviders = client.NewOidcProvidersUpdate(expandOidcProviders(d))
		input.OidcProvider = client.NewOidcProviderUpdate(nil)
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	_, err := client.UpdateDeployment(ctx, c, d.Id(), input)
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

// expandOidcProviders reads the oidc_provider blocks from configuration,
// returning nil when none is present. One block per trusted issuer, all of
// them stored in the API's oidc_providers field.
func expandOidcProviders(d *schema.ResourceData) []client.OidcConfig {
	raw := d.Get("oidc_provider").([]any)
	if len(raw) == 0 {
		return nil
	}
	out := make([]client.OidcConfig, 0, len(raw))
	for _, entry := range raw {
		if entry == nil {
			continue
		}
		fields := entry.(map[string]any)
		cfg := client.OidcConfig{
			Issuer:   fields["issuer"].(string),
			Audience: fields["audience"].(string),
		}
		if subs, ok := fields["allowed_subjects"].([]any); ok && len(subs) > 0 {
			cfg.AllowedSubjects = make([]string, len(subs))
			for i, s := range subs {
				cfg.AllowedSubjects[i] = s.(string)
			}
		}
		out = append(out, cfg)
	}
	return out
}

func deploymentDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteDeployment(ctx, c, d.Id())
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
