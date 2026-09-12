package deployment

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	idDesc              = "The unique identifier of the deployment"
	nameDesc            = "The name of the deployment"
	descriptionDesc     = "The description of the deployment"
	envTypeDesc         = "The environment type for the deployment (dev, prod)"
	kindDesc            = "The deployment kind (k8s, hosted, ecs, serverless). Only 'hosted' and 'k8s' can carry an agent gateway."
	statusDesc          = "The current status of the deployment"
	tokenDesc           = "The deployment token for authentication"
	passwordDesc        = "The deployment password for authentication"
	imagePullSecretDesc = "The image pull secret for accessing private container images"

	agwDesc           = "Agent gateway configuration. What it describes follows 'kind': on a 'hosted' deployment the block is required and places a gateway Hush runs, so it takes 'region' and no 'hostname'; on a 'k8s' deployment it is optional and records a gateway you run yourself, so it takes 'hostname' and no 'region'. No other kind accepts the block."
	agwDataDesc       = "The deployment's agent gateway, or nothing when it has none. A gateway Hush runs reports the 'region' it is placed in, one you run yourself reports the 'hostname' it answers on, and both report the derived 'gateway_url'."
	agwHostnameDesc   = "The external hostname (FQDN) clients and the OAuth browser flow reach the gateway on, as a bare lowercase domain name without a scheme, port or path. Required on a 'k8s' deployment and not valid on a 'hosted' one, which derives its own address. It must be exactly the 'agw.hostname' the hush-agw chart was installed with: nothing reconciles the two, and connecting an application fails while they disagree. Hush's own domains are reserved."
	agwRegionDesc     = "The region a Hush-hosted gateway is placed in. Required on a 'hosted' deployment and not valid on a 'k8s' one. The accepted values are 'iad' (US East, N. Virginia) and 'fra' (Europe, Frankfurt). Placement is fixed when the gateway is built and the API offers no way to change it, so Terraform refuses a change rather than acting on one. Moving a gateway means deleting the deployment and creating another, which detaches every application bound to the gateway, since applications follow the deployment id."
	agwGatewayURLDesc = "The URL agents connect to the gateway on, derived by Hush. For a hosted gateway this is the only way to learn the address, since the name is generated from the deployment id and the region."

	oidcProviderDesc        = "Optional OIDC provider configuration enabling passwordless deployment token exchange. When set, the deployment can exchange a signed OIDC token (for example a Kubernetes service account token) for a deployment token instead of using the password. Repeat the block to trust more than one issuer. Every block is stored in the API's 'oidc_providers' field, and each issuer may appear once."
	oidcIssuerDesc          = "The OIDC issuer URL (must be HTTPS). Its OpenID configuration and JWKS are used to verify presented assertions."
	oidcAudienceDesc        = "The audience claim expected in presented OIDC assertions."
	oidcAllowedSubjectsDesc = "Optional list of allowed subject claims. A trailing '*' acts as a prefix wildcard (for example 'system:serviceaccount:hush-security:*'). When omitted, any subject is accepted."
)

// The two kinds that carry a gateway. A hosted one is run by Hush and placed
// by region; a k8s one may run a gateway the customer installed from the Helm
// chart, which is why no other self-hosted kind accepts the block.
const (
	deploymentKindK8s    = "k8s"
	deploymentKindHosted = "hosted"
)

// gatewayHostnamePattern matches a bare domain name: lowercase labels joined by
// at least one dot. It is the rule the hush-agw chart applies to the same
// value, so a hostname refused here is refused there too.
var gatewayHostnamePattern = regexp.MustCompile(
	`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)+$`)

// validateGatewayHostname refuses a malformed hostname at plan time and names
// the rule that was broken, the lowercase one separately because the API
// rejects mixed case rather than normalizing it -- so a caller reads back
// exactly what it wrote and nothing drifts.
//
// Checking the TLD against a list and reserving Hush's own domains stay API
// checks: both move on their own schedule, and a copy pinned to a provider
// release would refuse a hostname the API accepts.
func validateGatewayHostname(v any, path string) ([]string, []error) {
	host := v.(string)
	if host != strings.ToLower(host) {
		return nil, []error{fmt.Errorf("%s: must be lowercase", path)}
	}
	if !gatewayHostnamePattern.MatchString(host) {
		return nil, []error{fmt.Errorf(
			"%s: must be a bare domain name, without a scheme, port or path",
			path)}
	}
	return nil, nil
}

