package agw_consent_methods

import (
	"context"
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	resourceDescription = "The consent methods a deployment's agent gateway may prompt a user through, " +
		"when a tool call needs the user's consent. The deployment keeps one set of methods, so " +
		"declare this resource once per deployment. The API has no way to remove the setting: " +
		"destroying the resource only stops Terraform managing it, and the gateway keeps the " +
		"methods last applied."
	deploymentIDDesc = "The deployment whose gateway the methods apply to. Changing it moves " +
		"management to the other deployment and leaves this one's methods as they are."
	methodsDesc = "The methods allowed: 'oidc_callback' (the user signs in with the org's " +
		"identity provider and answers on a consent page) and 'push_slack' (a Slack direct " +
		"message, which needs a Slack integration that can message the org's users). At least " +
		"one is required. When the org's Slack integration is removed, Hush drops 'push_slack' " +
		"on its own, which shows up as drift."
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: consentMethodsSet,
		ReadContext:   consentMethodsRead,
		UpdateContext: consentMethodsSet,
		DeleteContext: consentMethodsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importByDeploymentID,
		},
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Description:  deploymentIDDesc,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
			},
			"methods": {
				Description: methodsDesc,
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(client.ConsentMethods, false),
				},
			},
		},
	}
}

// consentMethodsSet serves both create and update: the API has one call that
// replaces the stored methods whether or not any were stored before.
func consentMethodsSet(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	deploymentID := d.Get("deployment_id").(string)

	methods := make([]string, 0)
	for _, v := range d.Get("methods").(*schema.Set).List() {
		methods = append(methods, v.(string))
	}
	if _, err := client.SetAgwConsentMethods(ctx, c, deploymentID, methods); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(deploymentID)
	return consentMethodsRead(ctx, d, m)
}

func consentMethodsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	methods, err := client.GetAgwConsentMethods(ctx, c, d.Id())
	if err != nil {
		// Both a deployment that is gone and one with nothing stored answer
		// 404; either way there is nothing left for this resource to track.
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	names := make([]string, 0, len(methods))
	for _, method := range methods {
		names = append(names, method.Name)
	}
	if err := d.Set("deployment_id", d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("methods", names); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// consentMethodsDelete only forgets the resource. The API cannot clear a
// deployment's methods, and writing some other set in their place would be a
// guess at what the caller wanted the gateway left with.
func consentMethodsDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")
	return nil
}

func importByDeploymentID(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	if err := d.Set("deployment_id", d.Id()); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
