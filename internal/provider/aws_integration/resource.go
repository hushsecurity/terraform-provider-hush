package aws_integration

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

const resourceDescription = "Manages a Hush Security AWS integration for scanning AWS accounts for secrets and sensitive data."

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: resourceDescription,

		CreateContext: awsIntegrationCreate,
		ReadContext:   awsIntegrationRead,
		UpdateContext: awsIntegrationUpdate,
		DeleteContext: awsIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: AWSIntegrationResourceSchema(),
	}
}

func awsIntegrationCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.CreateAWSIntegrationInput{
		Name: d.Get("name").(string),
	}

	if desc := d.Get("description").(string); desc != "" {
		input.Description = desc
	}
	if roleArn := d.Get("role_arn").(string); roleArn != "" {
		input.RoleArn = roleArn
	}
	if cfArn := d.Get("cf_stackset_arn").(string); cfArn != "" {
		input.CfStacksetArn = cfArn
	}
	if suffix := d.Get("unique_suffix").(string); suffix != "" {
		input.UniqueSuffix = suffix
	}

	resp, err := client.CreateAWSIntegration(ctx, c, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return awsIntegrationRead(ctx, d, m)
}

func awsIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	input := &client.UpdateAWSIntegrationInput{}
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
		_, err := client.UpdateAWSIntegration(ctx, c, d.Id(), input)
		if err != nil {
			errResponse, ok := err.(*client.APIError)
			if ok && errResponse.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return awsIntegrationRead(ctx, d, m)
}

func awsIntegrationDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(*client.Client)

	err := client.DeleteAWSIntegration(ctx, c, d.Id())
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
