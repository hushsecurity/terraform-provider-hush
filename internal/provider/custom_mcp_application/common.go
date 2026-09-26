package custom_mcp_application

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	resourceDescription = "A custom MCP server: one that is not in Hush's catalog, described here so " +
		"`hush_mcp_application` can create applications from it (with `app_catalog_id` set to this " +
		"resource's `app_catalog_id`). Changes to the addresses, tools, scopes, headers and " +
		"credentials carry over to every application created from it. It cannot be destroyed while " +
		"an application still uses it."
	nameDesc = "The custom app's name: lowercase letters, digits and dashes, starting with a letter " +
		"or digit, at most 63 characters. It is fixed at creation."
	appCatalogIDDesc = "The id an application is created from: the name prefixed with 'custom-'."
	displayNameDesc  = "The name the custom app is shown under."
	descriptionDesc  = "A description of the custom app."
	urlDesc          = "An address the server is reachable at. Give several to let each application " +
		"pick one, by `url_label`; each then needs a label. Removing an address an application uses " +
		"is refused."
	urlLabelDesc = "The label that names the address: letters, digits, '_' and '-', at most 32 " +
		"characters. Required on every address when there is more than one."
	urlURLDesc   = "The address of the server's MCP endpoint."
	scopesDesc   = "The OAuth scopes to request from the server."
	toolDesc     = "A tool the server offers. The gateway can also detect them from the server itself; when no block is written, whatever the custom app holds is left alone. The operation the gateway applies to a tool is set per application, by `tool_operation`."
	toolNameDesc = "The tool's name, as the server reports it."
	toolTypeDesc = "The tool's class: 'read', 'write' or 'destructive'. Each application applies the class's default operation unless it sets one for the tool."
	toolDescDesc = "What the tool does."
	oauthDesc    = "Whether the server's OAuth flow goes through Hush's relay."
	headersDesc  = "Extra HTTP headers the gateway sends with every request to the server. " +
		"Names the gateway owns (Authorization, Host, Content-Type and the like, and any starting " +
		"with 'proxy-', 'x-hush-' or 'hush-') are refused."
	clientIDDesc = "The client id of an OAuth app registered with the server, for servers that do " +
		"not register clients on their own. Removing it removes the client secret too."
	clientSecretDesc      = "The client secret of that OAuth app. Stored in Terraform state; prefer client_secret_wo. The API never returns it."
	clientSecretWODesc    = "The client secret, kept out of Terraform state. Sent when client_secret_wo_version changes."
	clientSecretWOVerDesc = "Change to re-send client_secret_wo. Required with it: a write-only value is not in state, so without a version a rotated secret would never be sent."
	authDesc              = "A fixed credential the gateway presents to the server instead of, or next to, OAuth."
	authTypeDesc          = "'bearer' (sent as 'Authorization: Bearer <secret>'), 'basic' (sent as basic auth with `username`) or 'header' (sent as header `name`). 'bearer' and 'basic' use the Authorization header, which OAuth also needs, so they cannot be combined with `client_id`."
	authUsernameDesc      = "The user name, for type 'basic' only."
	authNameDesc          = "The header name, for type 'header' only. It cannot be 'Authorization' (use 'bearer' or 'basic') or a name `headers` already sets."
	authSecretDesc        = "The token, password or header value. Stored in Terraform state; prefer secret_wo. The API never returns it."
	authSecretWODesc      = "The token, password or header value, kept out of Terraform state. Sent when secret_wo_version changes."
	authSecretWOVerDesc   = "Change to re-send secret_wo. Required with it."
)

var (
	namePattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	urlLabelPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
)

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description:  nameDesc,
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(namePattern, "must be lowercase letters, digits and dashes, starting with a letter or digit, at most 63 characters"),
		},
		"app_catalog_id": {
			Description: appCatalogIDDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"display_name": {
			Description:  displayNameDesc,
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotWhiteSpace,
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Optional:    true,
		},
		"url": {
			Description: urlDesc,
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"label": {
					Description:  urlLabelDesc,
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringMatch(urlLabelPattern, "must be letters, digits, '_' and '-', at most 32 characters"),
				},
				"url": {
					Description:  urlURLDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				},
			}},
		},
		"scopes": {
			Description: scopesDesc,
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"tool": {
			Description: toolDesc,
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {Description: toolNameDesc, Type: schema.TypeString, Required: true},
				"type": {
					Description:  toolTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(client.ToolTypes, false),
				},
				"description": {Description: toolDescDesc, Type: schema.TypeString, Optional: true},
			}},
		},
		"oauth_relay": {
			Description: oauthDesc,
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"headers": {
			Description: headersDesc,
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Optional:    true,
		},
		"client_secret": {
			Description:   clientSecretDesc,
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			ConflictsWith: []string{"client_secret_wo"},
			RequiredWith:  []string{"client_id"},
		},
		"client_secret_wo": {
			Description:   clientSecretWODesc,
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			WriteOnly:     true,
			ConflictsWith: []string{"client_secret"},
			RequiredWith:  []string{"client_id", "client_secret_wo_version"},
		},
		"client_secret_wo_version": {
			Description:  clientSecretWOVerDesc,
			Type:         schema.TypeString,
			Optional:     true,
			RequiredWith: []string{"client_secret_wo"},
		},
		"auth": {
			Description: authDesc,
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"type": {
					Description:  authTypeDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(client.MCPAuthTypes, false),
				},
				"username": {Description: authUsernameDesc, Type: schema.TypeString, Optional: true},
				"name":     {Description: authNameDesc, Type: schema.TypeString, Optional: true},
				"secret": {
					Description: authSecretDesc,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
				},
				"secret_wo": {
					Description: authSecretWODesc,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					WriteOnly:   true,
				},
				"secret_wo_version": {
					Description: authSecretWOVerDesc,
					Type:        schema.TypeString,
					Optional:    true,
				},
			}},
		},
	}
}
