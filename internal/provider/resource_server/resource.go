package resource_server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "A resource server: an API behind your own gateway that agents get " +
	"Hush tokens for. A `hush_resource_application` says which agents and users get them, and " +
	"with which scopes."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique name, used by `hush_resource_application.resource`. Fixed at creation.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description.",
			},
			"audience": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The API's URI, the `aud` of its tokens, as your gateway checks it. " +
					"Unique in the organization.",
			},
			"scope": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A scope the API supports. To drop one an application grants, remove it from the application first.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Scope name, as it appears in the token's `scope`.",
					},
					"description": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Description.",
					},
				}},
			},
			"token_lifetime": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 24*3600),
				Description:  "Token lifetime in seconds, 60 to 86400. Default 3600.",
			},
			"token_claim": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    20,
				Description: "A claim added to the API's tokens from the user's profile.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Claim name in the token.",
					},
					"attribute": {
						Type:     schema.TypeString,
						Required: true,
						Description: "The user's attribute: a profile field (e.g. `email`), `groups`, " +
							"or a claim from the IdP.",
					},
				}},
			},
		},
	}
}

func expand(d *schema.ResourceData) *client.ResourceServer {
	rs := &client.ResourceServer{
		Name:            d.Get("name").(string),
		DisplayName:     d.Get("display_name").(string),
		Audience:        d.Get("audience").(string),
		ScopesSupported: []client.ResourceScope{},
		TokenClaims:     []client.TokenClaim{},
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		rs.Description = &description
	}
	if v, ok := d.GetOk("token_lifetime"); ok {
		lifetime := v.(int)
		rs.TokenLifetime = &lifetime
	}
	for _, item := range d.Get("scope").([]any) {
		scope := item.(map[string]any)
		s := client.ResourceScope{Name: scope["name"].(string)}
		if description := scope["description"].(string); description != "" {
			s.Description = &description
		}
		rs.ScopesSupported = append(rs.ScopesSupported, s)
	}
	for _, item := range d.Get("token_claim").([]any) {
		claim := item.(map[string]any)
		rs.TokenClaims = append(rs.TokenClaims, client.TokenClaim{
			Name: claim["name"].(string), Attribute: claim["attribute"].(string),
		})
	}
	return rs
}

func flatten(d *schema.ResourceData, rs *client.ResourceServer) error {
	scopes := []any{}
	for _, s := range rs.ScopesSupported {
		scopes = append(scopes, map[string]any{"name": s.Name, "description": deref(s.Description)})
	}
	claims := []any{}
	for _, c := range rs.TokenClaims {
		claims = append(claims, map[string]any{"name": c.Name, "attribute": c.Attribute})
	}
	lifetime := 0
	if rs.TokenLifetime != nil {
		lifetime = *rs.TokenLifetime
	}
	for key, value := range map[string]any{
		"name":           rs.Name,
		"display_name":   rs.DisplayName,
		"description":    deref(rs.Description),
		"audience":       rs.Audience,
		"scope":          scopes,
		"token_lifetime": lifetime,
		"token_claim":    claims,
	} {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rs, err := client.CreateResourceServer(ctx, m.(*client.Client), expand(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(rs.Name)
	return diag.FromErr(flatten(d, rs))
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rs, err := client.GetResourceServer(ctx, m.(*client.Client), d.Id())
	if err != nil {
		if client.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	return diag.FromErr(flatten(d, rs))
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	rs, err := client.UpdateResourceServer(ctx, m.(*client.Client), expand(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(flatten(d, rs))
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	err := client.DeleteResourceServer(ctx, m.(*client.Client), d.Id())
	if err != nil && !client.IsNotFoundError(err) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
