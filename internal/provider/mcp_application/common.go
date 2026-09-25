package mcp_application

import (
	"maps"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const (
	resourceDescription = "An MCP application: an MCP server made available to agents through the " +
		"agent gateways of the deployments it is placed on. It is created from an entry of Hush's " +
		"catalog (see the `hush_mcp_catalog_entry` data source) or from a " +
		"`hush_custom_mcp_application`."
	appCatalogIDDesc = "The catalog entry the application is created from, such as 'github', or " +
		"a custom app's `app_catalog_id` ('custom-<name>'). Fixed at creation."
	nameDesc          = "The name agents see the application under: the catalog id, suffixed with the lowercased `url_label` when one was picked."
	catalogNameDesc   = "The name the catalog shows for the entry."
	displayNameDesc   = "The name the application is shown under. Unique within the organization."
	descriptionDesc   = "A description of the application."
	deploymentIDsDesc = "The deployments whose gateways serve the application. A server Hush hosts " +
		"(the catalog entry's `hosted`) can only be placed on deployments of kind 'hosted'."
	urlLabelDesc = "Which of the entry's addresses to use, by label, when it has more than one " +
		"(the catalog entry's `urls`). Not allowed when it has one. Fixed at creation. It is " +
		"checked at plan time against the custom app as it stands, so an address added to a " +
		"custom app in the same apply that creates this application is refused: apply the " +
		"custom app first."
	enabledDesc       = "Whether the gateways serve the application. A disabled application keeps its configuration."
	allowedAgentsDesc = "Limit the application to these agent types: 'claude-code', 'claude', 'cursor', 'windsurf', 'vscode' and 'openclaw'. Leave unset, or empty, to allow any agent."
	scopesDesc        = "The OAuth scopes to request from the server. Left unset, an application starts with the catalog entry's; removing the attribute later keeps the scopes the application has. An application of a custom app is rewritten whenever the custom app's scopes change, so set scopes on one of the two, not both."
	clientIDDesc      = "The client id of an OAuth app registered with the server. Required for an " +
		"entry with `manual_registration`. An application of a custom app left without one " +
		"inherits the custom app's client id and secret, and follows the custom app when they " +
		"change. Removing it removes the client secret too, except on an application of a " +
		"custom app, where removing it leaves the credentials the application has."
	clientSecretDesc      = "The client secret of that OAuth app. Stored in Terraform state; prefer client_secret_wo. The API never returns it."
	clientSecretWODesc    = "The client secret, kept out of Terraform state. Sent when client_secret_wo_version changes."
	clientSecretWOVerDesc = "Change to re-send client_secret_wo. Required with it: a write-only value is not in state, so without a version a rotated secret would never be sent."
	googleProjectIDDesc   = "The Google Cloud project the server bills its quota to. Required, and only allowed, for the Google Workspace entries: gcalendar, gdocs, gdrive, gmail, gpeople, gsheets and gslides."
	quickbooksDesc        = "The QuickBooks company the server works on. Required, and only allowed, for the 'quickbooks' entry."
	qbCompanyIDDesc       = "The QuickBooks company (realm) id."
	qbSandboxDesc         = "Whether the company is a sandbox one, reached with the Intuit app's development keys."
	urlDesc               = "The address the gateway reaches the server at. Empty for a server Hush hosts."
	hostedDesc            = "Whether Hush hosts the server."
	oauthRelayDesc        = "Whether the server's OAuth flow goes through Hush's relay."
	toolsDesc             = "The tools the application offers, as copied from its catalog entry or custom app."
	toolNameDesc          = "The tool's name."
	toolTypeDesc          = "The tool's class: 'read', 'write' or 'destructive'."
	toolDescDesc          = "What the tool does."
	toolOperationDesc     = "What the gateway does when an agent calls the tool: 'allow', 'block' or 'user_consent'. This is the tool's own setting where it has one and its class default otherwise."
)

var (
	catalogIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	googleProject    = regexp.MustCompile(`^([a-z][a-z0-9-]{4,28}[a-z0-9]|[1-9][0-9]{5,19})$`)
	qbCompanyID      = regexp.MustCompile(`^[0-9]{1,20}$`)
)

func ResourceSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"app_catalog_id": {
			Description:  appCatalogIDDesc,
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(catalogIDPattern, "must be lowercase letters, digits and dashes"),
		},
		"name":                 {Description: nameDesc, Type: schema.TypeString, Computed: true},
		"catalog_display_name": {Description: catalogNameDesc, Type: schema.TypeString, Computed: true},
		"display_name": {
			Description:  displayNameDesc,
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotWhiteSpace,
		},
		"description": {Description: descriptionDesc, Type: schema.TypeString, Optional: true},
		"deployment_ids": {
			Description: deploymentIDsDesc,
			Type:        schema.TypeSet,
			Required:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^dep-`), "deployment_id must start with 'dep-'"),
			},
		},
		"url_label": {
			Description: urlLabelDesc,
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			// A read can only recover the label lowercased, from the name.
			DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
				return strings.EqualFold(old, new)
			},
		},
		"enabled": {Description: enabledDesc, Type: schema.TypeBool, Optional: true, Default: true},
		// No MinItems: an empty set means no limit, exactly as leaving it out
		// does, and it is what an import's generated configuration writes for
		// an application that has none.
		"allowed_agents": {
			Description: allowedAgentsDesc,
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(client.AgentTypes, false),
			},
		},
		"scopes": {
			Description: scopesDesc,
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"client_id": {
			Description: clientIDDesc,
			Type:        schema.TypeString,
			Optional:    true,
			// An application of a custom app takes the custom app's OAuth
			// client when it names none, and heimdall copies the custom app's
			// later changes to it. What a read finds then is inherited, not
			// drift, and "correcting" it to empty would send a null that
			// strips the credentials the application was given.
			DiffSuppressFunc: func(_, _, new string, d *schema.ResourceData) bool {
				return new == "" && strings.HasPrefix(d.Get("app_catalog_id").(string), client.CustomAppCatalogPrefix)
			},
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
		"google_project_id": {
			Description:  googleProjectIDDesc,
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringMatch(googleProject, "must be a Google Cloud project id or number"),
		},
		"quickbooks": {
			Description: quickbooksDesc,
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"company_id": {
					Description:  qbCompanyIDDesc,
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(qbCompanyID, "must be up to 20 digits"),
				},
				"sandbox": {Description: qbSandboxDesc, Type: schema.TypeBool, Optional: true, Default: false},
			}},
		},
		"url":         {Description: urlDesc, Type: schema.TypeString, Computed: true},
		"hosted":      {Description: hostedDesc, Type: schema.TypeBool, Computed: true},
		"oauth_relay": {Description: oauthRelayDesc, Type: schema.TypeBool, Computed: true},
		"tools": {
			Description: toolsDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name":        {Description: toolNameDesc, Type: schema.TypeString, Computed: true},
				"type":        {Description: toolTypeDesc, Type: schema.TypeString, Computed: true},
				"description": {Description: toolDescDesc, Type: schema.TypeString, Computed: true},
				"operation":   {Description: toolOperationDesc, Type: schema.TypeString, Computed: true},
			}},
		},
	}
	maps.Copy(s, assignmentSchema())
	return s
}
