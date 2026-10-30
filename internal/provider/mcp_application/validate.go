package mcp_application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

// customizeDiff refuses at plan time what heimdall refuses at apply. A value
// another resource supplies is unknown here and reads as empty, so each rule
// skips what it cannot see rather than guess.
func customizeDiff(ctx context.Context, d *schema.ResourceDiff, m any) error {
	if err := validateAssignments(d); err != nil {
		return err
	}
	if err := validateToolOperations(d); err != nil {
		return err
	}
	if err := validateRateLimits(d); err != nil {
		return err
	}
	if err := planToolsRecompute(d); err != nil {
		return err
	}
	if err := validateAuth(d); err != nil {
		return err
	}
	if err := forceNewOnAuthRemoval(d); err != nil {
		return err
	}
	if !d.NewValueKnown("app_catalog_id") {
		return nil
	}
	appCatalogID := d.Get("app_catalog_id").(string)
	if err := validateEntrySettings(d, appCatalogID); err != nil {
		return err
	}
	// url_label and the credentials are only judged against the entry at
	// creation or when auth changes: url_label cannot change afterwards, and
	// a credential the entry needs was already given or the application would
	// not exist.
	if (d.Id() != "" && !d.HasChange("auth")) || m == nil {
		return nil
	}
	return validateAgainstEntry(ctx, d, m.(*client.Client), appCatalogID)
}

// The Google and QuickBooks entries take a setting of their own, which their
// routes require and every other route refuses.
func validateEntrySettings(d *schema.ResourceDiff, appCatalogID string) error {
	google := client.IsGoogleMCPCatalogID(appCatalogID)
	hasProject := writeonly.IsSet(d, "google_project_id")
	switch {
	case google && !hasProject:
		return fmt.Errorf("google_project_id: required for %q", appCatalogID)
	case !google && hasProject && d.NewValueKnown("google_project_id"):
		return fmt.Errorf("google_project_id: only valid for the Google Workspace entries (%s)",
			strings.Join(client.GoogleMCPCatalogIDs, ", "))
	}

	quickbooks := appCatalogID == client.QuickBooksMCPCatalogID
	if !d.NewValueKnown("quickbooks") {
		return nil
	}
	hasCompany := len(d.Get("quickbooks").([]any)) > 0
	switch {
	case quickbooks && !hasCompany:
		return fmt.Errorf("quickbooks: a quickbooks block is required for %q", appCatalogID)
	case !quickbooks && hasCompany:
		return fmt.Errorf("quickbooks: only valid for %q", client.QuickBooksMCPCatalogID)
	}
	return nil
}

// entryShape is what the checks need of a catalog entry or a custom app.
type entryShape struct {
	labels             []string // one per address; empty strings for unlabelled ones
	manualRegistration bool
	tools              []string // the tools an application of it starts with
	authTypes          []string
	authRequired       bool
}

// validateAgainstEntry reads the entry the application is created from. An
// entry that cannot be read is left to the apply -- a custom app declared in
// the same configuration does not exist yet -- except that a built-in id the
// catalog does not have is refused outright, since nothing will create it.
func validateAgainstEntry(ctx context.Context, d *schema.ResourceDiff, c *client.Client, appCatalogID string) error {
	var shape entryShape
	if name, custom := strings.CutPrefix(appCatalogID, client.CustomAppCatalogPrefix); custom {
		app, err := client.GetCustomMCPApplication(ctx, c, name)
		if err != nil {
			// A 404 is expected: a custom app declared in the same
			// configuration does not exist until the apply. Any other error
			// is skipped as well; heimdall checks again at apply, so a flaky
			// read here costs only the earlier message, not the outcome.
			return nil
		}
		for _, option := range app.URLs {
			shape.labels = append(shape.labels, labelOf(option))
		}
		for _, tool := range app.Tools {
			shape.tools = append(shape.tools, tool.Name)
		}
	} else {
		entry, err := client.GetMCPCatalogEntry(ctx, c, appCatalogID)
		if err != nil {
			var apiErr *client.APIError
			if errors.As(err, &apiErr) && apiErr.IsNotFound() {
				return fmt.Errorf("app_catalog_id: the catalog has no MCP entry %q", appCatalogID)
			}
			// Any other error only costs the earlier message: heimdall checks
			// again at apply.
			return nil
		}
		for _, option := range entry.URLs {
			shape.labels = append(shape.labels, labelOf(option))
		}
		// A hosted entry lists only labelled addresses, so none means one.
		if len(shape.labels) == 0 {
			shape.labels = []string{""}
		}
		shape.manualRegistration = entry.ManualRegistration
		shape.authTypes = entry.AuthTypes
		shape.authRequired = entry.AuthRequired
		for _, tool := range entry.Tools {
			shape.tools = append(shape.tools, tool.Name)
		}
	}

	if err := validateURLLabel(d, appCatalogID, shape.labels); err != nil {
		return err
	}
	if err := validateRateLimitToolNames(d, appCatalogID, shape.tools); err != nil {
		return err
	}
	if err := validateToolNames(d, appCatalogID, shape.tools); err != nil {
		return err
	}
	if err := validateAuthType(d, appCatalogID, shape); err != nil {
		return err
	}
	// heimdall refuses the create unless it gets both halves, or auth.
	hasSecret := writeonly.IsSet(d, "client_secret") || writeonly.IsSet(d, "client_secret_wo")
	hasAuth := !d.NewValueKnown("auth") || len(d.Get("auth").([]any)) > 0
	if shape.manualRegistration && !hasAuth && (!writeonly.IsSet(d, "client_id") || !hasSecret) {
		return fmt.Errorf("client_id: %q needs an OAuth app you registered with it; "+
			"set client_id and client_secret or client_secret_wo", appCatalogID)
	}
	return nil
}