// maxOidcProviders mirrors the API cap on the field these blocks are stored
// in. Every entry is another key set able to mint tokens for the deployment,
// so the ceiling is worth stating here rather than discovering on a round trip.
const maxOidcProviders = 8

func DeploymentResourceSchema() map[string]*schema.Schema {
	s := DeploymentDataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description: nameDesc,
		Type:        schema.TypeString,
		Required:    true,
	}
	s["description"] = &schema.Schema{
		Description: descriptionDesc,
		Type:        schema.TypeString,
		Optional:    true,
	}
	s["env_type"] = &schema.Schema{
		Description: envTypeDesc,
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "dev",
		ValidateFunc: validation.StringInSlice([]string{
			"dev",
			"prod",
		}, false),
	}
	s["kind"] = &schema.Schema{
		Description: kindDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.StringInSlice([]string{
			deploymentKindK8s,
			deploymentKindHosted,
			"ecs",
			"serverless",
		}, false),
	}

	s["token"] = &schema.Schema{
		Description: tokenDesc,
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}
	s["password"] = &schema.Schema{
		Description: passwordDesc,
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}
	s["image_pull_secret"] = &schema.Schema{
		Description: imagePullSecretDesc,
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
	}

	// Neither member can be Required, because which one applies is decided by
	// the deployment kind and the schema cannot see it. deploymentCustomizeDiff
	// demands the right one and refuses the other, so the pair is still settled
	// before any request.
	s["agw"] = &schema.Schema{
		Description: agwDesc,
		Type:        schema.TypeList,
		Optional:    true,
		// One gateway per deployment: the API holds a single agw object.
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"hostname": {
					Description:  agwHostnameDesc,
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateGatewayHostname,
				},
				"region": {
					Description: agwRegionDesc,
					Type:        schema.TypeString,
					Optional:    true,
					// Deliberately not ForceNew, though the value is set once:
					// see validateRegionImmutable, which refuses the change
					// instead of turning it into a destroy.
				},
				"gateway_url": {
					Description: agwGatewayURLDesc,
					Type:        schema.TypeString,
					Computed:    true,
				},
			},
		},
	}

	s["oidc_provider"] = &schema.Schema{
		Description: oidcProviderDesc,
		Type:        schema.TypeList,
		Optional:    true,
		// The API caps the list it is stored in. Rejecting here says so while
		// the caller can still act on it, rather than after a round trip.
		MaxItems: maxOidcProviders,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"issuer": {
					Description:  oidcIssuerDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.IsURLWithHTTPS,
				},
				"audience": {
					Description: oidcAudienceDesc,
					Type:        schema.TypeString,
					Required:    true,
				},
				"allowed_subjects": {
					Description: oidcAllowedSubjectsDesc,
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}

	return s
}

func DeploymentDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description:   idDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   nameDesc,
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"id"},
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"env_type": {
			Description: envTypeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: statusDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"agw": {
			Description: agwDataDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hostname": {
						Description: agwHostnameDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"region": {
						Description: agwRegionDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"gateway_url": {
						Description: agwGatewayURLDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"oidc_provider": {
			Description: oidcProviderDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"issuer": {
						Description: oidcIssuerDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"audience": {
						Description: oidcAudienceDesc,
						Type:        schema.TypeString,
						Computed:    true,
					},
					"allowed_subjects": {
						Description: oidcAllowedSubjectsDesc,
						Type:        schema.TypeList,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
	}
}

// Helper Functions

// deploymentRead serves the resource, which owns the deployment it reads.
func deploymentRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return readDeployment(ctx, d, m, true)
}

// deploymentDataSourceRead serves the data source, which only reports. It
// reads deployments this configuration does not manage and cannot plan a
// change to any of them, so it surfaces every field as the API answers it.
func deploymentDataSourceRead(
	ctx context.Context, d *schema.ResourceData, m any,
) diag.Diagnostics {
	return readDeployment(ctx, d, m, false)
}

func readDeployment(
	ctx context.Context, d *schema.ResourceData, m any, managed bool,
) diag.Diagnostics {
	c := m.(*client.Client)

	var deployment *client.Deployment
	var err error

	if id := d.Id(); id != "" {
		deployment, err = client.GetDeployment(ctx, c, id)
		if err != nil {
			// Handle 404 errors gracefully by removing from state
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			} else {
				return diag.FromErr(err)
			}
		}
	} else if id, exists := d.GetOk("id"); exists {
		// Lookup by ID provided in configuration
		deploymentID := id.(string)
		deployment, err = client.GetDeployment(ctx, c, deploymentID)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				return diag.Errorf("no deployment found with ID: %s", deploymentID)
			} else {
				return diag.FromErr(err)
			}
		}
	} else if name, exists := d.GetOk("name"); exists {
		// Lookup by name
		deploymentName := name.(string)
		deployments, err := client.GetDeploymentsByName(ctx, c, deploymentName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to lookup deployment by name '%s': %w", deploymentName, err))
		}

		switch len(deployments) {
		case 0:
			return diag.Errorf("no deployment found with name: %s", deploymentName)
		case 1:
			deployment = &deployments[0]
		default:
			return diag.Errorf("multiple deployments found with name '%s'. Deployment names must be unique. Consider using the deployment ID instead for exact matching", deploymentName)
		}
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if d.Id() == "" {
		d.SetId(deployment.ID)
	}

	if diags := setDeploymentFields(d, deployment, managed); diags.HasError() {
		return diags
	}

	return nil
}

func setDeploymentFields(
	d *schema.ResourceData, deployment *client.Deployment, managed bool,
) diag.Diagnostics {
	fields := map[string]any{
		"name":        deployment.Name,
		"description": deployment.Description,
		"env_type":    deployment.EnvType,
		"kind":        deployment.Kind,
		"status":      deployment.Status,
	}

	for field, value := range fields {
		if err := d.Set(field, value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field, err))
		}
	}

	if err := d.Set(
		"oidc_provider", flattenOidcProviders(deployment, managed),
	); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set oidc_provider: %w", err))
	}

	if err := d.Set("agw", flattenAgw(deployment)); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set agw: %w", err))
	}

	return nil
}

// flattenAgw converts the deployment's agw object into the block list. Both
// kinds of gateway read back through the same block, each leaving the member
// that does not apply to it empty; an empty list rather than null when there is
// no gateway at all, or a deployment without one would show a permanent diff.
func flattenAgw(deployment *client.Deployment) []map[string]any {
	if deployment.Agw == nil {
		return []map[string]any{}
	}
	return []map[string]any{{
		"hostname":    deployment.Agw.Hostname,
		"region":      deployment.Agw.Region,
		"gateway_url": deployment.Agw.GatewayURL,
	}}
}

// flattenOidcProviders converts whichever OIDC field the deployment holds into
// the block list, or an empty list when it holds neither.
//
// The list is preferred and the singular field is the fallback, so a deployment
// this provider has not written since the list was introduced reads back as one
// block and produces no diff against a configuration that declares one.
//
// managed marks the resource, which owns the field. A hosted deployment's
// issuer is set by the API and cannot be sent back to it, so the resource has
// to hide it: a block in state that no configuration wrote plans a removal the
// API refuses. The data source is told nothing of the sort, because it has no
// configuration to diff against and hiding a value the deployment really holds
// would only make it report less than the API does.
func flattenOidcProviders(
	deployment *client.Deployment, managed bool,
) []map[string]any {
	if managed && deployment.Kind == deploymentKindHosted {
		return []map[string]any{}
	}

	configs := deployment.OidcProviders
	if len(configs) == 0 && deployment.OidcProvider != nil {
		configs = []client.OidcConfig{*deployment.OidcProvider}
	}

	out := make([]map[string]any, 0, len(configs))
	for _, oidc := range configs {
		out = append(out, map[string]any{
			"issuer":           oidc.Issuer,
			"audience":         oidc.Audience,
			"allowed_subjects": oidc.AllowedSubjects,
		})
	}
	return out
}
