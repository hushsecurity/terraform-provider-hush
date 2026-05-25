package gitlab_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Manages a Hush Security GitLab integration for scanning GitLab repositories for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: gitlabIntegrationCreate,
		ReadContext:   gitlabIntegrationRead,
		UpdateContext: gitlabIntegrationUpdate,
		DeleteContext: gitlabIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: GitlabIntegrationResourceSchema(),
	}
}

func gitlabIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	token := d.Get("token").(string)
	if token == "" {
		token = d.Get("token_wo").(string)
	}

	input := &client.CreateGitlabIntegrationInput{
		Name:  d.Get("name").(string),
		Token: token,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}
	if depID := d.Get("onprem_deployment_id").(string); depID != "" {
		input.OnpremDeploymentID = depID
	}
	if v, ok := d.GetOk("group_id"); ok {
		groupID := v.(int)
		input.GroupID = &groupID
	}
	if v, ok := d.GetOk("project_id"); ok {
		projectID := v.(int)
		input.ProjectID = &projectID
	}
	if v, ok := d.GetOk("visibilities"); ok {
		input.Visibilities = expandStringList(v.([]any))
	}
	if v := d.Get("base_url").(string); v != "" {
		input.BaseURL = v
	}
	if v, ok := d.GetOk("selected_repos"); ok {
		input.SelectedRepos = expandStringList(v.([]any))
	}
	if v, ok := d.GetOk("enable_pr_scans"); ok {
		enablePR := v.(bool)
		input.EnablePRScans = &enablePR
	}

	resp, err := client.CreateGitlabIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	if depID := d.Get("onprem_deployment_id").(string); depID != "" {
		if err := client.WaitForAccessBridge(ctx, c, depID); err != nil {
			return diag.Errorf("error waiting for on-prem deployment %s to become available: %s", depID, err)
		}
	}

	return gitlabIntegrationRead(ctx, d, m)
}

func gitlabIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle token rotation separately
	if d.HasChanges("token", "token_wo", "token_wo_version") {
		token := d.Get("token").(string)
		if token == "" {
			token = d.Get("token_wo").(string)
		}

		if token != "" {
			replaceInput := &client.ReplaceGitlabTokenInput{
				Token: token,
			}
			if err := client.ReplaceGitlabToken(ctx, c, d.Id(), replaceInput); err != nil {
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
	input := &client.UpdateGitlabIntegrationInput{}
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
	if d.HasChange("onprem_deployment_id") {
		depID := d.Get("onprem_deployment_id").(string)
		input.OnpremDeploymentID = &depID
		hasChanges = true
	}
	if d.HasChange("visibilities") {
		if v, ok := d.GetOk("visibilities"); ok {
			input.Visibilities = expandStringList(v.([]any))
		} else {
			input.Visibilities = []string{}
		}
		hasChanges = true
	}
	if d.HasChange("base_url") {
		baseURL := d.Get("base_url").(string)
		input.BaseURL = &baseURL
		hasChanges = true
	}
	if d.HasChange("selected_repos") {
		if v, ok := d.GetOk("selected_repos"); ok {
			input.SelectedRepos = expandStringList(v.([]any))
		} else {
			input.SelectedRepos = []string{}
		}
		hasChanges = true
	}
	if d.HasChange("enable_pr_scans") {
		enablePR := d.Get("enable_pr_scans").(bool)
		input.EnablePRScans = &enablePR
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateGitlabIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	if d.HasChange("onprem_deployment_id") {
		if depID := d.Get("onprem_deployment_id").(string); depID != "" {
			if err := client.WaitForAccessBridge(ctx, c, depID); err != nil {
				return diag.Errorf("error waiting for on-prem deployment %s to become available: %s", depID, err)
			}
		}
	}

	return gitlabIntegrationRead(ctx, d, m)
}

func gitlabIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteGitlabIntegration(ctx, c, d.Id())
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

func expandStringList(input []any) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}
