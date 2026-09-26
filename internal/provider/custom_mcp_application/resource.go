package custom_mcp_application

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: customAppCreate,
		ReadContext:   customAppRead,
		UpdateContext: customAppUpdate,
		DeleteContext: customAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customizeDiff,
		Schema:        ResourceSchema(),
	}
}

func customAppCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CustomMCPApplicationInput{
		Name:        d.Get("name").(string),
		DisplayName: d.Get("display_name").(string),
		URLs:        expandURLs(d),
		Scopes:      expandStrings(d.Get("scopes").([]any)),
		OAuthRelay:  d.Get("oauth_relay").(bool),
		Headers:     expandHeaders(d),
		Auth:        expandAuth(d),
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		input.Description = &description
	}
	// Left out when no block is written, so heimdall starts the list empty
	// and the gateway's detection can fill it.
	if tools := expandTools(d); len(tools) > 0 {
		input.Tools = tools
	}
	if v, ok := d.GetOk("client_id"); ok {
		clientID := v.(string)
		input.ClientID = &clientID
		if secret := clientSecret(d); secret != "" {
			input.ClientSecret = &secret
		}
	}

	app, err := client.CreateCustomMCPApplication(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(app.Name)
	return customAppRead(ctx, d, m)
}

func customAppRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	app, err := client.GetCustomMCPApplication(ctx, c, d.Id())
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := flatten(d, app, true); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func customAppUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CustomMCPApplicationUpdate{}
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
	if d.HasChange("url") {
		input.URLs = expandURLs(d)
		changed = true
	}
	if d.HasChange("scopes") {
		scopes := expandStrings(d.Get("scopes").([]any))
		input.Scopes = &scopes
		changed = true
	}
	// A diff on tool only ever comes from written blocks: with none, the
	// attribute is computed and keeps what heimdall holds.
	if d.HasChange("tool") {
		tools := expandTools(d)
		input.Tools = &tools
		changed = true
	}
	if d.HasChange("oauth_relay") {
		relay := d.Get("oauth_relay").(bool)
		input.OAuthRelay = &relay
		changed = true
	}
	if d.HasChange("headers") {
		// An empty map rather than null: heimdall stores whatever it is sent,
		// and a null would leave the app unreadable.
		headers := expandHeaders(d)
		input.Headers = &headers
		changed = true
	}
	if d.HasChange("client_id") {
		input.ClientID = optionalString(d, "client_id")
		changed = true
	}
	// Removing client_id clears the secret with it, so the secret is only sent
	// while there is a client id for it to belong to.
	if d.HasChanges("client_secret", "client_secret_wo_version") && d.Get("client_id").(string) != "" {
		var secret *string
		if s := clientSecret(d); s != "" {
			secret = &s
		}
		input.ClientSecret = &secret
		changed = true
	}
	if d.HasChange("auth") {
		// The credential travels whole: heimdall takes no auth without its
		// secret, so a new user name is sent with the secret it goes with.
		auth := expandAuth(d)
		input.Auth = &auth
		changed = true
	}
	if !changed {
		return customAppRead(ctx, d, m)
	}

	if _, err := client.UpdateCustomMCPApplication(ctx, c, d.Id(), input); err != nil {
		return diag.FromErr(err)
	}
	return customAppRead(ctx, d, m)
}

func customAppDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if err := client.DeleteCustomMCPApplication(ctx, c, d.Id()); err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			if apiErr.IsNotFound() {
				d.SetId("")
				return nil
			}
			if apiErr.IsConflict() {
				return diag.Errorf("custom app %s is still used by applications; "+
					"destroy those hush_mcp_application resources first: %s", d.Id(), apiErr.Detail)
			}
		}
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func expandStrings(values []any) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func expandURLs(d *schema.ResourceData) []client.URLOption {
	blocks := d.Get("url").([]any)
	out := make([]client.URLOption, 0, len(blocks))
	for _, b := range blocks {
		block := b.(map[string]any)
		option := client.URLOption{URL: block["url"].(string)}
		if label := block["label"].(string); label != "" {
			option.Label = &label
		}
		out = append(out, option)
	}
	return out
}

func expandTools(d *schema.ResourceData) []client.CustomMCPTool {
	blocks := d.Get("tool").([]any)
	out := make([]client.CustomMCPTool, 0, len(blocks))
	for _, b := range blocks {
		block := b.(map[string]any)
		out = append(out, client.CustomMCPTool{
			Name:        block["name"].(string),
			Type:        block["type"].(string),
			Description: block["description"].(string),
		})
	}
	return out
}

func expandHeaders(d *schema.ResourceData) map[string]string {
	out := map[string]string{}
	for name, value := range d.Get("headers").(map[string]any) {
		out[name] = value.(string)
	}
	return out
}

// expandAuth returns nil when no block is written, which an update sends as
// null to remove the credential.
func expandAuth(d *schema.ResourceData) map[string]string {
	blocks := d.Get("auth").([]any)
	if len(blocks) == 0 || blocks[0] == nil {
		return nil
	}
	block := blocks[0].(map[string]any)
	secret, _ := block["secret"].(string)
	if secret == "" {
		secret = writeonly.GetNestedString(d, "auth", 0, "secret_wo")
	}
	return client.MCPAuthInput(client.MCPAuth{
		Type:     block["type"].(string),
		Username: block["username"].(string),
		Name:     block["name"].(string),
	}, secret)
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

// flatten writes what heimdall returns. It never returns the client secret or
// the auth secret, so a managed resource leaves those, and the versions that
// go with them, as the configuration last set them; the data source has no
// such attributes at all.
func flatten(d *schema.ResourceData, app *client.CustomMCPApplication, managed bool) error {
	urls := make([]any, 0, len(app.URLs))
	for _, option := range app.URLs {
		label := ""
		if option.Label != nil {
			label = *option.Label
		}
		urls = append(urls, map[string]any{"label": label, "url": option.URL})
	}
	tools := make([]any, 0, len(app.Tools))
	for _, tool := range app.Tools {
		tools = append(tools, map[string]any{
			"name": tool.Name, "type": tool.Type, "description": tool.Description,
		})
	}
	var auth []any
	if app.Auth != nil {
		block := map[string]any{
			"type":     app.Auth.Type,
			"username": app.Auth.Username,
			"name":     app.Auth.Name,
		}
		if managed {
			block["secret"] = d.Get("auth.0.secret")
			block["secret_wo_version"] = d.Get("auth.0.secret_wo_version")
		}
		auth = []any{block}
	}
	description, clientID := "", ""
	if app.Description != nil {
		description = *app.Description
	}
	if app.ClientID != nil {
		clientID = *app.ClientID
	}

	for key, value := range map[string]any{
		"name":           app.Name,
		"app_catalog_id": app.AppCatalogID,
		"display_name":   app.DisplayName,
		"description":    description,
		"url":            urls,
		"scopes":         app.Scopes,
		"tool":           tools,
		"oauth_relay":    app.OAuthRelay,
		"headers":        app.Headers,
		"client_id":      clientID,
		"auth":           auth,
	} {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}
	return nil
}
