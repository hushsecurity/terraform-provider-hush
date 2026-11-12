package mcp_application

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/assignment"
)

const (
	apiResourceDescription = "An API application: a `hush_custom_api_application`'s enabled tools " +
		"made available to agents through the agent gateways of the deployments it is placed on. " +
		"The address, headers and credential are the custom app's."
	apiAppCatalogIDDesc = "The custom API app's `app_catalog_id` ('custom-<name>'). It needs an " +
		"enabled tool. Fixed at creation."
	apiNameDesc          = "The name agents see the application under: its `app_catalog_id`."
	apiURLDesc           = "The API's address, from the custom app."
	apiToolsDesc         = "The custom app's enabled tools."
	apiDeploymentIDsDesc = "The deployments whose gateways serve the application."
	apiToolOperationDesc = "The operation for one tool, overriding its class default. The " +
		"tool must be one of the application's (see `tools`). A tool without a block takes its " +
		"class default, so removing a block returns the tool to it, and an override set " +
		"elsewhere, such as in the console, is removed on the next apply."
)

// The fields every application has; the rest are MCP-only.
var apiFields = []string{
	"app_catalog_id", "name", "catalog_display_name", "display_name", "description",
	"deployment_ids", "enabled", "allowed_agents", "url", "tools",
	"assign_all", "assignment", "tool_defaults", "tool_operation", "rate_limit",
}

func APIResourceSchema() map[string]*schema.Schema {
	mcp := ResourceSchema()
	s := map[string]*schema.Schema{}
	for _, key := range apiFields {
		s[key] = mcp[key]
	}
	s["app_catalog_id"].Description = apiAppCatalogIDDesc
	s["name"].Description = apiNameDesc
	s["url"].Description = apiURLDesc
	s["tools"].Description = apiToolsDesc
	s["deployment_ids"].Description = apiDeploymentIDsDesc
	s["tool_operation"].Description = apiToolOperationDesc
	return s
}

func APIResource() *schema.Resource {
	return &schema.Resource{
		Description: apiResourceDescription,

		CreateContext: apiApplicationCreate,
		ReadContext:   apiApplicationRead,
		UpdateContext: apiApplicationUpdate,
		DeleteContext: applicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			if err := assignment.Validate(d, "assignment"); err != nil {
				return err
			}
			if err := validateToolOperations(d); err != nil {
				return err
			}
			if err := validateRateLimits(d); err != nil {
				return err
			}
			return planToolsRecompute(d)
		},
		Schema: APIResourceSchema(),
	}
}

func apiApplicationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Created disabled and enabled by the patch, as an MCP application.
	input := &client.MCPApplicationInput{
		DisplayName:   d.Get("display_name").(string),
		DeploymentIDs: setStrings(d, "deployment_ids"),
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		input.Description = &description
	}
	if _, ok := d.GetOk("allowed_agents"); ok {
		input.AllowedAgents = setStrings(d, "allowed_agents")
	}
	app, err := client.CreateAPIApplication(ctx, c, d.Get("app_catalog_id").(string), input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(app.ID)

	update := &client.MCPApplicationUpdate{}
	patch := false
	if d.Get("enabled").(bool) {
		enabled := true
		update.Enabled = &enabled
		patch = true
	}
	if assignments := assignment.Expand(d, "assignment"); len(assignments) > 0 {
		update.Assignments = &assignments
		patch = true
	}
	if d.Get("assign_all").(bool) {
		all := true
		update.AssignAll = &all
		patch = true
	}
	if patch {
		if _, err := client.UpdateAPIApplication(ctx, c, app.ID, update); err != nil {
			return explainGatewayNotReady(err)
		}
	}
	if err := applyToolOperations(ctx, c, d, "api", true); err != nil {
		return diag.FromErr(err)
	}
	if err := applyRateLimits(ctx, c, d, "api"); err != nil {
		return diag.FromErr(err)
	}
	return apiApplicationRead(ctx, d, m)
}

func apiApplicationRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	app, err := client.GetTypedApplication(ctx, c, d.Id(), "api")
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := flattenAPI(d, app); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func apiApplicationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	if input, changed := baseUpdate(d); changed {
		if _, err := client.UpdateAPIApplication(ctx, c, d.Id(), input); err != nil {
			return explainGatewayNotReady(err)
		}
	}
	if err := applyToolOperations(ctx, c, d, "api", false); err != nil {
		return diag.FromErr(err)
	}
	if err := applyRateLimits(ctx, c, d, "api"); err != nil {
		return diag.FromErr(err)
	}
	return apiApplicationRead(ctx, d, m)
}

func flattenAPI(d *schema.ResourceData, app *client.MCPApplication) error {
	for key, value := range map[string]any{
		"app_catalog_id":       app.AppCatalogID,
		"name":                 app.Name,
		"catalog_display_name": deref(app.CatalogDisplayName),
		"display_name":         app.DisplayName,
		"description":          deref(app.Description),
		"deployment_ids":       app.DeploymentIDs,
		"enabled":              app.Enabled,
		"allowed_agents":       app.AllowedAgents,
		"url":                  deref(app.URL),
		"tools":                flattenTools(app),
		"assign_all":           app.AssignAll,
		"assignment":           assignment.Flatten(app.Assignments),
		"tool_defaults":        flattenToolDefaults(app.ToolGroups),
		"tool_operation":       flattenToolOperations(app.Tools),
		"rate_limit":           flattenRateLimits(app.RateLimits),
	} {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}
	return nil
}
