package mcp_application

import (
	"context"
	"errors"
	"fmt"
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
	if !d.NewValueKnown("app_catalog_id") {
		return nil
	}
	appCatalogID := d.Get("app_catalog_id").(string)
	if err := validateEntrySettings(d, appCatalogID); err != nil {
		return err
	}
	// url_label and the credentials are only judged against the entry at
	// creation: url_label cannot change afterwards, and a credential the entry
	// needs was already given or the application would not exist.
	if d.Id() != "" || m == nil {
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
			return nil
		}
		for _, option := range app.URLs {
			shape.labels = append(shape.labels, labelOf(option))
		}
	} else {
		entry, err := client.GetMCPCatalogEntry(ctx, c, appCatalogID)
		if err != nil {
			var apiErr *client.APIError
			if errors.As(err, &apiErr) && apiErr.IsNotFound() {
				return fmt.Errorf("app_catalog_id: the catalog has no MCP entry %q", appCatalogID)
			}
			return nil
		}
		// A hosted entry has exactly one address, Hush's own, which the API
		// leaves out; whatever it does list is not what url_label picks from.
		if entry.Hosted {
			shape.labels = []string{""}
		} else {
			for _, option := range entry.URLs {
				shape.labels = append(shape.labels, labelOf(option))
			}
		}
		shape.manualRegistration = entry.ManualRegistration
	}

	if err := validateURLLabel(d, appCatalogID, shape.labels); err != nil {
		return err
	}
	// heimdall refuses the create unless it gets both halves.
	hasSecret := writeonly.IsSet(d, "client_secret") || writeonly.IsSet(d, "client_secret_wo")
	if shape.manualRegistration && (!writeonly.IsSet(d, "client_id") || !hasSecret) {
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
