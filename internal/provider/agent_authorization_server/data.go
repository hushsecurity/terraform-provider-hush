package agent_authorization_server

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	dataSourceDescription = "The authorization server a deployment's service agents sign in at, " +
		"as its OAuth metadata reports it. Not the gateway's own one, which MCP clients use for users."
	deploymentIDDesc = "The deployment whose agent authorization server to read."
	issuerDesc       = "The issuer, which is also the audience ('aud') a service agent's workload token must carry, " +
		"e.g. the 'Audience' of sts:GetWebIdentityToken or of a projected Kubernetes service account token. " +
		"Not the deployment's 'oidc_provider' audience."
	tokenEndpointDesc = "Where a service agent exchanges its workload token."
)

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: dataRead,
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Description:  deploymentIDDesc,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
			},
			"issuer":         {Description: issuerDesc, Type: schema.TypeString, Computed: true},
			"token_endpoint": {Description: tokenEndpointDesc, Type: schema.TypeString, Computed: true},
		},
	}
}

func dataRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	id := d.Get("deployment_id").(string)
	as, err := client.GetAgentAuthorizationServer(ctx, c, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	if err := d.Set("issuer", as.Issuer); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("token_endpoint", as.TokenEndpoint); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
