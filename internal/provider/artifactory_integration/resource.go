package artifactory_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

const resourceDescription = "Manages a Hush Security Artifactory integration for scanning Artifactory repositories for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: artifactoryIntegrationCreate,
		ReadContext:   artifactoryIntegrationRead,
		UpdateContext: artifactoryIntegrationUpdate,
		DeleteContext: artifactoryIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: ArtifactoryIntegrationResourceSchema(),
	}
}

func artifactoryIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	token := writeonly.GetString(d, "token", "token_wo")

	input := &client.CreateArtifactoryIntegrationInput{
		Name:  d.Get("name").(string),
		Token: token,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	if orgURL := d.Get("org_url").(string); orgURL != "" {
		input.OrgURL = orgURL
	}

	if onpremID := d.Get("onprem_deployment_id").(string); onpremID != "" {
		input.OnpremDeploymentID = onpremID
	}

	resp, err := client.CreateArtifactoryIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return artifactoryIntegrationRead(ctx, d, m)
}

func artifactoryIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle token rotation separately
	if d.HasChanges("token", "token_wo", "token_wo_version") {
		token := writeonly.GetString(d, "token", "token_wo")

		if token != "" {
			replaceInput := &client.ReplaceArtifactoryTokenInput{
				Token: token,
			}
			if err := client.ReplaceArtifactoryToken(ctx, c, d.Id(), replaceInput); err != nil {
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
	input := &client.UpdateArtifactoryIntegrationInput{}
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
	if d.HasChange("org_url") {
		orgURL := d.Get("org_url").(string)
		input.OrgURL = &orgURL
		hasChanges = true
	}
	if d.HasChange("onprem_deployment_id") {
		onpremID := d.Get("onprem_deployment_id").(string)
		input.OnpremDeploymentID = &onpremID
		hasChanges = true
	}

	if hasChanges {
		_, err := client.UpdateArtifactoryIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return artifactoryIntegrationRead(ctx, d, m)
}

func artifactoryIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteArtifactoryIntegration(ctx, c, d.Id())
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
