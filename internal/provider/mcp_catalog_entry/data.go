package mcp_catalog_entry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	dataSourceDescription = "An entry of Hush's built-in MCP catalog: a server a `hush_mcp_application` " +
		"can be created from. Read it to learn, before applying, what the application will need -- " +
		"which `url_label` to pick, whether it takes `client_id` and `client_secret`, whether it runs " +
		"on hosted gateways only, and the tool names `tool_operation` may set. Custom apps are not in " +
		"the catalog; a `hush_custom_mcp_application` describes one of those itself."
	appCatalogIDDesc       = "The id of the catalog entry, such as 'github'."
	displayNameDesc        = "The name the catalog shows for the server."
	categoryDesc           = "The catalog category the server is listed under."
	hostedDesc             = "Whether Hush hosts the server itself. A hosted server is reachable only from gateways Hush runs, so an application made from it can only be placed on deployments of kind 'hosted'."
	manualRegistrationDesc = "Whether the server does not register OAuth clients on its own, so an application made from it needs `client_id` and `client_secret` from an OAuth app you registered with the provider."
	oauthRelayDesc         = "Whether the server's OAuth flow goes through Hush's relay."
	scopesDesc             = "The OAuth scopes an application made from the entry requests by default."
	urlsDesc               = "The addresses the server is reachable at. When there is more than one, an application picks one by its label through `url_label`. Empty for a hosted server, whose address is Hush's to choose."
	urlLabelDesc           = "The label that names the address, or empty when the entry has a single one."
	urlURLDesc             = "The address of the server's MCP endpoint."
	toolsDesc              = "The tools the server offers, which an application copies when it is created."
	toolNameDesc           = "The tool's name, as `tool_operation` refers to it."
	toolTypeDesc           = "The tool's class -- 'read', 'write' or 'destructive' -- whose default operation it takes unless an application sets one of its own."
	toolDescriptionDesc    = "What the tool does."
)

func DataSource() *schema.Resource {
	return &schema.Resource{
		Description: dataSourceDescription,
		ReadContext: catalogEntryRead,
		Schema: map[string]*schema.Schema{
			"app_catalog_id": {
				Description: appCatalogIDDesc,
				Type:        schema.TypeString,
				Required:    true,
			},
			"display_name":        {Description: displayNameDesc, Type: schema.TypeString, Computed: true},
			"category":            {Description: categoryDesc, Type: schema.TypeString, Computed: true},
			"hosted":              {Description: hostedDesc, Type: schema.TypeBool, Computed: true},
			"manual_registration": {Description: manualRegistrationDesc, Type: schema.TypeBool, Computed: true},
			"oauth_relay":         {Description: oauthRelayDesc, Type: schema.TypeBool, Computed: true},
			"scopes": {
				Description: scopesDesc,
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"urls": {
				Description: urlsDesc,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"label": {Description: urlLabelDesc, Type: schema.TypeString, Computed: true},
					"url":   {Description: urlURLDesc, Type: schema.TypeString, Computed: true},
				}},
			},
			"tools": {
				Description: toolsDesc,
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name":        {Description: toolNameDesc, Type: schema.TypeString, Computed: true},
					"type":        {Description: toolTypeDesc, Type: schema.TypeString, Computed: true},
					"description": {Description: toolDescriptionDesc, Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func catalogEntryRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	entry, err := client.GetMCPCatalogEntry(ctx, c, d.Get("app_catalog_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	urls := make([]any, 0, len(entry.URLs))
	for _, option := range entry.URLs {
		label := ""
		if option.Label != nil {
			label = *option.Label
		}
		urls = append(urls, map[string]any{"label": label, "url": option.URL})
	}
	tools := make([]any, 0, len(entry.Tools))
	for _, tool := range entry.Tools {
		tools = append(tools, map[string]any{
			"name": tool.Name, "type": tool.Type, "description": tool.Description,
		})
	}

	d.SetId(entry.AppCatalogID)
	for key, value := range map[string]any{
		"display_name":        entry.DisplayName,
		"category":            entry.Category,
		"hosted":              entry.Hosted,
		"manual_registration": entry.ManualRegistration,
		"oauth_relay":         entry.OAuthRelay,
		"scopes":              entry.Scopes,
		"urls":                urls,
		"tools":               tools,
	} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
