package auth0_access_credential

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc                = "The unique identifier of the Auth0 access credential"
	nameDesc              = "The name of the Auth0 access credential"
	descriptionDesc       = "The description of the Auth0 access credential"
	deploymentIDsDesc     = "List of deployment IDs that can access this credential. Currently limited to a single deployment"
	domainDesc            = "The tenant's canonical Auth0 domain, ending in `.auth0.com` (for example `acme.eu.auth0.com`). It is shown under Settings > General in the Auth0 dashboard. A custom domain is not accepted here: the Management API audience is fixed at tenant creation as `https://{canonical}/api/v2/` and does not move when a custom domain is added. Use `custom_domain` to route requests through one" //nolint: lll
	customDomainDesc      = "An Auth0 custom domain to route requests through, for example `auth.acme.com`. Optional. When set, the Management API and the token endpoint are called on this host and it is the domain delivered to the workload, while the audience is still derived from `domain`"                                                                                                                        //nolint: lll
	clientIDDesc          = "The client ID of the Auth0 machine-to-machine application authorized for the Management API"
	clientSecretDesc      = "The client secret of the Auth0 machine-to-machine application"
	clientSecretWODesc    = "The client secret (write-only). This is a write-only attribute that is more secure than `client_secret` because Terraform will not store this value in the state file. Either `client_secret` or `client_secret_wo` must be specified." //nolint: lll
	clientSecretWOVerDesc = "Used to trigger updates for `client_secret_wo`. This value should be changed when the secret content changes. Can be any value (e.g., a timestamp, version number, or hash)."                                                           //nolint: lll
	typeDesc              = "The type of access credential"
	kindDesc              = "The kind of access credential"
	secretStoreIDDesc     = "The ID of the secret store where this credential is saved (optional)"
)

// The audience derives from `domain` alone, so a custom domain 403s.
var canonicalDomainRe = regexp.MustCompile(`\.auth0\.com$`)

// Matches midgard's DomainStr, so a bad value fails at plan not apply.
var domainRe = regexp.MustCompile(
	`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`,
)

func ResourceSchema() map[string]*schema.Schema {
	s := DataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 1000),
	}
	s["deployment_ids"] = &schema.Schema{
		Description: deploymentIDsDesc + ". Changing this after creation is not supported; the credential must be deleted and recreated.",
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		MaxItems:    1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
		},
	}
	s["secret_store_id"] = &schema.Schema{
		Description:  secretStoreIDDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(`^sst-`), "secret_store_id must start with 'sst-'"),
	}
	s["domain"] = &schema.Schema{
		Description: domainDesc,
		Type:        schema.TypeString,
		Required:    true,
		ValidateFunc: validation.All(
			validation.StringMatch(domainRe, "domain must be a valid domain name"),
			validation.StringMatch(
				canonicalDomainRe,
				"domain must be the tenant's canonical Auth0 domain, ending in '.auth0.com'; set custom_domain to route requests through a custom domain",
			),
		),
	}
	s["custom_domain"] = &schema.Schema{
		Description: customDomainDesc,
		Type:        schema.TypeString,
		Optional:    true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 255),
			validation.StringMatch(domainRe, "custom_domain must be a valid domain name"),
		),
	}
	s["client_id"] = &schema.Schema{
		Description:  clientIDDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 256),
	}
	s["client_secret"] = &schema.Schema{
		Description:   clientSecretDesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		ConflictsWith: []string{"client_secret_wo"},
		ExactlyOneOf:  []string{"client_secret", "client_secret_wo"},
	}
	s["client_secret_wo"] = &schema.Schema{
		Description:   clientSecretWODesc,
		Type:          schema.TypeString,
		Optional:      true,
		Sensitive:     true,
		WriteOnly:     true,
		ConflictsWith: []string{"client_secret"},
		ExactlyOneOf:  []string{"client_secret", "client_secret_wo"},
		RequiredWith:  []string{"client_secret_wo_version"},
	}
	s["client_secret_wo_version"] = &schema.Schema{
		Description:  clientSecretWOVerDesc,
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"client_secret_wo"},
	}

	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: idDesc,
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: nameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"deployment_ids": {
			Description: deploymentIDsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"domain": {
			Description: domainDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"custom_domain": {
			Description: customDomainDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kind": {
			Description: kindDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"secret_store_id": {
			Description: secretStoreIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
