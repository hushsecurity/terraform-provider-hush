package mcp_application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/provider/assignment"
)

const resourceAppDescription = "A resource application: which agents and users get Hush tokens " +
	"for a `hush_resource_server`, and with which of its scopes. An agent acting for a user " +
	"needs both to be assigned."

var resourceAppFields = []string{
	"display_name", "description", "deployment_ids", "enabled", "assign_all", "assignment",
}

func ResourceAppSchema() map[string]*schema.Schema {
	mcp := ResourceSchema()
	s := map[string]*schema.Schema{}
	for _, key := range resourceAppFields {
		s[key] = mcp[key]
	}
	s["deployment_ids"].Description = "The deployments whose agents can get the resource's tokens."
	s["assign_all"].Description = "Grant the resource to every user of the organization. Agents still need an `assignment`."
	s["resource"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The `hush_resource_server`'s name. Fixed at creation.",
	}
	s["scopes"] = &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "The resource's scopes this application grants. Empty grants all, ones added later too.",
	}
	return s
}

func ResourceApp() *schema.Resource {
	return &schema.Resource{
		Description: resourceAppDescription,

		CreateContext: resourceAppCreate,
		ReadContext:   resourceAppRead,
		UpdateContext: resourceAppUpdate,
		DeleteContext: applicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			return assignment.Validate(d, "assignment")
		},
		Schema: ResourceAppSchema(),
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.ResourceApplicationInput{
		DisplayName:   d.Get("display_name").(string),
		DeploymentIDs: setStrings(d, "deployment_ids"),
		Resource:      d.Get("resource").(string),
		Scopes:        setStrings(d, "scopes"),
	}
	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		input.Description = &description
	}
	// Assignments can only be set by a change: created disabled until then.
	assignments := assignment.Expand(d, "assignment")
	all := d.Get("assign_all").(bool)
	enabled := d.Get("enabled").(bool)
	patch := len(assignments) > 0 || all
	input.Enabled = enabled && !patch
	app, err := client.CreateResourceApplication(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(app.ID)
	if patch {
		update := &client.MCPApplicationUpdate{Assignments: &assignments, AssignAll: &all, Enabled: &enabled}
		if _, err := client.UpdateResourceApplication(ctx, c, app.ID, update); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceAppRead(ctx, d, m)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	app, err := client.GetTypedApplication(ctx, m.(*client.Client), d.Id(), "resource")
	if err != nil {
		if client.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	for key, value := range map[string]any{
		"display_name":   app.DisplayName,
		"description":    deref(app.Description),
		"deployment_ids": app.DeploymentIDs,
		"enabled":        app.Enabled,
		"assign_all":     app.AssignAll,
		"assignment":     assignment.Flatten(app.Assignments),
		"resource":       app.Resource,
		"scopes":         app.Scopes,
	} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input, changed := baseUpdate(d)
	if d.HasChange("scopes") {
		scopes := setStrings(d, "scopes")
		input.Scopes = &scopes
		changed = true
	}
	if changed {
		if _, err := client.UpdateResourceApplication(ctx, c, d.Id(), input); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceAppRead(ctx, d, m)
}
