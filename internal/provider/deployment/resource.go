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

// deploymentCustomizeDiff refuses at plan time configurations the API also
// refuses, so the caller hears about them while it can still change them.
//
// A duplicate issuer, because selection matches on the issuer and takes the
// first entry: a second entry for one issuer is unreachable and the audience or
// subjects it carried would be dropped without a word.
//
// The rest is the agw block: against the deployment kind, in
// validateAgwForKind, and against what an existing gateway will accept, in
// validateRegionImmutable.
//
// A value taken from another resource is unknown at plan time and ResourceDiff
// yields the zero value for it, so nothing here may judge a value it cannot
// see: refusing a configuration that is very likely fine is worse than leaving
// it to the API. d.NewValueKnown is what tells the two apart -- an unknown kind
// is skipped entirely, and an unknown member of the agw block counts as
// written, which is all these rules need of it.
func deploymentCustomizeDiff(
	ctx context.Context, d *schema.ResourceDiff, m any,
) error {
	if err := validateAgwForKind(d); err != nil {
		return err
	}
	if err := validateRegionImmutable(d); err != nil {
		return err
	}

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

// validateAgwForKind holds the agw block to what the deployment kind allows.
// The API decides this per kind, and the schema cannot express it, so the whole
// rule set lives here:
//
// A hosted deployment is a gateway Hush runs, so the block is required and
// places it by region; the address is derived, hence no hostname, and the
// cluster the gateway runs in is trusted for token exchange, hence no
// oidc_provider -- the API fills that in itself and then refuses to be sent it.
//
// A k8s deployment may carry a gateway the customer installed, found by the
// hostname it was installed with; Hush chooses nothing, so there is no region.
//
// No other kind can run a gateway at all: it ships as a Helm chart.
//
// Skipped wherever kind is unknown at plan time, for the reason given on
// deploymentCustomizeDiff.
func validateAgwForKind(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("kind") {
		return nil
	}
	kind := d.Get("kind").(string)
	blocks := d.Get("agw").([]any)

	if kind == deploymentKindHosted {
		if len(blocks) == 0 {
			return fmt.Errorf(
				"agw: an agw block with a region is required on kind %q",
				deploymentKindHosted)
		}
		if len(d.Get("oidc_provider").([]any)) > 0 {
			return fmt.Errorf(
				"oidc_provider: not valid on kind %q, where Hush trusts the "+
					"cluster its gateway runs in and sets the issuer itself",
				deploymentKindHosted)
		}
	}
	if len(blocks) == 0 {
		return nil
	}
	if kind != deploymentKindHosted && kind != deploymentKindK8s {
		return fmt.Errorf("agw: a gateway is only valid on kind %q or %q, got %q",
			deploymentKindK8s, deploymentKindHosted, kind)
	}

	hasHostname := agwMemberSet(d, "hostname")
	hasRegion := agwMemberSet(d, "region")

	if kind == deploymentKindHosted {
		if !hasRegion {
			return fmt.Errorf(
				"agw.region: required on kind %q, which places the gateway Hush runs",
				deploymentKindHosted)
		}
		if hasHostname {
			return fmt.Errorf(
				"agw.hostname: not valid on kind %q, which derives its own address",
				deploymentKindHosted)
		}
		return nil
	}

	if !hasHostname {
		return fmt.Errorf(
			"agw.hostname: required on kind %q, which records a gateway you run",
			deploymentKindK8s)
	}
	if hasRegion {
		return fmt.Errorf(
			"agw.region: not valid on kind %q, where Hush places nothing",
			deploymentKindK8s)
	}
	return nil
}

// agwMemberSet reports whether the configuration writes a member of the agw
// block, which is not the same as whether a value can be read for it.
//
// A member filled from another resource -- a hostname from the DNS record that
// publishes it, say -- is unknown until the apply, and ResourceDiff yields the
// empty string for it, exactly as it does for a member nobody wrote. Reading
// the value alone would refuse that configuration for a missing hostname the
// caller did supply. NewValueKnown separates the two: unknown means written.
func agwMemberSet(d *schema.ResourceDiff, name string) bool {
	key := "agw.0." + name
	if !d.NewValueKnown(key) {
		return true
	}
	value, _ := d.Get(key).(string)
	return value != ""
}

// validateRegionImmutable refuses a region change on a deployment that already
// exists. The API has no field for it on an update, so there is nothing to
// send: placement is decided when the gateway is built.
//
// Marking the field ForceNew would be the usual answer to that and is the wrong
// one here. It would turn a one-word edit into a destroy, and the replacement
// carries a new deployment id -- which is what applications are bound to, so
// deleting the old one detaches every app from the gateway. Re-attaching them
// is manual work that a plan reading "must be replaced" gives no warning of.
//
// Moving a gateway is therefore left outside this resource's update path, as a
// deliberate delete and create.
func validateRegionImmutable(d *schema.ResourceDiff) error {
	if d.Id() == "" || !d.NewValueKnown("agw.0.region") {
		return nil
	}
	before, after := d.GetChange("agw.0.region")
	was, _ := before.(string)
	now, _ := after.(string)
	// An empty side is the block being added or dropped, which the kind rules
	// above judge; only a move between two regions belongs here.
	if was == "" || now == "" || was == now {
		return nil
	}
	return fmt.Errorf(
		"agw.region: cannot be changed from %q to %q on an existing deployment. "+
			"Placement is fixed when the gateway is built and the API has no "+
			"field to change it. Restore %q, or move the gateway deliberately by "+
			"deleting this deployment and creating another -- which detaches "+
			"every application bound to the gateway",
		was, now, was)
}

func deploymentCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateDeploymentInput{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		EnvType:       d.Get("env_type").(string),
		Kind:          d.Get("kind").(string),
		OidcProviders: expandOidcProviders(d),
		Agw:           expandAgw(d),
	}

	resp, err := client.CreateDeploymentWithCredentials(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	// Sync the returned deployment (including oidc_provider) into state.
	if diags := setDeploymentFields(d, &resp.Deployment, true); diags.HasError() {
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
	// Dropping the block sends an explicit null, which removes the whole agw
	// object -- how a decommissioned gateway leaves a deployment whose sensor
	// stays. A block that is still there sends its hostname, and the API applies
	// only the fields it was sent, so the idp_id it may also hold is kept.
	//
	// A hosted gateway is never patched. Nothing in its block is writable, and
	// the API refuses both the removal and a region, so the only change that can
	// reach here is a derived gateway_url moving under an unrelated edit.
	if d.HasChange("agw") && d.Get("kind").(string) != deploymentKindHosted {
		input.Agw = client.NewAgwUpdate(expandAgwUpdate(d))
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	updated, err := client.UpdateDeployment(ctx, c, d.Id(), input)
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Take the state from the response rather than from the plan. The API
	// derives gateway_url from the hostname, and the plan has no entry for a
	// computed member of a block whose sibling changed, so leaving it would
	// keep the old gateway's URL after a move and an empty one after the block
	// is first added. Anything reading gateway_url in the same apply would get
	// that stale value; there is no read after an update to correct it.
	return setDeploymentFields(d, updated, true)
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

// expandAgw reads the agw block for a create, returning nil when none is
// present. Whichever member the kind does not use is left empty and so is not
// sent, which is what the API demands: it forbids unknown keys and refuses the
// field belonging to the other kind. validateAgwForKind has already settled
// which one that is, so an object with neither cannot get here.
//
// gateway_url is never sent. It is computed, and the request model has no
// field for it.
func expandAgw(d *schema.ResourceData) *client.Agw {
	raw := d.Get("agw").([]any)
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}
	fields := raw[0].(map[string]any)
	hostname, _ := fields["hostname"].(string)
	region, _ := fields["region"].(string)
	return &client.Agw{Hostname: hostname, Region: region}
}

// expandAgwUpdate is the update counterpart. It carries the hostname alone,
// because that is the only member a patch may change: a region is fixed when
// the gateway is placed and the field does not exist on the update model, so a
// change to one is refused at plan time rather than written.
//
// nil means the block is gone, which removes the gateway.
func expandAgwUpdate(d *schema.ResourceData) *client.Agw {
	agw := expandAgw(d)
	if agw == nil {
		return nil
	}
	return &client.Agw{Hostname: agw.Hostname}
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