func validateURLLabel(d *schema.ResourceDiff, appCatalogID string, labels []string) error {
	if !d.NewValueKnown("url_label") {
		return nil
	}
	label := d.Get("url_label").(string)
	if len(labels) <= 1 {
		if label != "" {
			return fmt.Errorf("url_label: %q has a single address and takes no label", appCatalogID)
		}
		return nil
	}
	for _, candidate := range labels {
		if candidate == label {
			return nil
		}
	}
	if label == "" {
		return fmt.Errorf("url_label: %q has several addresses; pick one of %s",
			appCatalogID, strings.Join(labels, ", "))
	}
	return fmt.Errorf("url_label: %q is not one of %s's labels: %s",
		label, appCatalogID, strings.Join(labels, ", "))
}

func labelOf(option client.URLOption) string {
	if option.Label == nil {
		return ""
	}
	return *option.Label
}

// validateToolNames refuses a tool_operation for a tool the application will
// not have. heimdall refuses it too, but only after the application exists,
// which leaves it tainted and replaced on the next apply for what is a typo.
func validateToolNames(d *schema.ResourceDiff, appCatalogID string, tools []string) error {
	if !d.NewValueKnown("tool_operation") {
		return nil
	}
	for _, v := range d.Get("tool_operation").(*schema.Set).List() {
		name, _ := v.(map[string]any)["name"].(string)
		if name != "" && !slices.Contains(tools, name) {
			return fmt.Errorf("tool_operation: %q has no tool %q (see the hush_mcp_catalog_entry "+
				"data source for its tools)", appCatalogID, name)
		}
	}
	return nil
}

// validateRateLimitToolNames is validateToolNames for a 'tool' rate_limit.
func validateRateLimitToolNames(d *schema.ResourceDiff, appCatalogID string, tools []string) error {
	if !d.NewValueKnown("rate_limit") {
		return nil
	}
	for _, v := range d.Get("rate_limit").(*schema.Set).List() {
		block := v.(map[string]any)
		name, _ := block["target"].(string)
		if block["target_type"] == "tool" && name != "" && !slices.Contains(tools, name) {
			return fmt.Errorf("rate_limit: %q has no tool %q (see the hush_mcp_catalog_entry "+
				"data source for its tools)", appCatalogID, name)
		}
	}
	return nil
}

// validateAuth checks the auth block on its own; validateAuthType checks it
// against the entry.
func validateAuth(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("auth") || len(d.Get("auth").([]any)) == 0 {
		return nil
	}
	set := func(name string) bool { return writeonly.IsSetNested(d, "auth", 0, name) }
	switch {
	case set("secret") && set("secret_wo"):
		return fmt.Errorf("auth: secret and secret_wo are mutually exclusive")
	case !set("secret") && !set("secret_wo"):
		return fmt.Errorf("auth: one of secret or secret_wo is required")
	case set("secret_wo") && !set("secret_wo_version"):
		return fmt.Errorf("auth: secret_wo requires secret_wo_version")
	}
	if !d.NewValueKnown("auth.0.type") {
		return nil
	}
	authType := d.Get("auth.0.type").(string)
	for _, rule := range [][2]string{
		{"username", client.MCPAuthTypeBasic},
		{"name", client.MCPAuthTypeHeader},
		{"client_id", client.MCPAuthTypeClientCredentials},
	} {
		field, wantType := rule[0], rule[1]
		if authType == wantType && !set(field) {
			return fmt.Errorf("auth: %s is required with type %q", field, authType)
		}
		if authType != wantType && set(field) {
			return fmt.Errorf("auth: %s is only valid with type %q", field, wantType)
		}
	}
	// OAuth sends its token in Authorization too, so only one of them can.
	if authType != client.MCPAuthTypeHeader && writeonly.IsSet(d, "client_id") {
		return fmt.Errorf("auth: type %q sets the Authorization header, which OAuth needs "+
			"for client_id; use type %q instead", authType, client.MCPAuthTypeHeader)
	}
	return nil
}

func validateAuthType(d *schema.ResourceDiff, appCatalogID string, shape entryShape) error {
	if !d.NewValueKnown("auth") {
		return nil
	}
	if len(d.Get("auth").([]any)) == 0 {
		if shape.authRequired {
			return fmt.Errorf("auth: %q takes no OAuth, so an auth block is required (type %s)",
				appCatalogID, strings.Join(shape.authTypes, " or "))
		}
		return nil
	}
	if !d.NewValueKnown("auth.0.type") {
		return nil
	}
	authType := d.Get("auth.0.type").(string)
	if !slices.Contains(shape.authTypes, authType) {
		if len(shape.authTypes) == 0 {
			return fmt.Errorf("auth: %q takes no auth block", appCatalogID)
		}
		return fmt.Errorf("auth: %q takes type %s, not %q",
			appCatalogID, strings.Join(shape.authTypes, " or "), authType)
	}
	return nil
}

// heimdall replaces an application's auth but can't remove it.
func forceNewOnAuthRemoval(d *schema.ResourceDiff) error {
	old, new := d.GetChange("auth")
	if d.Id() != "" && len(old.([]any)) > 0 && len(new.([]any)) == 0 {
		return d.ForceNew("auth")
	}
	return nil
}
