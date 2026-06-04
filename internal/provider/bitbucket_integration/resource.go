package bitbucket_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
	"github.com/hushsecurity/terraform-provider-hush/internal/writeonly"
)

const resourceDescription = "Manages a Hush Security Bitbucket integration for scanning Bitbucket repositories for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: bitbucketIntegrationCreate,
		ReadContext:   bitbucketIntegrationRead,
		UpdateContext: bitbucketIntegrationUpdate,
		DeleteContext: bitbucketIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: BitbucketIntegrationResourceSchema(),
	}
}

func bitbucketIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	token := writeonly.GetString(d, "token", "token_wo")

	input := &client.CreateBitbucketIntegrationInput{
		Name:          d.Get("name").(string),
		WorkspaceSlug: d.Get("workspace_slug").(string),
		Token:         token,
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}

	resp, err := client.CreateBitbucketIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return bitbucketIntegrationRead(ctx, d, m)
}

func bitbucketIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	// Handle token rotation separately
	if d.HasChanges("token", "token_wo", "token_wo_version") {
		token := writeonly.GetString(d, "token", "token_wo")

		if token != "" {
			replaceInput := &client.ReplaceBitbucketTokenInput{
				Token: token,
			}
			if err := client.ReplaceBitbucketToken(ctx, c, d.Id(), replaceInput); err != nil {
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
	input := &client.UpdateBitbucketIntegrationInput{}
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

	if hasChanges {
		_, err := client.UpdateBitbucketIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return bitbucketIntegrationRead(ctx, d, m)
}

func bitbucketIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteBitbucketIntegration(ctx, c, d.Id())
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
