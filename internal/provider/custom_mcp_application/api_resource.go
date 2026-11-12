package custom_mcp_application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func APIResource() *schema.Resource {
	return &schema.Resource{
		Description: apiResourceDescription,

		CreateContext: customAPIAppCreate,
		ReadContext:   customAPIAppRead,
		UpdateContext: customAPIAppUpdate,
		DeleteContext: customAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: apiCustomizeDiff,
		Schema:        APIResourceSchema(),
	}
}

func apiCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if err := validateHeaders(d); err != nil {
		return err
	}
	if err := validateAuth(d); err != nil {
		return err
	}
	if d.Id() != "" && d.HasChanges("url", "spec_url") {
		if err := d.SetNewComputed("detect_required"); err != nil {
			return err
		}
	}
	return validateAPITools(d)
}

// validateAPITools refuses what heimdall's tools PUT refuses: a name or
// operation_id listed twice, and any operation_id on a new app (or a
// replacement), which has no tools for one to name.
func validateAPITools(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("tool") {
		return nil
	}
	creating := d.Id() == "" || d.HasChange("name")
	names, ids := map[string]bool{}, map[string]bool{}
	for i, v := range d.Get("tool").([]any) {
		block, _ := v.(map[string]any)
		name, _ := block["name"].(string)
		id, _ := block["operation_id"].(string)
		switch {
		case name != "" && names[name]:
			return fmt.Errorf("tool[%d]: name %q is listed more than once", i, name)
		case id != "" && ids[id]:
			return fmt.Errorf("tool[%d]: operation_id %q is listed more than once", i, id)
		case id != "" && creating:
			return fmt.Errorf("tool[%d]: operation_id %q can only name a tool already detected "+
				"on this app: create it without operation_id, detect in the console, then import", i, id)
		}
		names[name], ids[id] = true, true
	}
	return nil
}

func customAPIAppCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CustomAPIApplicationInput{
		Name:        d.Get("name").(string),
		DisplayName: d.Get("display_name").(string),
		URL:         d.Get("url").(string),
		Headers:     expandHeaders(d),
		Auth:        expandAuth(d),
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		input.Description = &description
	}
	if v, ok := d.GetOk("spec_url"); ok {
		specURL := v.(string)
		input.SpecURL = &specURL
	}

	app, err := client.CreateCustomAPIApplication(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(app.Name)
	if tools := expandAPITools(d); len(tools) > 0 {
		if err := setAPITools(ctx, c, d.Id(), tools); err != nil {
			return diag.FromErr(err)
		}
	}
	return customAPIAppRead(ctx, d, m)
}

func customAPIAppRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	app, err := client.GetCustomAPIApplication(ctx, c, d.Id())
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	tools, err := client.GetCustomAPIApplicationTools(ctx, c, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := flattenAPI(d, app, tools, true); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func customAPIAppUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CustomAPIApplicationUpdate{}
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
		url := d.Get("url").(string)
		input.URL = &url
		changed = true
	}
	if d.HasChange("spec_url") {
		input.SpecURL = optionalString(d, "spec_url")
		changed = true
	}
	if d.HasChange("headers") {
		headers := expandHeaders(d)
		input.Headers = &headers
		changed = true
	}
	if d.HasChange("auth") {
		auth := expandAuth(d)
		input.Auth = &auth
		changed = true
	}
	if changed {
		if _, err := client.UpdateCustomAPIApplication(ctx, c, d.Id(), input); err != nil {
			return diag.FromErr(err)
		}
	}
	// Only written blocks diff: with none, tool is computed.
	if d.HasChange("tool") {
		if err := setAPITools(ctx, c, d.Id(), expandAPITools(d)); err != nil {
			return diag.FromErr(err)
		}
	}
	return customAPIAppRead(ctx, d, m)
}

func setAPITools(ctx context.Context, c *client.Client, name string, tools []client.APITool) error {
	err := client.SetCustomAPIApplicationTools(ctx, c, name, tools)
	var apiErr *client.APIError
	// heimdall also answers 409 while the console holds the app's tools.
	if errors.As(err, &apiErr) && apiErr.IsConflict() &&
		!slices.ContainsFunc(tools, func(t client.APITool) bool { return t.Enabled }) {
		return fmt.Errorf("tool: applications use custom app %s, so at least one tool must "+
			"be enabled: %w", name, err)
	}
	return err
}

func expandAPITools(d *schema.ResourceData) []client.APITool {
	blocks := d.Get("tool").([]any)
	out := make([]client.APITool, 0, len(blocks))
	for _, b := range blocks {
		block := b.(map[string]any)
		tool := client.APITool{
			Name:        block["name"].(string),
			Title:       block["title"].(string),
			Description: block["description"].(string),
			Type:        block["type"].(string),
			Enabled:     block["enabled"].(bool),
			Method:      block["method"].(string),
			Path:        block["path"].(string),
			InputSchema: json.RawMessage(block["input_schema"].(string)),
			ContentType: block["content_type"].(string),
		}
		if id := block["operation_id"].(string); id != "" {
			tool.OperationID = &id
		}
		out = append(out, tool)
	}
	return out
}

func flattenAPI(d *schema.ResourceData, app *client.CustomAPIApplication, apiTools []client.APITool, managed bool) error {
	tools := make([]any, 0, len(apiTools))
	for _, tool := range apiTools {
		operationID := ""
		if tool.OperationID != nil {
			operationID = *tool.OperationID
		}
		tools = append(tools, map[string]any{
			"name":         tool.Name,
			"title":        tool.Title,
			"description":  tool.Description,
			"type":         tool.Type,
			"enabled":      tool.Enabled,
			"method":       tool.Method,
			"path":         tool.Path,
			"input_schema": string(tool.InputSchema),
			"content_type": tool.ContentType,
			"operation_id": operationID,
		})
	}
	description, specURL := "", ""
	if app.Description != nil {
		description = *app.Description
	}
	if app.SpecURL != nil {
		specURL = *app.SpecURL
	}

	for key, value := range map[string]any{
		"name":            app.Name,
		"app_catalog_id":  app.AppCatalogID,
		"display_name":    app.DisplayName,
		"description":     description,
		"url":             app.URL,
		"spec_url":        specURL,
		"detect_required": app.DetectRequired,
		"headers":         app.Headers,
		"auth":            flattenAuth(d, app.Auth, managed),
		"tool":            tools,
	} {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}
	return nil
}
