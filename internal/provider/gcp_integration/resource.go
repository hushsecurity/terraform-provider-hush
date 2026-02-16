package gcp_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "GCP integration resource for managing Hush Security GCP onboarding"

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: gcpIntegrationCreate,
		ReadContext:   gcpIntegrationResourceRead,
		UpdateContext: gcpIntegrationUpdate,
		DeleteContext: gcpIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: GCPIntegrationResourceSchema(),
	}
}

func gcpIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateGCPIntegrationInput{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	// Extract projects
	if projects := extractProjectInputs(d); projects != nil {
		input.Projects = projects
	}

	// Extract features
	if features := extractFeatureInputs(d); features != nil {
		input.Features = features
	}

	// Step 1: Create integration (status=pending)
	resp, err := client.CreateGCPIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	// Step 2: If service_account_email is provided, complete the integration
	if saEmail, ok := d.GetOk("service_account_email"); ok {
		email := saEmail.(string)
		if email != "" {
			completeInput := &client.CompleteGCPIntegrationInput{
				ServiceAccountEmail: email,
			}
			_, err := client.CompleteGCPIntegration(ctx, c, resp.ID, completeInput)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return gcpIntegrationResourceRead(ctx, d, m)
}

func gcpIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateGCPIntegrationInput{}
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
	if d.HasChange("project") {
		projects := extractProjectInputs(d)
		input.Projects = &projects
		hasChanges = true
	}
	if d.HasChange("feature") {
		features := extractFeatureInputs(d)
		input.Features = &features
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateGCPIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	// Handle service_account_email change: if changed and integration is pending, complete it
	if d.HasChange("service_account_email") {
		if saEmail, ok := d.GetOk("service_account_email"); ok {
			email := saEmail.(string)
			if email != "" {
				// Check current status - only complete if pending
				integ, err := client.GetGCPIntegration(ctx, c, d.Id())
				if err != nil {
					return diag.FromErr(err)
				}
				if integ.Status == "pending" {
					completeInput := &client.CompleteGCPIntegrationInput{
						ServiceAccountEmail: email,
					}
					_, err := client.CompleteGCPIntegration(ctx, c, d.Id(), completeInput)
					if err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	return gcpIntegrationResourceRead(ctx, d, m)
}

func gcpIntegrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteGCPIntegration(ctx, c, d.Id())
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
