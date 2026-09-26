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
	dataSourceDescription = "The consent methods a deployment's agent gateway may prompt a user through, " +
		"when a tool call needs the user's consent."
	dataDeploymentIDDesc = "The deployment whose gateway's methods to read."
	dataMethodsDesc      = "The methods allowed, such as 'oidc_callback' and 'push_slack'. Empty when none were ever set for the deployment."
	dataMethodDesc       = "Each method with how it prompts the user, as Hush describes it."
	methodNameDesc       = "The method."
	methodDescDesc       = "How the method prompts the user."
)

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: consentMethodsDataRead,
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Description:  dataDeploymentIDDesc,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
			},
			"methods": {
				Description: dataMethodsDesc,
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"method": {
				Description: dataMethodDesc,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name":        {Description: methodNameDesc, Type: schema.TypeString, Computed: true},
					"description": {Description: methodDescDesc, Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

// consentMethodsDataRead reports an empty set for a deployment nothing was
// ever stored for, and fails for one that does not exist. heimdall answers
// both with a 404, so the deployment itself is what tells them apart.
func consentMethodsDataRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	deploymentID := d.Get("deployment_id").(string)

	methods, err := client.GetAgwConsentMethods(ctx, c, deploymentID)
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.IsNotFound() {
		if _, depErr := client.GetDeployment(ctx, c, deploymentID); depErr != nil {
			var depAPIErr *client.APIError
			if errors.As(depErr, &depAPIErr) && depAPIErr.IsNotFound() {
				return diag.Errorf("no deployment found with ID: %s", deploymentID)
			}
			return diag.FromErr(depErr)
		}
		methods, err = nil, nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	names := make([]string, 0, len(methods))
	described := make([]any, 0, len(methods))
	for _, method := range methods {
		names = append(names, method.Name)
		described = append(described, map[string]any{"name": method.Name, "description": method.Description})
	}
	d.SetId(deploymentID)
	if err := d.Set("methods", names); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("method", described); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
