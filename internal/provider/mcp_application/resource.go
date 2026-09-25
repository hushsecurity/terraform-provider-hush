package mcp_application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: applicationCreate,
		ReadContext:   applicationRead,
		UpdateContext: applicationUpdate,
		DeleteContext: applicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customizeDiff,
		Schema:        ResourceSchema(),
	}
}

func applicationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)
	appCatalogID := d.Get("app_catalog_id").(string)

	input := &client.MCPApplicationInput{
		DisplayName:   d.Get("display_name").(string),
		DeploymentIDs: setStrings(d, "deployment_ids"),
		// Always created disabled, and enabled by the patch that follows. An
		// enabled application encrypts its client secret to the gateway's key,
		// and heimdall saves the application before it finds a gateway that
		// has not published one yet. Failing then, on the create, would leave
		// an application behind that this resource has no id for, and every
		// later apply would be refused as a duplicate. Failing on the patch
		// leaves one Terraform tracks, and retries.
		Enabled: false,
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		input.Description = &description
	}
	if _, ok := d.GetOk("allowed_agents"); ok {
		input.AllowedAgents = setStrings(d, "allowed_agents")
	}
	if v, ok := d.GetOk("url_label"); ok {
		label := v.(string)
		input.URLLabel = &label
	}
	if v, ok := d.GetOk("client_id"); ok {
		clientID := v.(string)
		input.ClientID = &clientID
		if secret := clientSecret(d); secret != "" {
			input.ClientSecret = &secret
		}
	}
	if v, ok := d.GetOk("google_project_id"); ok {
		project := v.(string)
		input.GoogleProjectID = &project
	}
	if company, sandbox, ok := quickbooks(d); ok {
		input.QuickBooksCompanyID = &company
		input.QuickBooksSandbox = &sandbox
	}

	app, err := client.CreateMCPApplication(ctx, c, appCatalogID, input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(app.ID)

	// The create takes neither scopes nor assignments: an application starts
	// with its entry's scopes and granted to no one. Written ones are patched
	// in straight after, with enabled.
	update := &client.MCPApplicationUpdate{}
	patch := false
	if d.Get("enabled").(bool) {
		enabled := true
		update.Enabled = &enabled
		patch = true
	}
	if scopesConfigured(d) {
		scopes := listStrings(d, "scopes")
		update.Scopes = &scopes
		patch = true
	}
	if assignments := expandAssignments(d); len(assignments) > 0 {
		update.Assignments = &assignments
		patch = true
	}
	if d.Get("assign_all").(bool) {
		all := true
		update.AssignAll = &all
		patch = true
	}
	if patch {
		if _, err := client.UpdateMCPApplication(ctx, c, app.ID, appCatalogID, update); err != nil {
			return explainGatewayNotReady(err)
		}
	}
	if err := applyToolOperations(ctx, c, d, true); err != nil {
		return diag.FromErr(err)
	}
	return applicationRead(ctx, d, m)
}

func applicationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	appCatalogID := d.Get("app_catalog_id").(string)
	var app *client.MCPApplication
	var err error
	if appCatalogID == "" {
		// An import knows the id alone, and the id alone does not say which
		// route the application is read through.
		app, err = client.GetApplication(ctx, c, d.Id())
		if err == nil {
			if app.Type != "mcp" || app.AppCatalogID == "" {
				return diag.Errorf("application %s is a %s application: only MCP applications "+
					"can be managed by hush_mcp_application", d.Id(), app.Type)
			}
			appCatalogID = app.AppCatalogID
		}
	}
	if err == nil {
		app, err = client.GetMCPApplication(ctx, c, d.Id(), appCatalogID)
	}
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := flatten(d, app); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func applicationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.MCPApplicationUpdate{}
	changed := false
	if d.HasChange("display_name") {
		name := d.Get("display_name").(string)
		input.DisplayName = &name
		changed = true
	}
	if d.HasChange("description") {
		input.Description = optionalString(d, "description")
		changed = true
	}
	if d.HasChange("deployment_ids") {
		ids := setStrings(d, "deployment_ids")
		input.DeploymentIDs = &ids
		changed = true
	}
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		input.Enabled = &enabled
		changed = true
	}
	if d.HasChange("allowed_agents") {
		// null lifts the limit; an empty list would allow no agent at all,
		// which heimdall refuses.
		var agents *[]string
		if _, ok := d.GetOk("allowed_agents"); ok {
			list := setStrings(d, "allowed_agents")
			agents = &list
		}
		input.AllowedAgents = &agents
		changed = true
	}
	if d.HasChange("scopes") {
		scopes := listStrings(d, "scopes")
		input.Scopes = &scopes
		changed = true
	}
	if d.HasChange("assignment") {
		assignments := expandAssignments(d)
		input.Assignments = &assignments
		changed = true
	}
	if d.HasChange("assign_all") {
		all := d.Get("assign_all").(bool)
		input.AssignAll = &all
		changed = true
	}
	if d.HasChange("client_id") {
		input.ClientID = optionalString(d, "client_id")
		changed = true
		// Removing the client id is only complete with its secret: heimdall
		// keeps a stored secret until it is sent a null one.
		if d.Get("client_id").(string) == "" {
			var none *string
			input.ClientSecret = &none
		}
	}
	if d.HasChanges("client_secret", "client_secret_wo_version") && d.Get("client_id").(string) != "" {
		var secret *string
		if s := clientSecret(d); s != "" {
			secret = &s
		}
		input.ClientSecret = &secret
		changed = true
	}
	if d.HasChange("google_project_id") {
		project := d.Get("google_project_id").(string)
		input.GoogleProjectID = &project
		changed = true
	}
	if d.HasChange("quickbooks") {
		if company, sandbox, ok := quickbooks(d); ok {
			input.QuickBooksCompanyID = &company
			input.QuickBooksSandbox = &sandbox
			changed = true
		}
	}

	if changed {
		appCatalogID := d.Get("app_catalog_id").(string)
		if _, err := client.UpdateMCPApplication(ctx, c, d.Id(), appCatalogID, input); err != nil {
			return explainGatewayNotReady(err)
		}
	}
	if err := applyToolOperations(ctx, c, d, false); err != nil {
		return diag.FromErr(err)
	}
	return applicationRead(ctx, d, m)
}

func applicationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if err := client.DeleteApplication(ctx, c, d.Id()); err != nil {
		var apiErr *client.APIError
		if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}

// explainGatewayNotReady adds what to do to the error heimdall answers when an
// enabled application with a client secret is placed on a gateway that has
// not published the key the secret is encrypted to. A gateway publishes it
// once it runs, so a deployment declared in the same configuration hits this
// on its first apply.
//
// heimdall stores the patch before it finds the key missing, so the next
// refresh reads the application as enabled although no gateway serves it.
// On a create that leaves the resource tainted, and the next apply replaces
// it; on an update it does not, which the detail says.
func explainGatewayNotReady(err error) diag.Diagnostics {
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable ||
		!strings.Contains(apiErr.Detail, "no public key") {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  err.Error(),
		Detail: "A deployment's gateway publishes the key client secrets are encrypted to once " +
			"it is running, and until then an application with a client secret cannot be " +
			"enabled on it. Apply again once the gateway is up. The API may have recorded " +
			"the change regardless, so if the next plan shows no change, set enabled = false, " +
			"apply, and set it back.",
	}}
}

// scopesConfigured reports whether the configuration writes scopes at all,
// an empty list included -- which GetOk cannot tell from an unset attribute,
// and which asks for no scopes rather than the entry's.
func scopesConfigured(d *schema.ResourceData) bool {
	return !d.GetRawConfig().GetAttr("scopes").IsNull()
}

func setStrings(d *schema.ResourceData, key string) []string {
	values := d.Get(key).(*schema.Set).List()
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, v.(string))
	}
	return out
}

func listStrings(d *schema.ResourceData, key string) []string {
	values := d.Get(key).([]any)
	out := make([]string, 0, len(values))
	for _, v := range values {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func quickbooks(d *schema.ResourceData) (company string, sandbox bool, ok bool) {
	blocks := d.Get("quickbooks").([]any)
	if len(blocks) == 0 || blocks[0] == nil {
		return "", false, false
	}
	block := blocks[0].(map[string]any)
	return block["company_id"].(string), block["sandbox"].(bool), true
}

func clientSecret(d *schema.ResourceData) string {
	return writeonly.GetString(d, "client_secret", "client_secret_wo")
}

// optionalString is a patch value for a field that can be cleared: the
// configured value, or null when there is none.
func optionalString(d *schema.ResourceData, key string) **string {
	var value *string
	if s := d.Get(key).(string); s != "" {
		value = &s
	}
	return &value
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// flatten writes what heimdall returns. It never returns the client secret,
// so that and its version are left as the configuration last set them.
//
// url_label is not returned either. It survives in the name, lowercased, as
// its suffix, which is all an import has to go on; the schema compares labels
// without regard to case, so that is enough to keep a configured label from
// planning a replacement.
func flatten(d *schema.ResourceData, app *client.MCPApplication) error {
	if d.Get("url_label").(string) == "" {
		if label, ok := strings.CutPrefix(app.Name, app.AppCatalogID+"-"); ok {
			if err := d.Set("url_label", label); err != nil {
				return err
			}
		}
	}
	defaults := map[string]string{}
	for _, group := range app.ToolGroups {
		defaults[group.Type] = group.Operation
	}
	tools := make([]any, 0, len(app.Tools))
	for _, tool := range app.Tools {
		operation := defaults[tool.Type]
		if tool.Operation != nil {
			operation = *tool.Operation
		}
		tools = append(tools, map[string]any{
			"name": tool.Name, "type": tool.Type, "description": tool.Description, "operation": operation,
		})
	}
	var qb []any
	if app.QuickBooksCompanyID != nil {
		qb = []any{map[string]any{
			"company_id": *app.QuickBooksCompanyID,
			"sandbox":    deref(app.QuickBooksSandbox),
		}}
	}

	values := map[string]any{
		"app_catalog_id":       app.AppCatalogID,
		"name":                 app.Name,
		"catalog_display_name": deref(app.CatalogDisplayName),
		"display_name":         app.DisplayName,
		"description":          deref(app.Description),
		"deployment_ids":       app.DeploymentIDs,
		"enabled":              app.Enabled,
		"assign_all":           app.AssignAll,
		"assignment":           flattenAssignments(app.Assignments),
		"allowed_agents":       app.AllowedAgents,
		"scopes":               app.Scopes,
		"client_id":            deref(app.ClientID),
		"google_project_id":    deref(app.GoogleProjectID),
		"quickbooks":           qb,
		"url":                  deref(app.URL),
		"hosted":               app.Hosted,
		"oauth_relay":          app.OAuthRelay,
		"tools":                tools,
		"tool_defaults":        flattenToolDefaults(app.ToolGroups),
		"tool_operation":       flattenToolOperations(app.Tools),
	}
	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}
	return nil
}
