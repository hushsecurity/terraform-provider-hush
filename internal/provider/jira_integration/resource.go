package jira_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

const resourceDescription = "Manages a Hush Security Jira integration for scanning Jira issues for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: jiraIntegrationCreate,
		ReadContext:   jiraIntegrationRead,
		UpdateContext: jiraIntegrationUpdate,
		DeleteContext: jiraIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: JiraIntegrationResourceSchema(),
	}
}

func jiraIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	apiKey := writeonly.GetString(d, "api_key", "api_key_wo")

	input := &client.CreateJiraIntegrationInput{
		Name:      d.Get("name").(string),
		OrgDomain: d.Get("org_domain").(string),
		User:      d.Get("user").(string),
		ApiKey:    apiKey,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	syncIssues := d.Get("sync_issues_resolution").(bool)
	input.SyncIssuesResolution = &syncIssues

	enableScans := d.Get("enable_scans").(bool)
	input.EnableScans = &enableScans

	resp, err := client.CreateJiraIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return jiraIntegrationRead(ctx, d, m)
}

func jiraIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle api_key rotation separately
	if d.HasChanges("api_key", "api_key_wo", "api_key_wo_version", "user") {
		apiKey := writeonly.GetString(d, "api_key", "api_key_wo")

		if apiKey != "" {
			replaceInput := &client.ReplaceJiraApiKeyInput{
				User:   d.Get("user").(string),
				ApiKey: apiKey,
			}
			if err := client.ReplaceJiraApiKey(ctx, c, d.Id(), replaceInput); err != nil {
				errResponse, ok := err.(*client.APIError)
				if ok && errResponse.StatusCode == http.StatusNotFound {
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}
		}
	}

	// Handle metadata updates
	input := &client.UpdateJiraIntegrationInput{}
	hasChanges := false

	if d.HasChange("name") {
		name := d.Get("name").(string)
		input.Name = &name
		hasChanges = true
	}
	if d.HasChange("description") {
		desc := d.Get("description").(string)
		input.Description = &desc
		hasChanges = true
	}
	if d.HasChange("org_domain") {
		orgDomain := d.Get("org_domain").(string)
		input.OrgDomain = &orgDomain
		hasChanges = true
	}
	if d.HasChange("sync_issues_resolution") {
		syncIssues := d.Get("sync_issues_resolution").(bool)
		input.SyncIssuesResolution = &syncIssues
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateJiraIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return jiraIntegrationRead(ctx, d, m)
}

func jiraIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteJiraIntegration(ctx, c, d.Id())
	if err != nil {
		errResponse, ok := err.(*client.APIError)
		if ok && errResponse.StatusCode == http.StatusNotFound {
			d.SetId("")
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
